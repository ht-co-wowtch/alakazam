package comet

import (
	"context"
	"gitlab.com/jetfueltw/cpw/alakazam/protocol"
	"time"

	log "github.com/golang/glog"
	pd "gitlab.com/jetfueltw/cpw/alakazam/protocol/grpc"
	"google.golang.org/grpc"
	"google.golang.org/grpc/encoding/gzip"
)

// 告知logic service有人想要進入某個房間
func (s *Server) Connect(c context.Context, p *pd.Proto, cookie string) (key, rid string, heartbeat time.Duration, err error) {
	reply, err := s.rpcClient.Connect(c, &pd.ConnectReq{
		Server: s.serverID,
		Cookie: cookie,
		Token:  p.Body,
	})
	if err != nil {
		return
	}
	return reply.Id, reply.RoomID, time.Duration(reply.Heartbeat), nil
}

//  client連線中斷，告知logic service需清理此人的連線資料
func (s *Server) Disconnect(c context.Context, id string) (err error) {
	_, err = s.rpcClient.Disconnect(context.Background(), &pd.DisconnectReq{
		Server: s.serverID,
		Id:     id,
	})
	return
}

// 告知logic service要刷新某人的在線狀態(session 過期時間)
func (s *Server) Heartbeat(ctx context.Context, id string) (err error) {
	_, err = s.rpcClient.Heartbeat(ctx, &pd.HeartbeatReq{
		Server: s.serverID,
		Id:     id,
	})
	return
}

// 告知logic service現在房間的在線人數
func (s *Server) RenewOnline(ctx context.Context, serverID string, rommCount map[string]int32) (allRoom map[string]int32, err error) {
	reply, err := s.rpcClient.RenewOnline(ctx, &pd.OnlineReq{
		Server:    s.serverID,
		RoomCount: rommCount,
	}, grpc.UseCompressor(gzip.Name))
	if err != nil {
		return
	}
	return reply.AllRoomCount, nil
}

// 處理Proto相關邏輯
func (s *Server) Operate(ctx context.Context, p *pd.Proto, ch *Channel, b *Bucket) error {
	switch p.Op {
	// 更換房間
	case protocol.OpChangeRoom:
		if err := b.ChangeRoom(string(p.Body), ch); err != nil {
			log.Errorf("b.ChangeRoom(%s) error(%v)", p.Body, err)
		}
		p.Op = protocol.OpChangeRoomReply
	default:
		// TODO ack ok&failed
		p.Body = nil
	}
	return nil
}
