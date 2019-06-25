package e2e

import (
	"encoding/json"
	"github.com/stretchr/testify/assert"
	"gitlab.com/jetfueltw/cpw/alakazam/logic/store"
	"gitlab.com/jetfueltw/cpw/alakazam/pkg/encoding/binary"
	pd "gitlab.com/jetfueltw/cpw/alakazam/protocol"
	"gitlab.com/jetfueltw/cpw/alakazam/protocol/grpc"
	"gitlab.com/jetfueltw/cpw/alakazam/test/internal/request"
	"gitlab.com/jetfueltw/cpw/micro/errdefs"
	"gitlab.com/jetfueltw/cpw/micro/id"
	"io"
	"net/http"
	"testing"
	"time"
)

var (
	uidA = "1d7eff72ab49470882833853875340c1"
	uidB = "e617b86221f1486480a5631a64da1355"
)

type body struct {
	Uid     string `json:"uid"`
	Message string `json:"message"`
	Name    string `json:"name"`
}

func TestPushMessage(t *testing.T) {
	roomId := id.UUid32()
	userA := request.DialAuth(t, roomId, uidA)
	userB := request.DialAuth(t, roomId, uidB)

	resp := userA.PushRoom("test")
	if resp.Error != nil {
		t.Fatal(resp.Error)
	}

	time.Sleep(time.Second * 2)

	var err error
	var p []grpc.Proto
	if p, err = userB.ReadMessage(); err != nil {
		t.Fatal(err)
	}

	var b body
	if err := json.Unmarshal(p[0].Body, &b); err != nil {
		t.Fatal(err)
	}

	assert.Len(t, p, 1)
	assert.Equal(t, pd.OpRaw, p[0].Op)
	assert.Equal(t, "test", b.Message)
	assert.Equal(t, uidA, b.Uid)
	assert.NotEmpty(t, b.Name)
}

func TestChangeRoom(t *testing.T) {
	roomId := id.UUid32()
	userA := request.DialAuth(t, roomId, uidA)
	userB := request.DialAuth(t, roomId, uidB)

	if err := userA.ChangeRoom(id.UUid32()); err != nil {
		t.Fatal(err)
	}
	resp := userA.PushRoom("test-3")
	if resp.Error != nil {
		t.Fatal(resp.Error)
	}
	if err := userA.ChangeRoom(roomId); err != nil {
		t.Fatal(err)
	}
	resp = userA.PushRoom("test-2")
	if resp.Error != nil {
		t.Fatal(resp.Error)
	}

	time.Sleep(time.Second * 2)

	var err error
	var p []grpc.Proto
	if p, err = userB.ReadMessage(); err != nil {
		t.Fatal(err)
	}

	var b body
	if err := json.Unmarshal(p[0].Body, &b); err != nil {
		t.Fatal(err)
	}

	assert.Len(t, p, 1)
	assert.Equal(t, "test-2", b.Message)
	assert.Equal(t, uidA, b.Uid)
}

func TestHeartbeat(t *testing.T) {
	userA := request.DialAuth(t, id.UUid32(), uidA)
	if err := userA.Heartbeat(); err != nil {
		t.Fatal(err)
	}
	online := binary.BigEndian.Int32(userA.Proto.Body)

	assert.Equal(t, pd.OpHeartbeatReply, userA.Proto.Op)
	assert.Equal(t, int32(1), online)
}

func TestNotHeartbeat(t *testing.T) {
	userA := request.DialAuth(t, id.UUid32(), uidA)
	err := userA.Read()

	assert.Equal(t, io.EOF, err)
}

func TestNotSetRoomId(t *testing.T) {
	ws, err := request.Dial()

	b := make([]byte, 100)
	_, err = ws.Read(b)

	assert.Equal(t, io.EOF, err)
}

func TestConnectionRoomError(t *testing.T) {
	_, err := request.DialAuthToken(id.UUid32(), "")

	assert.Equal(t, io.EOF, err)
}

func TestBroadcast(t *testing.T) {
	roomIdA := id.UUid32()
	roomIdB := id.UUid32()
	userA := request.DialAuth(t, roomIdA, uidA)
	userB := request.DialAuth(t, roomIdB, uidB)

	request.PushBroadcast([]string{roomIdA, roomIdB}, "test")
	time.Sleep(time.Second * 2)

	pA, err := userA.ReadMessage()
	if err != nil {
		t.Fatal(err)
	}
	pB, err := userB.ReadMessage()
	if err != nil {
		t.Fatal(err)
	}

	var bA body
	if err := json.Unmarshal(pA[0].Body, &bA); err != nil {
		t.Fatal(err)
	}
	var bB body
	if err := json.Unmarshal(pB[0].Body, &bB); err != nil {
		t.Fatal(err)
	}

	assert.Len(t, pA, 1)
	assert.Len(t, pB, 1)
	assert.Equal(t, "test", bA.Message)
	assert.Equal(t, "test", bB.Message)
	assert.Equal(t, "管理员", bB.Name)
}

func TestCreateRoom(t *testing.T) {
	roomId := id.UUid32()
	request.CreateRoom(store.Room{
		Id:        roomId,
		IsMessage: true,
	})

	userA := request.DialAuth(t, roomId, uidA)
	r := userA.PushRoom("test")

	assert.Nil(t, r.Error)
	assert.NotEmpty(t, userA.Uid)
}

func TestCreateRoomBanned(t *testing.T) {
	roomId := id.UUid32()
	request.CreateRoom(store.Room{
		Id:        roomId,
		IsMessage: false,
	})

	userA := request.DialAuth(t, roomId, uidA)
	r := userA.PushRoom("test")

	e := new(errdefs.Error)
	if err := json.Unmarshal(r.Body, e); err != nil {
		t.Fatal(err)
	}

	assert.Equal(t, http.StatusUnauthorized, r.StatusCode)
	assert.Equal(t, 10024014, e.Code)
	assert.Equal(t, "聊天室目前禁言状态，无法发言", e.Message)
}

func TestUpdateRoom(t *testing.T) {
	roomId := id.UUid32()
	request.CreateRoom(store.Room{
		Id:        roomId,
		IsMessage: false,
	})
	request.UpdateRoom(roomId, store.Room{
		IsMessage: true,
	})

	r := request.GetRoom(roomId)

	var p store.Room

	if err := json.Unmarshal(r.Body, &p); err != nil {
		t.Fatal(err)
	}

	assert.Equal(t, store.Room{
		Id:        roomId,
		IsMessage: true,
	}, p)
}
