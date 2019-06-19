package front

import (
	"fmt"
	"github.com/google/uuid"
	. "github.com/smartystreets/goconvey/convey"
	"gitlab.com/jetfueltw/cpw/alakazam/errors"
	"gitlab.com/jetfueltw/cpw/alakazam/logic/store"
	"gitlab.com/jetfueltw/cpw/alakazam/test/internal/request"
	"net/http"
	"testing"
)

func TestRoomBanned(t *testing.T) {
	Convey("設定某房間相關權限", t, func() {
		id, _ := uuid.New().MarshalBinary()
		room := store.Room{Id: fmt.Sprintf("%x", id)}

		a, err := request.DialAuth(room.Id)

		if err != nil {
			t.Fatalf("request.DialAuth error(%v)", err)
		}

		Convey("設定可以發話", func() {
			room.IsMessage = true

			r := createAndPushRoom(t, room, &a)

			Convey("發話正常", func() {
				So(r.StatusCode, ShouldEqual, http.StatusNoContent)
			})
		})

		Convey("設定不可以發話", func() {
			room.IsMessage = false

			r := createAndPushRoom(t, room, &a)

			Convey("禁言中", func() {
				e := request.ToError(t, r.Body)
				e.Status = r.StatusCode

				So(e, ShouldResemble, errors.RoomBannedError)
			})
		})

		Convey("設定又可以發話", func() {
			room.IsMessage = true

			r := request.UpdateRoom(room.Id, room)
			if r.StatusCode != http.StatusNoContent {
				t.Fatal("request.UpdateRoom Fatal")
			}

			request.PushRoom(a.Uid, a.Key, "測試")

			Convey("發話正常", func() {
				So(r.StatusCode, ShouldEqual, http.StatusNoContent)
			})
		})

		Convey("設定有打碼量&充值發話限制", func() {
			room.IsMessage = true

			room.Limit = store.Limit{
				Day:    1,
				Amount: 1000,
				Dml:    100,
			}

			Convey("打碼量&充值足夠", func() {
				mockDepositAndDmlApi(t, &a, 2000, 200)

				r := createAndPushRoom(t, room, &a)

				Convey("可以發話", func() {
					So(r.StatusCode, ShouldEqual, http.StatusNoContent)
				})
			})

			Convey("打碼量&充值不足夠", func() {
				mockDepositAndDmlApi(t, &a, 500, 200)

				r := createAndPushRoom(t, room, &a)

				Convey("不可以發話", func() {
					e := request.ToError(t, r.Body)
					e.Status = r.StatusCode

					So(e, ShouldResemble, errors.MoneyError.Format(1, room.Limit.Amount, room.Limit.Dml))
				})
			})
		})
	})
}

func createAndPushRoom(t *testing.T, room store.Room, auth *request.Auth) request.Response {
	r := request.CreateRoom(room)
	if r.StatusCode != http.StatusOK {
		t.Fatal("request.CreateRoom Fatal")
	}
	return request.PushRoom(auth.Uid, auth.Key, "測試")
}
