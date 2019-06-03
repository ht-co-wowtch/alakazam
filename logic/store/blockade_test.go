package store

import (
	"github.com/DATA-DOG/go-sqlmock"
	. "github.com/smartystreets/goconvey/convey"
	"testing"
)

func TestSetBlockade(t *testing.T) {
	Convey("設定封鎖", t, func() {
		uid := "82ea16cd2d6a49d887440066ef739669"

		Convey("封鎖某會員成功", func() {
			mockBanned(uid, true)

			aff, err := store.SetBlockade(uid, "")

			Convey("sql執行成功", func() {
				So(err, ShouldBeNil)
				So(mock.ExpectationsWereMet(), ShouldBeNil)
			})

			Convey("回傳update affected == 1", func() {
				So(aff, ShouldEqual, 1)
			})
		})
	})
}

func TestStore_DeleteBanned(t *testing.T) {
	Convey("解除封鎖", t, func() {
		uid := "82ea16cd2d6a49d887440066ef739669"

		Convey("解除某會員成功", func() {
			mockBanned(uid, false)

			aff, err := store.DeleteBanned(uid)

			Convey("sql執行成功", func() {
				So(err, ShouldBeNil)
				So(mock.ExpectationsWereMet(), ShouldBeNil)
			})

			Convey("回傳update affected == 1", func() {
				So(err, ShouldBeNil)
				So(aff, ShouldEqual, 1)
			})
		})
	})
}

func mockBanned(uid string, isBlockade bool) *sqlmock.ExpectedExec {
	return mock.ExpectExec("UPDATE members SET is_blockade = \\? WHERE uid = \\?").
		WithArgs(isBlockade, uid).
		WillReturnResult(sqlmock.NewResult(1, 1))
}
