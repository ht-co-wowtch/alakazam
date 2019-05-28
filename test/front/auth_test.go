package front

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/stretchr/testify/assert"
	"gitlab.com/jetfueltw/cpw/alakazam/errors"
	pd "gitlab.com/jetfueltw/cpw/alakazam/protocol"
	"gitlab.com/jetfueltw/cpw/alakazam/protocol/grpc"
	"gitlab.com/jetfueltw/cpw/alakazam/test/internal/protocol"
	"gitlab.com/jetfueltw/cpw/alakazam/test/internal/request"
	"gitlab.com/jetfueltw/cpw/alakazam/test/internal/run"
	"golang.org/x/net/websocket"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"testing"
	"time"
)

func TestMain(m *testing.M) {
	r := run.Run("../config")
	i := m.Run()
	r()
	os.Exit(i)
}

// 進入房間成功
func TestAuth(t *testing.T) {
	roomId := "1000"
	a, err := request.DialAuth(roomId)
	if err != nil {
		assert.Error(t, err)
		return
	}
	shouldBeAuthReply(t, a, roomId)
}

// 只連線不進房間
func TestNotAuth(t *testing.T) {
	ws, err := request.Dial()
	if err != nil {
		assert.Error(t, err)
		return
	}
	shouldBeCloseConnection(err, ws, t)
}

// 進入房間失敗
func TestAuthError(t *testing.T) {
	_, err := request.DialAuthUserByAuthApi("1111", "", func(i *http.Request) (response *http.Response, e error) {
		return &http.Response{
			StatusCode: http.StatusBadRequest,
			Body:       ioutil.NopCloser(bytes.NewReader([]byte(``))),
		}, nil
	})
	assert.Equal(t, io.EOF, err)
}

// 房間心跳成功
func TestHeartbeat(t *testing.T) {
	a, err := request.DialAuth("1001")
	if err != nil {
		assert.Error(t, err)
		return
	}
	shouldBeHeartbeatReply(t, a, givenHeartbeat())
}

// 房間不心跳
func TestNotHeartbeat(t *testing.T) {
	a, err := request.DialAuth("1002")
	if err != nil {
		assert.Error(t, err)
		return
	}
	shouldBeTimeoutConnection(err, a, t)
}

// 封鎖
func TestRoomBlockade(t *testing.T) {
	a, err := request.DialAuth("1003")
	assert.Nil(t, err)
	assert.NotEmpty(t, a.Uid)

	r := request.SetBlockade(a.Uid, "測試")
	assert.Empty(t, r.Body)

	a, err = request.DialAuthUser(a.Uid, "1003")
	assert.Nil(t, err)

	e := new(errors.Error)
	json.Unmarshal(a.Proto.Body, e)

	assert.Equal(t, 10024011, e.Code)
	assert.Equal(t, "您在封鎖状态，无法进入聊天室", e.Message)
}

// 解封鎖
func TestRoomDeleteBlockade(t *testing.T) {
	roomId := "1004"
	a, err := request.DialAuth(roomId)
	uid := a.Uid
	assert.Nil(t, err)
	assert.NotEmpty(t, a.Uid)

	request.SetBlockade(a.Uid, "測試")

	a, err = request.DialAuthUser(a.Uid, roomId)
	assert.Nil(t, err)

	e := new(errors.Error)
	json.Unmarshal(a.Proto.Body, e)
	assert.Equal(t, 10024011, e.Code)

	request.DeleteBlockade(uid)

	a, err = request.DialAuthUser(uid, roomId)
	assert.Nil(t, err)

	assert.NotEmpty(t, a.Uid)
}

// 切換房間
func TestChangeRoom(t *testing.T) {
	a, err := request.DialAuth("1005")
	if err != nil {
		assert.Fail(t, err.Error())
		return
	}

	proto := new(grpc.Proto)
	proto.Op = pd.OpChangeRoom
	proto.Body = []byte(`1001`)

	if err = protocol.Write(a.Wr, proto); err != nil {
		assert.Fail(t, err.Error())
		return
	}
	if err := protocol.Read(a.Rd, a.Proto); err != nil {
		assert.Fail(t, err.Error())
		return
	}

	assert.Equal(t, pd.OpChangeRoomReply, a.Proto.Op)
	assert.Equal(t, "1001", string(a.Proto.Body))
}

func shouldBeTimeoutConnection(err error, a request.Auth, t *testing.T) {
	fmt.Println(time.Now())
	err = protocol.Read(a.Rd, a.Proto)
	fmt.Println(time.Now())
	assert.Equal(t, io.EOF, err)
}

func shouldBeCloseConnection(err error, ws *websocket.Conn, t *testing.T) {
	b := make([]byte, 100)
	_, err = ws.Read(b)
	assert.Equal(t, io.EOF, err)
}

func givenHeartbeat() *grpc.Proto {
	hbProto := new(grpc.Proto)
	hbProto.Op = pd.OpHeartbeat
	hbProto.Body = nil
	return hbProto
}

func shouldBeAuthReply(t *testing.T, a request.Auth, roomId string) {
	assert.Equal(t, pd.OpAuthReply, a.Proto.Op)
	assert.NotEmpty(t, a.Reply.Uid)
	assert.NotEmpty(t, a.Reply.Key)
	assert.NotEmpty(t, a.Reply.Token)
	assert.Equal(t, roomId, a.Reply.RoomId)
	assert.True(t, a.Reply.Permission.Message)
	assert.True(t, a.Reply.Permission.SendBonus)
	assert.True(t, a.Reply.Permission.GetBonus)
	assert.True(t, a.Reply.Permission.SendFollow)
	assert.True(t, a.Reply.Permission.GetFollow)
}

func shouldBeHeartbeatReply(t *testing.T, a request.Auth, hbProto *grpc.Proto) {
	fmt.Println("send heartbeat")
	if err := protocol.Write(a.Wr, hbProto); err != nil {
		assert.Error(t, err)
		return
	}
	if err := protocol.Read(a.Rd, a.Proto); err != nil {
		assert.Error(t, err)
		return
	}
	fmt.Println("heartbeat Reply")
	assert.Equal(t, pd.OpHeartbeatReply, a.Proto.Op)
}
