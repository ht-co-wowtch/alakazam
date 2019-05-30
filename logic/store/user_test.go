package store

import (
	"github.com/DATA-DOG/go-sqlmock"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestCreateUser(t *testing.T) {
	user := &User{Uid: "1", Name: "test", Avatar: "/", Permission: 100}

	mock.ExpectExec("^INSERT INTO members \\(uid, name, avatar, permission, create_at\\) VALUES \\(\\?, \\?, \\?, \\?, CURRENT_TIMESTAMP\\)").
		WithArgs(user.Uid, user.Name, user.Avatar, user.Permission).
		WillReturnResult(sqlmock.NewResult(1, 1))

	aff, err := store.CreateUser(user)

	assert.Nil(t, err)
	assert.Equal(t, aff, int64(1))
}

func TestFind(t *testing.T) {
	expectedUser := &User{Uid: "1", Name: "test", Avatar: "/", Permission: 100}

	mock.ExpectQuery("^SELECT name, avatar, permission, is_blockade FROM members WHERE uid = \\?").
		WithArgs(expectedUser.Uid).
		WillReturnRows(
			sqlmock.NewRows([]string{"name", "avatar", "permission", "is_blockade"}).
				AddRow(expectedUser.Name, expectedUser.Avatar, expectedUser.Permission, expectedUser.IsBlockade),
		)
	user, err := store.Find(expectedUser.Uid)

	assert.Nil(t, err)
	assert.Equal(t, expectedUser, user)
}

func TestUpdateUser(t *testing.T) {
	user := &User{Uid: "1", Name: "test", Avatar: "/"}

	mock.ExpectExec("UPDATE members SET name = \\?, avatar = \\? WHERE uid = \\?").
		WithArgs(user.Name, user.Avatar, user.Uid).
		WillReturnResult(sqlmock.NewResult(1, 1))

	aff, err := store.UpdateUser(user)

	assert.Nil(t, err)
	assert.Equal(t, int64(1), aff)
}
