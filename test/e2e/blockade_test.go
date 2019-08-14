package e2e

import (
	"encoding/json"
	"github.com/stretchr/testify/assert"
	"gitlab.com/jetfueltw/cpw/alakazam/comet/pb"
	"gitlab.com/jetfueltw/cpw/alakazam/test/internal/request"
	"gitlab.com/jetfueltw/cpw/micro/errdefs"
	"gitlab.com/jetfueltw/cpw/micro/id"
	"testing"
)

func TestRoomBlockade(t *testing.T) {
	roomId := id.UUid32()
	userA := request.DialAuth(t, roomId, uidA)

	defer userA.DeleteBanned()

	assert.Equal(t, pb.OpAuthReply, userA.Proto.Op)

	userA.SetBlockade("test")
	userA = request.DialAuth(t, roomId, uidA)

	e := new(errdefs.Error)
	if err := json.Unmarshal(userA.Proto.Body, e); err != nil {
		t.Fatal(err)
	}

	assert.Equal(t, 10024011, e.Code)
	assert.Equal(t, "您在封鎖状态，无法进入聊天室", e.Message)
}

func TestDeleteBlockade(t *testing.T) {
	roomId := id.UUid32()
	request.DialAuth(t, roomId, uidA)
	request.SetBlockade(uidA, "test")
	request.DeleteBlockade(uidA)

	userA := request.DialAuth(t, roomId, uidA)

	assert.Equal(t, pb.OpAuthReply, userA.Proto.Op)
}
