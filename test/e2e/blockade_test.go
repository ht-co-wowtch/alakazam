package e2e

import (
	"encoding/json"
	"github.com/stretchr/testify/assert"
	pd "gitlab.com/jetfueltw/cpw/alakazam/protocol"
	"gitlab.com/jetfueltw/cpw/alakazam/test/internal/request"
	"gitlab.com/jetfueltw/cpw/micro/errdefs"
	"gitlab.com/jetfueltw/cpw/micro/id"
	"testing"
)

func TestBlockade(t *testing.T) {
	request.SetBlockade(uidA, "test")

	userA, _ := request.DialAuthToken(id.UUid32(), request.GetToken(t, uidA))

	defer userA.DeleteBanned()

	e := new(errdefs.Error)
	if err := json.Unmarshal(userA.Proto.Body, e); err != nil {
		t.Fatal(err)
	}

	assert.Equal(t, 10024011, e.Code)
	assert.Equal(t, "您在封鎖状态，无法进入聊天室", e.Message)
}

func TestDeleteBlockade(t *testing.T) {
	request.SetBlockade(uidA, "test")
	request.DeleteBlockade(uidA)

	userA := request.DialAuth(t, id.UUid32(), uidA)

	assert.Equal(t, pd.OpAuthReply, userA.Proto.Op)
}

func TestInRoomBlockade(t *testing.T) {
	roomId := id.UUid32()
	userA := request.DialAuth(t, roomId, uidA)

	defer userA.DeleteBanned()

	assert.Equal(t, pd.OpAuthReply, userA.Proto.Op)

	userA.SetBlockade("test")
	userA = request.DialAuth(t, roomId, uidA)

	e := new(errdefs.Error)
	if err := json.Unmarshal(userA.Proto.Body, e); err != nil {
		t.Fatal(err)
	}

	assert.Equal(t, 10024011, e.Code)
	assert.Equal(t, "您在封鎖状态，无法进入聊天室", e.Message)
}
