package test

import (
	"encoding/json"
	"fmt"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"gitlab.com/jetfueltw/cpw/alakazam/pkg/bufio"
	"gitlab.com/jetfueltw/cpw/alakazam/pkg/encoding/binary"
	"gitlab.com/jetfueltw/cpw/alakazam/protocol"
	"gitlab.com/jetfueltw/cpw/alakazam/protocol/grpc"
	"gitlab.com/jetfueltw/cpw/alakazam/server/logic"
	"golang.org/x/net/websocket"
	"io"
	"io/ioutil"
	"math/rand"
	"net/http"
	"net/url"
	"os"
	"strings"
	"testing"
	"time"
)

const (
	host = "http://127.0.0.1:3111"
)

type AuthToken struct {
	RoomID string `json:"room_id"`
	Token  string `json:"token"`
}

type auth struct {
	uid   string
	key   string
	wr    *bufio.Writer
	rd    *bufio.Reader
	proto *grpc.Proto
}

type resp struct {
	a          auth
	p          []grpc.Proto
	body       []byte
	statusCode int
	err        error

	otherProto []grpc.Proto
	otherErr   error
}

var (
	authToken  AuthToken
	httpClient *http.Client
)

func TestMain(m *testing.M) {
	rand.Seed(time.Now().Unix())
	authToken = AuthToken{
		"1000",
		uuid.New().String(),
	}

	httpClient = &http.Client{
		Timeout: time.Second * 5,
	}
	os.Exit(m.Run())
}

// 進入房間成功
func Test_auth(t *testing.T) {
	a, err := dialAuth(authToken)
	if err != nil {
		assert.Error(t, err)
		return
	}
	shouldBeAuthReply(t, a)
}

// 進入房間失敗
func Test_not_auth(t *testing.T) {
	ws, err := dial()
	if err != nil {
		assert.Error(t, err)
		return
	}
	shouldBeCloseConnection(err, ws, t)
}

// 房間心跳成功
func Test_heartbeat(t *testing.T) {
	a, err := dialAuth(authToken)
	if err != nil {
		assert.Error(t, err)
		return
	}
	shouldBeHeartbeatReply(t, a, givenHeartbeat())
}

// 房間不心跳
func Test_not_heartbeat(t *testing.T) {
	a, err := dialAuth(authToken)
	if err != nil {
		assert.Error(t, err)
		return
	}
	shouldBeTimeoutConnection(err, a, t)
}

// 房間訊息推送成功
func Test_push_room(t *testing.T) {
	a, err := dialAuth(authToken)
	if err != nil {
		assert.Fail(t, err.Error())
		return
	}
	r := pushRoom(a.uid, a.key, "測試")

	assert.Equal(t, http.StatusNoContent, r.statusCode)
	assert.Empty(t, r.body)
	fmt.Println("ok")
}

// 讀取房間訊息
func Test_read_room_message(t *testing.T) {
	pushTest(t, authToken, func(a auth) resp {
		return pushRoom(a.uid, a.key, "測試")
	}, func(r resp) {
		assert.Equal(t, protocol.OpBatchRaw, r.a.proto.Op)
		assert.Len(t, r.p, 1)
		assert.Nil(t, r.otherErr)
		assert.Len(t, r.otherProto, 1)
	})
}

// 讀取房間訊息格式
func Test_read_room_message_payload(t *testing.T) {
	pushTest(t, authToken, func(a auth) resp {
		return pushRoom(a.uid, a.key, "測試")
	}, func(r resp) {
		l := new(logic.Message)
		json.Unmarshal(r.p[0].Body, l)
		tz, _ := time.Parse("15:04:05", l.Time)
		assert.Equal(t, "test", l.Name)
		assert.Equal(t, "", l.Avatar)
		assert.Equal(t, "測試", l.Message)
		assert.False(t, tz.IsZero())
	})
}

// 廣播訊息推送
func Test_push_broadcast(t *testing.T) {
	a, err := dialAuth(authToken)
	if err != nil {
		assert.Fail(t, err.Error())
		return
	}
	r := pushBroadcast(a.uid, a.key, "測試", []string{"1000", "1001"})

	assert.Equal(t, http.StatusNoContent, r.statusCode)
	assert.Empty(t, r.body)
	fmt.Println("ok")
}

// 讀取廣播房間訊息
func Test_read_broadcast_message(t *testing.T) {
	pushTest(t, authToken, func(a auth) resp {
		return pushBroadcast(a.uid, a.key, "測試", []string{"1000", "1001"})
	}, func(r resp) {
		assert.Equal(t, protocol.OpBatchRaw, r.a.proto.Op)
		assert.Len(t, r.p, 1)
		assert.Nil(t, r.otherErr)
		assert.Len(t, r.otherProto, 1)
	})
}

// 切換房間
func Test_change_room(t *testing.T) {
	a, err := dialAuth(authToken)
	if err != nil {
		assert.Fail(t, err.Error())
		return
	}

	proto := new(grpc.Proto)
	proto.Op = protocol.OpChangeRoom
	proto.Body = []byte(`1001`)

	if err = writeProto(a.wr, proto); err != nil {
		assert.Fail(t, err.Error())
		return
	}
	if err := readProto(a.rd, a.proto); err != nil {
		assert.Fail(t, err.Error())
		return
	}

	assert.Equal(t, protocol.OpChangeRoomReply, a.proto.Op)
	assert.Equal(t, "1001", string(a.proto.Body))
	fmt.Println("ok")
}

func pushTest(t *testing.T, otherAuth AuthToken, f func(a auth) (resp), ass func(r resp)) {
	a, err := dialAuth(authToken)
	if err != nil {
		assert.Fail(t, err.Error())
		return
	}

	var (
		other      auth
		otherErr   error
		otherProto []grpc.Proto
	)

	go func() {
		other, otherErr = dialAuth(otherAuth)
		otherProto, otherErr = readMessageProto(other.rd, other.proto)
	}()

	r := f(a)
	if err != nil {
		assert.Fail(t, err.Error())
		return
	}
	time.Sleep(time.Second * 3)
	var p []grpc.Proto
	if p, err = readMessageProto(a.rd, a.proto); err != nil {
		assert.Fail(t, err.Error())
		return
	}

	r.p = p
	r.a = a
	r.otherErr = otherErr
	r.otherProto = otherProto

	ass(r)
	fmt.Println("ok")
}

func shouldBeTimeoutConnection(err error, a auth, t *testing.T) {
	fmt.Println(time.Now())
	err = readProto(a.rd, a.proto)
	fmt.Println(time.Now())
	assert.Equal(t, io.EOF, err)
	fmt.Println("ok")
}

func shouldBeCloseConnection(err error, ws *websocket.Conn, t *testing.T) {
	b := make([]byte, 100)
	_, err = ws.Read(b)
	assert.Equal(t, io.EOF, err)
	fmt.Println("ok")
}

func givenHeartbeat() *grpc.Proto {
	hbProto := new(grpc.Proto)
	hbProto.Op = protocol.OpHeartbeat
	hbProto.Body = nil
	return hbProto
}

func shouldBeAuthReply(t *testing.T, a auth) {
	assert.Equal(t, protocol.OpAuthReply, a.proto.Op)
	fmt.Println("ok")
}

func shouldBeHeartbeatReply(t *testing.T, a auth, hbProto *grpc.Proto) {
	fmt.Println("send heartbeat")
	if err := writeProto(a.wr, hbProto); err != nil {
		assert.Error(t, err)
		return
	}
	if err := readProto(a.rd, a.proto); err != nil {
		assert.Error(t, err)
		return
	}
	fmt.Println("heartbeat reply")
	assert.Equal(t, protocol.OpHeartbeatReply, a.proto.Op)
	fmt.Println("ok")
}

func dial() (conn *websocket.Conn, err error) {
	conn, err = websocket.Dial("ws://127.0.0.1:3102/sub", "", "http://127.0.0.1")

	return
}

func dialAuth(authToken AuthToken) (auth auth, err error) {
	authToken.Token = uuid.New().String()
	var (
		conn *websocket.Conn
	)

	conn, err = dial()

	wr := bufio.NewWriter(conn)
	rd := bufio.NewReader(conn)

	proto := new(grpc.Proto)
	proto.Op = protocol.OpAuth
	proto.Body, _ = json.Marshal(authToken)

	fmt.Printf("send auth: %s\n", proto.Body)
	if err = writeProto(wr, proto); err != nil {
		return
	}
	if err = readProto(rd, proto); err != nil {
		return
	}
	fmt.Printf("auth reply: %s\n", proto.Body)

	auth.wr = wr
	auth.rd = rd
	auth.proto = proto

	var reply struct {
		Uid string `json:"uid"`
		Key string `json:"key"`
	}
	if err = json.Unmarshal(proto.Body, &reply); err != nil {
		return
	}
	auth.uid = string(reply.Uid)
	auth.key = reply.Key
	return
}

func writeProto(wr *bufio.Writer, p *grpc.Proto) (err error) {
	var (
		buf     []byte
		packLen int32
	)

	packLen = grpc.RawHeaderSize + int32(len(p.Body))
	if buf, err = wr.Peek(grpc.RawHeaderSize); err != nil {
		return
	}
	binary.BigEndian.PutInt32(buf[grpc.PackOffset:], packLen)
	binary.BigEndian.PutInt16(buf[grpc.HeaderOffset:], int16(grpc.RawHeaderSize))
	binary.BigEndian.PutInt32(buf[grpc.OpOffset:], p.Op)
	if p.Body != nil {
		_, err = wr.Write(p.Body)
	}
	return wr.Flush()
}

func readProto(rr *bufio.Reader, p *grpc.Proto) (err error) {
	var (
		bodyLen   int
		headerLen int16
		packLen   int32
	)

	packLen, headerLen, err = read(rr, p)
	if bodyLen = int(packLen - int32(headerLen)); bodyLen > 0 {
		p.Body, err = rr.Pop(bodyLen)
	} else {
		p.Body = nil
	}
	return
}

func readMessageProto(rr *bufio.Reader, p *grpc.Proto) (protos []grpc.Proto, err error) {
	var (
		bodyLen   int
		headerLen int16
		packLen   int32
	)

	packLen, headerLen, err = read(rr, p)
	for offset := int(headerLen); offset < int(packLen); offset += int(packLen) {
		proto := new(grpc.Proto)
		packLen, headerLen, err = read(rr, proto)
		if bodyLen = int(packLen - int32(headerLen)); bodyLen > 0 {
			proto.Body, err = rr.Pop(bodyLen)
		} else {
			proto.Body = nil
		}
		protos = append(protos, *proto)
	}
	return
}

func read(rr *bufio.Reader, p *grpc.Proto) (packLen int32, headerLen int16, err error) {
	var (
		buf []byte
	)
	if buf, err = rr.Pop(grpc.RawHeaderSize); err != nil {
		return
	}

	packLen = binary.BigEndian.Int32(buf[grpc.PackOffset:grpc.HeaderOffset])
	headerLen = binary.BigEndian.Int16(buf[grpc.HeaderOffset:grpc.OpOffset])
	p.Op = binary.BigEndian.Int32(buf[grpc.OpOffset:])
	return
}

func pushRoom(uid, key, message string) resp {
	data := url.Values{}
	data.Set("uid", uid)
	data.Set("key", key)
	data.Set("message", message)
	return push(host+"/push/room", data)
}

func pushBroadcast(uid, key, message string, roomId []string, ) resp {
	data := url.Values{
		"room_id": roomId,
	}
	data.Set("uid", uid)
	data.Set("key", key)
	data.Set("message", message)
	return push(fmt.Sprintf(host+"/push/all"), data)
}

func push(url string, data url.Values) (re resp) {
	r, err := httpPost(url, data)
	if err != nil {
		return
	}

	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		return
	}

	re.err = r.Body.Close()
	re.statusCode = r.StatusCode
	re.body = body
	fmt.Printf("response %s\n", string(body))
	return
}

func httpPost(url string, body url.Values) (*http.Response, error) {
	req, err := http.NewRequest("POST", url, strings.NewReader(body.Encode()))
	if err != nil {
		return nil, err
	}

	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	resp, err := httpClient.Do(req)
	if err != nil {
		return nil, err
	}

	return resp, nil
}
