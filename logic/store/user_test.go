package store

import (
	"github.com/DATA-DOG/go-sqlmock"
	. "github.com/smartystreets/goconvey/convey"
	"testing"
)

func TestCreateUser(t *testing.T) {
	Convey("建立會員", t, func() {
		user := &User{Uid: "1", Name: "test", Avatar: "/", Permission: 100}

		c := mock.ExpectExec("^INSERT INTO members \\(uid, name, avatar, permission, create_at\\) VALUES \\(\\?, \\?, \\?, \\?, CURRENT_TIMESTAMP\\)").
			WithArgs(user.Uid, user.Name, user.Avatar, user.Permission)

		Convey("建立成功", func() {
			c.WillReturnResult(sqlmock.NewResult(1, 1))

			aff, err := store.CreateUser(user)

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

func TestFind(t *testing.T) {
	Convey("取會員資料", t, func() {
		expectedUser := &User{Uid: "1", Name: "test", Avatar: "/", Permission: 100}

		c := mock.ExpectQuery("^SELECT name, avatar, permission, is_blockade FROM members WHERE uid = \\?").
			WithArgs(expectedUser.Uid)

		Convey("有資料", func() {
			c.WillReturnRows(
				sqlmock.NewRows([]string{"name", "avatar", "permission", "is_blockade"}).
					AddRow(expectedUser.Name, expectedUser.Avatar, expectedUser.Permission, expectedUser.IsBlockade),
			)

			user, err := store.Find(expectedUser.Uid)

			Convey("sql執行成功", func() {
				So(err, ShouldBeNil)
				So(mock.ExpectationsWereMet(), ShouldBeNil)
			})

			Convey("回傳資料正確", func() {
				So(user, ShouldResemble, expectedUser)
			})
		})
	})
}

func TestUpdateUser(t *testing.T) {
	Convey("更新會員資料", t, func() {
		user := &User{Uid: "1", Name: "test", Avatar: "/"}

		c := mock.ExpectExec("UPDATE members SET name = \\?, avatar = \\? WHERE uid = \\?").
			WithArgs(user.Name, user.Avatar, user.Uid)

		Convey("更新成功", func() {
			c.WillReturnResult(sqlmock.NewResult(1, 1))

			aff, err := store.UpdateUser(user)

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
