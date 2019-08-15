package comet

import (
	"context"
	"encoding/json"
	"gitlab.com/jetfueltw/cpw/alakazam/app/comet/pb"
	"gitlab.com/jetfueltw/cpw/alakazam/models"
	"gitlab.com/jetfueltw/cpw/alakazam/pkg/bytes"
	xtime "gitlab.com/jetfueltw/cpw/alakazam/pkg/time"
	"gitlab.com/jetfueltw/cpw/alakazam/pkg/websocket"
	"gitlab.com/jetfueltw/cpw/micro/log"
	"go.uber.org/zap"
	"io"
	"net"
	"strings"
	"time"
)

const (
	maxInt = 1<<31 - 1
)

var blockadeMessage []byte

func init() {
	errBlockade := struct {
		Message string `json:"message"`
	}{
		Message: "您在封鎖状态，无法进入聊天室",
	}
	b, err := json.Marshal(errBlockade)
	if err != nil {
		panic(err)
	}
	blockadeMessage = b
}

// 開始監聽Websocket
func InitWebsocket(server *Server, host string, accept int) (err error) {
	var (
		listener *net.TCPListener
		addr     *net.TCPAddr
	)

	// 監聽Tcp Port
	if addr, err = net.ResolveTCPAddr("tcp", host); err != nil {
		return
	}
	if listener, err = net.ListenTCP("tcp", addr); err != nil {
		return
	}

	// 一個Tcp Port根據CPU核心數開goroutine監聽Tcp
	for i := 0; i < accept; i++ {
		go acceptWebsocket(server, listener)
	}
	return
}

// 處理Websocket連線
func acceptWebsocket(server *Server, lis *net.TCPListener) {
	var (
		conn *net.TCPConn
		err  error
		r    int
	)
	for {
		// tcp監聽並連線
		if conn, err = lis.AcceptTCP(); err != nil {
			log.Error("listener accept", zap.Error(err), zap.String("addr", lis.Addr().String()))
			return
		}
		// tcp 開啟KeepAlive
		if err = conn.SetKeepAlive(server.c.TCP.KeepAlive); err != nil {
			log.Error("conn setKeepAlive", zap.Error(err))
			return
		}
		// tcp讀取資料的緩衝區大小，該緩衝區為0時會阻塞，此值通常設定完後，系統會自行在多一倍，設定1024會變2304
		if err = conn.SetReadBuffer(server.c.TCP.Rcvbuf); err != nil {
			log.Error("conn setReadBuffer", zap.Error(err))
			return
		}
		// tcp寫資料的緩衝區大小，該緩衝區滿到無法發送時會阻塞，此值通常設定完後系統會自行在多一倍，設定1024會變2304
		if err = conn.SetWriteBuffer(server.c.TCP.Sndbuf); err != nil {
			log.Error("conn setWriteBuffer", zap.Error(err))
			return
		}
		go serveWebsocket(server, conn, r)
		if r++; r == maxInt {
			r = 0
		}
	}
}

// Websocket連線後的邏輯處理
func serveWebsocket(s *Server, conn net.Conn, r int) {
	var (
		// 任務倒數計時器
		tr = s.round.Timer(r)

		// Reader Buffer
		rp = s.round.Reader(r)

		// Writer Buffer
		wp = s.round.Writer(r)

		err error

		// 房間id
		rid string

		// 心跳時間週期
		hb time.Duration

		// grpc 自訂Protocol
		p *pb.Proto

		// 管理Channel與Room
		b *Bucket

		// 時間倒數任務
		trd *xtime.TimerData

		// 現在時間
		lastHB = time.Now()

		// 用於讀的Buffer
		rb = rp.Get()

		// 此tcp連線的Channel
		ch = NewChannel(s.c.Protocol.ProtoSize, s.c.Protocol.RevBuffer)

		// Reader byte
		rr = &ch.Reader

		// Writer byte
		wr = &ch.Writer

		ws *websocket.Conn

		req *websocket.Request
	)

	// Channel設置的讀Buffer(由Pool取得之後會還給Pool做復用)
	ch.Reader.ResetBuffer(conn, rb.Bytes())

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	step := 0

	// 心跳超時後的邏輯
	trd = tr.Add(s.c.Protocol.HandshakeTimeout, func() {
		_ = conn.SetDeadline(time.Now().Add(time.Millisecond * 100))
		_ = conn.Close()
		log.Error("ws handshake timeout", zap.String("uid", ch.Uid), zap.Int("step", step))
	})

	ch.IP, _, _ = net.SplitHostPort(conn.RemoteAddr().String())
	step = 1

	// 判斷連線的 url path info正不正確
	if req, err = websocket.ReadRequest(rr); err != nil || req.RequestURI != "/sub" {
		// 關掉連線
		// 移除心跳timeout任務
		// 回收讀Buffer
		conn.Close()
		tr.Del(trd)
		rp.Put(rb)
		if err != io.EOF {
			log.Error("websocket readRequest", zap.Error(err))
		}
		return
	}

	// 用於寫的Buffer
	// Channel設置的寫Buffer(由Pool取得之後會還給Pool做復用)
	wb := wp.Get()
	ch.Writer.ResetBuffer(conn, wb.Bytes())
	step = 2

	// 將Tcp連線包成websocket
	if ws, err = websocket.Upgrade(conn, rr, wr, req); err != nil {
		// 關掉連線
		// 移除心跳timeout任務
		// 回收讀寫Buffer
		conn.Close()
		tr.Del(trd)
		rp.Put(rb)
		wp.Put(wb)
		if err != io.EOF {
			log.Error("websocket new server conn", zap.Error(err))
		}
		return
	}

	step = 3

	// websocket連線等待read做auth
	if p, err = ch.protoRing.Set(); err == nil {
		if rid, hb, err = s.authWebsocket(ctx, ws, ch, p); err == nil {
			// 將user Channel放到某一個Bucket內做保存
			b = s.Bucket(ch.Key)
			err = b.Put(rid, ch)
		}
	}

	step = 4

	// 如果操作有異常則
	// 1. 回收讀寫Buffer
	// 2. 解除心跳任務(close 連線)
	// 3. 關閉連線
	if err != nil {
		ws.Close()
		rp.Put(rb)
		wp.Put(wb)
		tr.Del(trd)
		if err != io.EOF && err != websocket.ErrMessageClose {
			log.Error("ws handshake failed", zap.Error(err), zap.String("uid", ch.Uid), zap.Int("step", step))
		}
		return
	}

	// 進入某房間成功後重置心跳任務
	trd.Key = ch.Key
	tr.Set(trd, hb)

	step = 5

	// 處理訊息推送
	go s.dispatchWebsocket(ws, wp, wb, ch)

	serverHeartbeat := s.RandServerHearbeat()

	for {
		if p, err = ch.protoRing.Set(); err != nil {
			break
		}
		if err = p.ReadWebsocket(ws); err != nil {
			break
		}
		if p.Op == pb.OpHeartbeat {
			// comet有心跳機制維護連線狀態，對於logic來說也需要有人利用心跳機制去告知哪個user還在線
			// 目前在不在線這個狀態都是由comet控管，但不需要每次webSocket -> 心跳 -> comet就 -> 心跳 -> logic
			// 所以webSocket -> comet 心跳週期會比 comet -> logic還要短
			// 假設
			// 1. webSocket -> comet 5分鐘沒心跳就過期
			// 2. comet -> logic 20分鐘沒心跳就過期
			// webSocket -> 每30秒心跳 -> comet <====== 每次只要不超過5分鐘沒心跳則comet會認為連線沒問題
			// webSocket -> 每30秒心跳 -> comet -> 判斷是否已經快20分鐘沒通知logic(是就發) -> logic
			tr.Set(trd, hb)
			p.Op = pb.OpHeartbeatReply
			p.Body = nil
			if now := time.Now(); now.Sub(lastHB) > serverHeartbeat {
				if err = s.Heartbeat(ctx, ch); err != nil {
					break
				}
				lastHB = now
			}
			step++
		} else {
			// 非心跳動作
			if err = s.Operate(ctx, p, ch, b); err != nil {
				break
			}
		}
		// 寫的游標要++讓Get可以取得已寫入的Proto
		ch.protoRing.SetAdv()
		// 通知負責訊息推播goroutine處理本次接收到的資料
		ch.Signal()
	}

	// 如果某人連線有異常或是server要踢人則
	// 1. 從Bucket移除user Channel，這樣對Bucket內的Channel才都是活人
	// 2. 解除心跳任務(close 連線)
	// 3. 回收讀Buffer，不回收寫的Buffer是因為Channel close後dispatchTCP會被通知到並回收寫的Buffer
	// 4. 關閉連線
	// 5. 通知logic某人下線了
	if err != nil && err != io.EOF && err != websocket.ErrMessageClose && !strings.Contains(err.Error(), "closed") {
		log.Error("server ws failed", zap.Error(err), zap.String("uid", ch.Uid), zap.Int("step", step))
	}
	b.Del(ch)
	tr.Del(trd)
	ws.Close()
	ch.Close()
	rp.Put(rb)
	if err = s.Disconnect(ctx, ch.Uid, ch.Key); err != nil {
		log.Error("grpc client disconnect", zap.Error(err), zap.String("uid", ch.Uid))
	}
}

// 處理Websocket訊息推送
func (s *Server) dispatchWebsocket(ws *websocket.Conn, wp *bytes.Pool, wb *bytes.Buffer, ch *Channel) {
	var (
		err    error
		finish bool
		online int32
	)

	for {
		// 接收到別人通知説有資料要推送，沒資料時就阻塞
		var p = ch.Ready()
		switch p {

		// websocket連線要關閉
		case pb.ProtoFinish:
			finish = true
			goto failed

			// 有資料需要推送
		case pb.ProtoReady:
			for {
				// 取得上次透過Set()寫入資料的Proto
				if p, err = ch.protoRing.Get(); err != nil {
					break
				}
				if p.Op == pb.OpHeartbeatReply {
					if ch.Room != nil {
						online = ch.Room.OnlineNum()
					}
					if err = p.WriteWebsocketHeart(ws, online); err != nil {
						goto failed
					}
				} else {
					if err = p.WriteWebsocket(ws); err != nil {
						goto failed
					}
				}
				p.Body = nil
				// 讀的游標++
				ch.protoRing.GetAdv()
			}
		default:
			if err = p.WriteWebsocket(ws); err != nil {
				goto failed
			}
		}
		// 送出資料給client
		if err = ws.Flush(); err != nil {
			break
		}
	}
	// 連線有異常或是server要踢人
	// 1. 連線close
	// 2. 回收寫的Buffter
failed:
	if err != nil && err != io.EOF && err != websocket.ErrMessageClose {
		log.Error("dispatch websocket", zap.Error(err), zap.String("uid", ch.Uid))
	}
	ws.Close()
	wp.Put(wb)
	// must ensure all channel message discard, for reader won't blocking Signal
	for !finish {
		finish = (ch.Ready() == pb.ProtoFinish)
	}
}

// websocket請求連線至某房間
func (s *Server) authWebsocket(ctx context.Context, ws *websocket.Conn, ch *Channel, p *pb.Proto) (string, time.Duration, error) {
	for {
		// 如果第一次連線送的資料不是請求連接到某房間則會一直等待
		if err := p.ReadWebsocket(ws); err != nil {
			return "", time.Duration(0), err
		}
		if p.Op == pb.OpAuth {
			break
		} else {
			log.Error("ws request not auth", zap.Int32("op", p.Op))
		}
	}

	// 有兩種情況會無法進入聊天室
	// 1. 請求進入聊天室資料有誤
	// 2. 被封鎖
	c, err := s.Connect(ctx, p)
	if err != nil {
		return "", time.Duration(0), err
	}
	if c.Status == models.Blockade {
		if e := authReply(ws, p, blockadeMessage); e != nil {
			log.Warn("auth reply", zap.Error(e), zap.String("uid", c.Uid), zap.String("room_id", c.RoomID))
		}
		return "", time.Duration(0), io.EOF
	}

	// 需要回覆給client告知uid與key
	// 因為後續發話需依靠這兩個欄位來做pk
	reply := struct {
		Uid        string `json:"uid"`
		Key        string `json:"key"`
		RoomId     string `json:"room_id"`
		Permission struct {
			IsBanned      bool `json:"is_banned"`
			IsRedEnvelope bool `json:"is_red_envelope"`
		} `json:"permission"`
	}{
		Uid:    c.Uid,
		Key:    c.Key,
		RoomId: c.RoomID,
	}
	reply.Permission.IsBanned = models.IsBanned(int(c.Status))
	reply.Permission.IsRedEnvelope = models.IsRedEnvelope(int(c.Status))

	b, err := json.Marshal(reply)
	if err != nil {
		return "", time.Duration(0), err
	}
	if err = authReply(ws, p, b); err != nil {
		return "", time.Duration(0), err
	}
	ch.Uid = c.Uid
	ch.Key = c.Key
	ch.Name = c.Name
	return c.RoomID, time.Duration(c.Heartbeat), nil
}

// 回覆連線至某房間結果
func authReply(ws *websocket.Conn, p *pb.Proto, b []byte) (err error) {
	p.Op = pb.OpAuthReply
	p.Body = b
	if err = p.WriteWebsocket(ws); err != nil {
		return
	}
	err = ws.Flush()
	return
}
