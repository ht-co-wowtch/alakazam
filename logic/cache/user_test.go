package cache

import (
	. "github.com/smartystreets/goconvey/convey"
	"gitlab.com/jetfueltw/cpw/alakazam/logic/permission"
	"testing"
)

func TestSetUser(t *testing.T) {
	Convey("假設有一個會員", t, func() {
		Reset(mockRestart)

		uid := "82ea16cd2d6a49d887440066ef739669"
		key := "0b7f8111-8781-4574-8cb8-2eda0adb7598"
		roomId := "1000"
		name := "test"
		p := permission.PlayDefaultPermission

		c1 := mock.Command("HMSET", keyUidInfo(uid), key, roomId, hashNameKey, name, hashStatusKey, p, hashServerKey, "")
		c2 := mock.Command("EXPIRE", keyUidInfo(uid), expireSec)

		Convey("設定會員cache資料成功", func() {
			c1.Expect("")
			c2.Expect("")

			err := d.SetUser(uid, key, roomId, name, "", p)

			Convey("error == nil", func() {
				So(err, ShouldBeNil)
				So(mock.ExpectationsWereMet(), ShouldBeNil)
			})

			Convey("有uid,key,房間id,name,權限", func() {
				So(1, ShouldEqual, mock.Stats(c1))
				So(1, ShouldEqual, mock.Stats(c2))
			})
		})
	})
}

func TestRefreshUserExpire(t *testing.T) {
	Convey("刷新會員資料expire", t, func() {
		Reset(mockRestart)

		uid := "82ea16cd2d6a49d887440066ef739669"

		Convey("刷新cache expire成功", func() {
			c1 := mock.Command("EXPIRE", keyUidInfo(uid), expireSec).
				Expect([]byte(`true`))
			ok, err := d.RefreshUserExpire(uid)

			Convey("error == nil", func() {
				So(err, ShouldBeNil)
				So(mock.ExpectationsWereMet(), ShouldBeNil)
			})

			Convey("已延長expire", func() {
				So(1, ShouldEqual, mock.Stats(c1))
				So(ok, ShouldBeTrue)
			})
		})
	})
}

func TestDeleteUser(t *testing.T) {
	Convey("刪除某會員資料", t, func() {
		Reset(mockRestart)

		uid := "82ea16cd2d6a49d887440066ef739669"
		key := "0b7f8111-8781-4574-8cb8-2eda0adb7598"

		c := mock.Command("HDEL", keyUidInfo(uid), key)

		Convey("刪除cache成功", func() {
			c.Expect([]byte(`true`))
			ok, err := d.DeleteUser(uid, key)

			Convey("error == nil", func() {
				So(err, ShouldBeNil)
				So(mock.ExpectationsWereMet(), ShouldBeNil)
			})

			Convey("已刪除", func() {
				So(1, ShouldEqual, mock.Stats(c))
				So(ok, ShouldBeTrue)
			})
		})
	})
}
