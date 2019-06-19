package front

import (
	"encoding/json"
	"fmt"
	. "github.com/smartystreets/goconvey/convey"
	"github.com/stretchr/testify/assert"
	"gitlab.com/jetfueltw/cpw/alakazam/client"
	"gitlab.com/jetfueltw/cpw/alakazam/errors"
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
	roomId := "3"
	Convey("假設要在房間發話", t, func() {
		a, err := request.DialAuth(roomId)
		if err != nil {
			assert.Fail(t, err.Error())
		}

		Reset(func() {
			roomId += "3"
		})

		Convey("當進入房間發話", func() {
			r := request.PushRoom(a.Uid, a.Key, "測試")

			Convey("應該發話成功", func() {
				So(http.StatusNoContent, ShouldEqual, r.StatusCode)
				So(r.Body, ShouldBeEmpty)
			})
		})

		Convey("當被禁言", func() {
			request.SetBanned(a.Uid, "測試", 3)
			r := request.PushRoom(a.Uid, a.Key, "測試")

			e := new(errors.Error)
			if err := json.Unmarshal(r.Body, e); err != nil {
				t.Fatal(err)
			}

			Convey("應該無法發話", func() {
				assert.Equal(t, http.StatusUnauthorized, r.StatusCode)
				assert.Equal(t, 10024013, e.Code)
				assert.Equal(t, "您在禁言状态，无法发言", e.Message)
			})
		})

		Convey("當禁言時效過期", func() {
			request.SetBanned(a.Uid, "測試", 2)
			time.Sleep(time.Second * 2)
			r := request.PushRoom(a.Uid, a.Key, "測試")

			Convey("應該發話成功", func() {
				So(http.StatusNoContent, ShouldEqual, r.StatusCode)
			})
		})

		Convey("當禁言被解除", func() {
			request.SetBanned(a.Uid, "測試", 3)
			request.DeleteBanned(a.Uid)
			r := request.PushRoom(a.Uid, a.Key, "測試")

			Convey("應該發話成功", func() {
				So(http.StatusNoContent, ShouldEqual, r.StatusCode)
			})
		})
	})
}

func TestDepositDml(t *testing.T) {
	Convey("假設房間有設定打碼與充值量發話限制", t, func() {
		uid := "009422e667c146379b3aa69f336ad4e5"

		Convey("當會員近一天打碼量有100,充值量有500", func() {
			giveUserDepositDmlMockApi(uid, 100, 500)
		})

		Convey("當房間限制近一天打碼量100,充值量要300", setRoomDepositDml(t, 1, 100, 300, func(a request.Auth) {
			Convey("應該發話成功", func() {
				r := request.PushRoom(a.Uid, a.Key, "測試")

				So(http.StatusNoContent, ShouldEqual, r.StatusCode)
			})
		}))

		Convey("當會員近一天打碼量有100,充值量有200", func() {
			giveUserDepositDmlMockApi(uid, 100, 200)
		})

		Convey("當房間限制近一天打碼量100,充值量要350", setRoomDepositDml(t, 1, 100, 350, func(a request.Auth) {
			Convey("應該不可發話", func() {
				r := request.PushRoom(a.Uid, a.Key, "測試")

				e := new(errors.Error)
				if err := json.Unmarshal(r.Body, e); err != nil {
					t.Fatal(err)
				}

				So(errors.MoneyError.Status, ShouldEqual, r.StatusCode)
				So(errors.MoneyError.Code, ShouldEqual, e.Code)
				So(errors.MoneyError.Format(1, 350, 100).Message, ShouldEqual, e.Message)
			})
		}))
	})
}

func setRoomDepositDml(t *testing.T, day, dml, deposit int, f func(a request.Auth)) func() {
	return func() {
		r := request.CreateRoom(store.Room{
			IsMessage: true,
			Limit: store.Limit{
				Day:    day,
				Dml:    dml,
				Amount: deposit,
			},
		})

		var room struct {
			Id string `json:"room_id"`
		}

		if err := json.Unmarshal(r.Body, &room); err != nil {
			t.Fatal(err)
		}

		a, err := request.DialAuthUser("009422e667c146379b3aa69f336ad4e5", room.Id)
		if err != nil {
			t.Fatal(err)
		}

		f(a)
	}
}

func giveUserDepositDmlMockApi(uid string, dml, amount int) {
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
		return request.ToResponse(b, http.StatusOK)
	})
}
