package comet

import (
	"context"
	"encoding/json"
	"github.com/spf13/viper"
	"github.com/stretchr/testify/assert"
	"gitlab.com/jetfueltw/cpw/alakazam/app/comet/conf"
	"gitlab.com/jetfueltw/cpw/alakazam/app/comet/pb"
	logicpb "gitlab.com/jetfueltw/cpw/alakazam/app/logic/pb"
	"gitlab.com/jetfueltw/cpw/alakazam/pkg/encoding/binary"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"io"
	"testing"
)

func TestVisitorConnectionRoom(t *testing.T) {
	conn := &fakeWsConn{}
	session := newLoginSession(false, false, false)
	err := connectionWebSocket(session, conn, t)

	var close struct {
		Message string `json:"message"`
	}
	_ = json.Unmarshal(conn.body[0], &close)

	var connect logicpb.Connect
	_ = json.Unmarshal(conn.body[1], &connect)

	assert.False(t, connect.Status)
	assert.Equal(t, connect.Message, "请先登入会员")
	assert.Equal(t, close.Message, "请先登入会员")
	assert.Equal(t, err, io.EOF)
	assert.Len(t, conn.body, 2)
}

func TestGuestConnectionRoom(t *testing.T) {
	conn := &fakeWsConn{}
	session := newLoginSession(true, false, false)
	_ = connectionWebSocket(session, conn, t)

	var b logicpb.Connect
	_ = json.Unmarshal(conn.body[0], &b)

	assert.True(t, b.Status)
	assert.False(t, b.Permission.IsMessage)
	assert.False(t, b.Permission.IsRedEnvelope)
	assert.Equal(t, b.PermissionMessage.IsMessage, "请先登入会员")
	assert.Equal(t, b.PermissionMessage.IsRedEnvelope, "请先登入会员")
	assert.Len(t, conn.body, 1)
}

func TestMemberConnectionRoom(t *testing.T) {
	conn := &fakeWsConn{}
	session := newLoginSession(true, true, true)
	_ = connectionWebSocket(session, conn, t)

	var b logicpb.Connect
	_ = json.Unmarshal(conn.body[0], &b)

	assert.True(t, b.Status)
	assert.True(t, b.Permission.IsMessage)
	assert.True(t, b.Permission.IsRedEnvelope)
	assert.Equal(t, b.PermissionMessage.IsMessage, "")
	assert.Equal(t, b.PermissionMessage.IsRedEnvelope, "")
	assert.Len(t, conn.body, 1)
}

func newLoginSession(isLogin, isMessage, isRedEnvelope bool) *fakeLogicRpc {
	var message, redEnvelopeMsg string
	if !isMessage {
		message = "请先登入会员"
	}
	if !isRedEnvelope {
		redEnvelopeMsg = "请先登入会员"
	}
	return &fakeLogicRpc{
		isLogin:        isLogin,
		isMessage:      isMessage,
		isRedEnvelope:  isRedEnvelope,
		message:        message,
		redEnvelopeMsg: redEnvelopeMsg,
	}
}

func connectionWebSocket(rpc *fakeLogicRpc, conn *fakeWsConn, t *testing.T) error {
	c, err := conf.New(viper.GetViper())
	if err != nil {
		t.Fatal(err)
	}
	server := &Server{
		c:     c,
		round: NewRound(c),
		logic: rpc,
	}
	_, _, err = server.authWebsocket(context.TODO(), conn, &Channel{}, &pb.Proto{})
	return err
}

type fakeWsConn struct {
	body [][]byte
}

func (c fakeWsConn) WriteMessage(msgType int, msg []byte) error { return nil }
func (c fakeWsConn) WriteHeader(msgType int, length int) error  { return nil }

func (c *fakeWsConn) WriteBody(b []byte) error {
	c.body = append(c.body, b)
	return nil
}

func (c fakeWsConn) ReadMessage() (int, []byte, error) {
	buf := make([]byte, 10)
	binary.BigEndian.PutInt32(buf[pb.PackOffset:], 10)
	binary.BigEndian.PutInt16(buf[pb.HeaderOffset:], int16(pb.RawHeaderSize))
	binary.BigEndian.PutInt32(buf[pb.OpOffset:], pb.OpAuth)
	return 0, buf, nil
}

func (c fakeWsConn) Peek(n int) ([]byte, error) { return make([]byte, n), nil }
func (c fakeWsConn) Flush() error               { return nil }
func (c fakeWsConn) Close() error               { return nil }

type fakeLogicRpc struct {
	isLogin        bool
	isMessage      bool
	isRedEnvelope  bool
	message        string
	redEnvelopeMsg string
}

func (f fakeLogicRpc) Ping(ctx context.Context, in *logicpb.PingReq, opts ...grpc.CallOption) (*logicpb.PingReply, error) {
	return nil, nil
}

func (f fakeLogicRpc) Connect(ctx context.Context, in *logicpb.ConnectReq, opts ...grpc.CallOption) (*logicpb.ConnectReply, error) {
	if !f.isLogin {
		return nil, status.Error(codes.FailedPrecondition, "请先登入会员")
	}
	return &logicpb.ConnectReply{
		Connect: &logicpb.Connect{
			Uid:    "0d641b03d4d548dbb3a73a2197811261",
			Key:    "a37f15491b1f452e96d87791bf068d6f",
			Status: true,
			RoomID: 1,
			Permission: &logicpb.Permission{
				IsMessage:     f.isMessage,
				IsRedEnvelope: f.isRedEnvelope,
			},
			PermissionMessage: &logicpb.PermissionMessage{
				IsMessage:     f.message,
				IsRedEnvelope: f.redEnvelopeMsg,
			},
		},
	}, nil
}

func (f fakeLogicRpc) Disconnect(ctx context.Context, in *logicpb.DisconnectReq, opts ...grpc.CallOption) (*logicpb.DisconnectReply, error) {
	return nil, nil
}

func (f fakeLogicRpc) ChangeRoom(ctx context.Context, in *logicpb.ChangeRoomReq, opts ...grpc.CallOption) (*logicpb.ChangeRoomReply, error) {
	return nil, nil
}

func (f *fakeLogicRpc) Heartbeat(ctx context.Context, in *logicpb.HeartbeatReq, opts ...grpc.CallOption) (*logicpb.HeartbeatReply, error) {
	return nil, nil
}

func (f fakeLogicRpc) RenewOnline(ctx context.Context, in *logicpb.OnlineReq, opts ...grpc.CallOption) (*logicpb.OnlineReply, error) {
	return nil, nil
}
