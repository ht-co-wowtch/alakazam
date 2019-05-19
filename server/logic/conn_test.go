package logic

import (
	"database/sql"
	"github.com/DATA-DOG/go-sqlmock"
	"github.com/stretchr/testify/assert"
	"gitlab.com/jetfueltw/cpw/alakazam/server/logic/business"
	"testing"
)

func TestConnect(t *testing.T) {
	mockDB.ExpectQuery("^SELECT").
		WithArgs(sqlmock.AnyArg()).
		WillReturnError(sql.ErrNoRows)
	mockDB.ExpectExec("^INSERT INTO members").
		WithArgs(sqlmock.AnyArg(), business.PlayDefaultPermission).
		WillReturnResult(sqlmock.NewResult(0, 1))

	c, err := l.Connect("", []byte(`{"token":"","room_id":"1000"}`))
	assert.Nil(t, err)

	assert.NotEmpty(t, c.Uid)
	assert.NotEmpty(t, c.Key)
	assert.NotEmpty(t, c.Name)
	assert.NotEmpty(t, c.Permission)
	assert.Equal(t, "1000", c.RoomId)

	r, n, s := getUser(t, c.Uid, c.Key)

	assert.Equal(t, c.RoomId, r)
	assert.Equal(t, c.Name, n)
	assert.Equal(t, c.Permission, s)
}

func TestDisconnect(t *testing.T) {
	addUser(t, "123", "", "", "", 0)

	has, err := l.Disconnect("123", "", "")

	assert.True(t, has)
	assert.Nil(t, err)
}

func TestChangeRoom(t *testing.T) {
	addUser(t, "456", "", "1000", "", 0)

	err := l.ChangeRoom("456", "", "1001")
	assert.Nil(t, err)

	r, _, _ := getUser(t, "456", "")
	assert.Equal(t, "1001", r)
}

func TestHeartbeat(t *testing.T) {
	addUser(t, "789", "", "1000", "", 0)

	err := l.Heartbeat("789", "", "", "", "")
	assert.Nil(t, err)
}
