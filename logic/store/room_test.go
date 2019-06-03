package store

import (
	"github.com/DATA-DOG/go-sqlmock"
	. "github.com/smartystreets/goconvey/convey"
	"testing"
)

func TestCreateRoom(t *testing.T) {
	Convey("設定一個房間", t, func() {
		c := mock.ExpectExec("INSERT INTO rooms \\(room_id, is_message, is_bonus, is_follow, day_limit, amount_limit, dml_limit\\) VALUES (.+)").
			WithArgs(sqlmock.AnyArg(), true, false, false, 5, 1000, 100)

		Convey("設定成功", func() {
			c.WillReturnResult(sqlmock.NewResult(1, 1))

			aff, err := store.CreateRoom(Room{
				IsMessage: true,
				Limit: Limit{
					Day:    5,
					Amount: 1000,
					Dml:    100,
				},
			})

			Convey("sql執行成功", func() {
				So(err, ShouldBeNil)
				So(mock.ExpectationsWereMet(), ShouldBeNil)
			})

			Convey("回傳create affected == 1", func() {
				So(err, ShouldBeNil)
				So(aff, ShouldEqual, 1)
			})
		})
	})
}

func TestGetRoom(t *testing.T) {
	Convey("取房間設定", t, func() {
		roomId := "82ea16cd2d6a49d887440066ef739669"
		room := Room{
			IsMessage: true,
			Limit: Limit{
				Day:    1,
				Amount: 1000,
				Dml:    100,
			},
		}

		c := mock.ExpectQuery("^SELECT \\* FROM rooms WHERE room_id = \\?").
			WithArgs(roomId)

		Convey("取成功", func() {
			c.WillReturnRows(
				sqlmock.NewRows([]string{"room_id", "is_message", "is_bonus", "is_follow", "day_limit", "amount_limit", "dml_limit"}).
					AddRow(room.RoomId, room.IsMessage, false, false, room.Limit.Day, room.Limit.Amount, room.Limit.Dml),
			)

			r, err := store.GetRoom(roomId)

			Convey("sql執行成功", func() {
				So(err, ShouldBeNil)
				So(mock.ExpectationsWereMet(), ShouldBeNil)
			})

			Convey("回傳的資料正確", func() {
				So(err, ShouldBeNil)
				So(r, ShouldResemble, room)
			})
		})
	})
}
