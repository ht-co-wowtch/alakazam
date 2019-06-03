package cache

import (
	"github.com/rafaeljusto/redigomock"
	. "github.com/smartystreets/goconvey/convey"
	"gitlab.com/jetfueltw/cpw/alakazam/logic/permission"
	"strconv"
	"testing"
	"time"
)

func TestSetBanned(t *testing.T) {
	Convey("禁言某會員5分鐘", t, func() {
		Reset(mockRestart)

		uid := "123"
		sec := time.Duration(5) * time.Second

		Convey("禁言成功時", func() {
			c1 := mock.Command("SET", keyBannedInfo(uid), time.Now().Add(sec).Unix()).
				Expect("")
			c2 := mock.Command("EXPIRE", keyBannedInfo(uid), 5).
				Expect("")
			c3 := mock.Command("HINCRBY", keyUidInfo(uid), hashStatusKey, -permission.Message).
				Expect("")

			err := d.SetBanned(uid, 5)

			Convey("cache設定禁言資料完成", func() {
				So(err, ShouldBeNil)
				So(mock.ExpectationsWereMet(), ShouldBeNil)
			})

			Convey("cache已有會員禁言資料", func() {
				So(1, ShouldEqual, mock.Stats(c1))
			})

			Convey("禁言資料五分鐘後過期", func() {
				So(1, ShouldEqual, mock.Stats(c2))
			})

			Convey("會員沒有發話權限", func() {
				So(1, ShouldEqual, mock.Stats(c3))
			})
		})
	})
}

func TestGetBanned(t *testing.T) {
	Convey("取得會員禁言資料", t, func() {
		Reset(mockRestart)

		uid := "123"
		unix := time.Now().Unix()

		Convey("cache有禁言資料", func() {
			c := mockGetBanned(uid, []byte(strconv.FormatInt(unix, 10)))
			ex, ok, err := d.GetBanned(uid)

			Convey("error == nil", func() {
				So(err, ShouldBeNil)
				So(mock.ExpectationsWereMet(), ShouldBeNil)
				So(1, ShouldEqual, mock.Stats(c))
			})

			Convey("資料存在", func() {
				So(ok, ShouldBeTrue)
			})

			Convey("有過期時間", func() {
				So(unix, ShouldEqual, ex.Unix())
			})
		})

		Convey("cache沒有禁言資料", func() {
			c := mockGetBanned(uid, nil)
			ex, ok, err := d.GetBanned(uid)

			Convey("error == nil", func() {
				So(err, ShouldBeNil)
				So(mock.ExpectationsWereMet(), ShouldBeNil)
				So(1, ShouldEqual, mock.Stats(c))
			})

			Convey("資料不存在", func() {
				So(ok, ShouldBeFalse)
			})

			Convey("沒有過期時間", func() {
				So(ex.IsZero(), ShouldBeTrue)
			})
		})
	})
}

func TestDeleteBanned(t *testing.T) {
	Convey("刪除會員禁言資料", t, func() {
		Reset(mockRestart)

		uid := "123"

		Convey("cache有禁言資料", func() {
			c1 := mock.Command("DEL", keyBannedInfo(uid)).
				Expect("")
			c2 := mock.Command("HINCRBY", keyUidInfo(uid), hashStatusKey, permission.Message).
				Expect("")

			err := d.DelBanned(uid)

			Convey("error == nil", func() {
				So(err, ShouldBeNil)
				So(mock.ExpectationsWereMet(), ShouldBeNil)
			})

			Convey("會員增加發話權限", func() {
				So(1, ShouldEqual, mock.Stats(c1))
			})

			Convey("刪除資料成功", func() {
				So(1, ShouldEqual, mock.Stats(c2))
			})
		})
	})
}

func mockGetBanned(uid string, expect interface{}) *redigomock.Cmd {
	return mock.Command("GET", keyBannedInfo(uid)).
		Expect(expect)
}
