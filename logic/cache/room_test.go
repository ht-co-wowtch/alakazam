package cache

import (
	"github.com/gomodule/redigo/redis"
	"github.com/rafaeljusto/redigomock"
	. "github.com/smartystreets/goconvey/convey"
	"gitlab.com/jetfueltw/cpw/alakazam/logic/permission"
	"strconv"
	"testing"
)

func TestGetRoom(t *testing.T) {
	Convey("假設A房間有發話權限", t, func() {
		Reset(mockRestart)

		roomId := "a1b4bbec3f624ecf84858632a730c688"

		Convey("cache取A房間資料成功", func() {
			c := mockGetRoom(roomId, []byte(strconv.Itoa(permission.Message)))
			i, err := d.GetRoom(roomId)

			Convey("error == nil", func() {
				So(err, ShouldBeNil)
				So(mock.ExpectationsWereMet(), ShouldBeNil)
				So(1, ShouldEqual, mock.Stats(c))
			})

			Convey("有發話權限", func() {
				So(permission.Message, ShouldEqual, i)
			})
		})

		Convey("cache沒有A房間資料", func() {
			c := mockGetRoom(roomId, nil)
			i, err := d.GetRoom(roomId)

			Convey("error == nil", func() {
				So(mock.ExpectationsWereMet(), ShouldBeNil)
				So(1, ShouldEqual, mock.Stats(c))
			})

			Convey("沒有資料", func() {
				So(redis.ErrNil, ShouldEqual, err)
			})

			Convey("沒有發話權限", func() {
				So(0, ShouldEqual, i)
			})
		})
	})
}

func TestSetRoom(t *testing.T) {
	var p, day, dml, amount int

	Convey("假設有一個房間", t, func() {
		Reset(mockRestart)

		roomId := "a1b4bbec3f624ecf84858632a730c688"
		p = permission.Message
		day = 1
		dml = 1000
		amount = 100

		c1, c2 := mockSetRoom(roomId, p, day, dml, amount, 60*60)

		Convey("設定房間cache資料成功", func() {
			c1.Expect("")
			c2.Expect("")

			err := d.SetRoom(roomId, p, day, dml, amount)

			Convey("error == nil", func() {
				So(err, ShouldBeNil)
				So(mock.ExpectationsWereMet(), ShouldBeNil)
			})

			Convey("有發話權限,打碼量金額&天數限制,充值發話限制,過期時間", func() {
				So(1, ShouldEqual, mock.Stats(c1))
				So(1, ShouldEqual, mock.Stats(c2))
			})
		})
	})
}

func mockSetRoom(roomId string, p, day, dml, amount, expire int) (c1 *redigomock.Cmd, c2 *redigomock.Cmd) {
	c1 = mock.Command("HMSET", keyRoom(roomId), hashPermissionKey, p, hashLimitDayKey, day, hashLimitDmlKey, dml, hashLimitAmountKey, amount)
	c2 = mock.Command("EXPIRE", keyRoom(roomId), expire)
	return
}

func mockGetRoom(roomId string, expect interface{}) *redigomock.Cmd {
	return mock.Command("HGET", keyRoom(roomId), hashPermissionKey).
		Expect(expect)
}
