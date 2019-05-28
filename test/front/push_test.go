package front

import (
	"encoding/json"
	"fmt"
	"github.com/stretchr/testify/assert"
	"gitlab.com/jetfueltw/cpw/alakazam/errors"
	"gitlab.com/jetfueltw/cpw/alakazam/logic/client"
	"gitlab.com/jetfueltw/cpw/alakazam/logic/store"
	"gitlab.com/jetfueltw/cpw/alakazam/test/internal/request"
	"gitlab.com/jetfueltw/cpw/alakazam/test/internal/run"
	"net/http"
	"strings"
	"testing"
	"time"
)

// 房間訊息推送成功
func TestPushRoom(t *testing.T) {
	a, err := request.DialAuth("2000")
	if err != nil {
		assert.Fail(t, err.Error())
		return
	}
	r := request.PushRoom(a.Uid, a.Key, "測試")

	assert.Equal(t, http.StatusNoContent, r.StatusCode)
	assert.Empty(t, r.Body)
}

// 禁言
func TestPushRoomBanned(t *testing.T) {
	a := givenBanned(t, "2001")
	time.Sleep(time.Second * 4)
	r := request.PushRoom(a.Uid, a.Key, "測試")

	assert.Equal(t, http.StatusNoContent, r.StatusCode)
}

// 解除禁言
func TestPushRoomRemoveBanned(t *testing.T) {
	a := givenBanned(t, "2002")
	request.DeleteBanned(a.Uid)
	r := request.PushRoom(a.Uid, a.Key, "測試")

	assert.Equal(t, http.StatusNoContent, r.StatusCode)
}

func TestIsMoney(t *testing.T) {
	r := request.CreateRoom(store.Room{
		IsMessage: true,
		Limit: store.Limit{
			Day:    1,
			Dml:    100,
			Amount: 1000,
		},
	})

	assert.Equal(t, http.StatusOK, r.StatusCode)

	json.Unmarshal(r.Body, &room)

	a, err := request.DialAuthUser("009422e667c146379b3aa69f336ad4e5", room.Id)
	assert.Nil(t, err)

	giveUserMoneyMockApi(a.Uid, 0, 0)

	r = request.PushRoom(a.Uid, a.Key, "測試")

	e := new(errors.Error)
	json.Unmarshal(r.Body, e)

	assert.Equal(t, errors.MoneyError.Status, r.StatusCode)
	assert.Equal(t, errors.MoneyError.Code, e.Code)
	assert.Equal(t, errors.MoneyError.Format(1, 1000, 100).Message, e.Message)
}

func giveUserMoneyMockApi(uid string, dml, amount int) {
	run.AddClient(fmt.Sprintf("/members/%s/deposit-dml", uid), func(res *http.Request) (response *http.Response, e error) {
		authorization := res.Header.Get("Authorization")
		token := strings.Split(authorization, " ")

		if token[0] != "Bearer" {
			return nil, fmt.Errorf("Authorization not Bearer")
		}

		if token[1] == "" {
			return nil, fmt.Errorf("Authorization not token")
		}

		m := client.Money{
			Dml:     dml,
			Deposit: amount,
		}
		b, err := json.Marshal(m)
		if err != nil {
			return nil, err
		}
		return request.ToResponse(b)
	})
}

func givenBanned(t *testing.T, roomId string) request.Auth {
	a, err := request.DialAuth(roomId)
	assert.Nil(t, err)

	request.SetBanned(a.Uid, "測試", 3)
	r := request.PushRoom(a.Uid, a.Key, "測試")

	e := new(errors.Error)
	json.Unmarshal(r.Body, e)

	assert.Equal(t, http.StatusUnauthorized, r.StatusCode)
	assert.Equal(t, 10024013, e.Code)
	assert.Equal(t, "您在禁言状态，无法发言", e.Message)
	return a
}
