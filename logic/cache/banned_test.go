package cache

import (
	"github.com/rafaeljusto/redigomock"
	"github.com/stretchr/testify/assert"
	"gitlab.com/jetfueltw/cpw/alakazam/logic/permission"
	"strconv"
	"testing"
	"time"
)

func TestSetBanned(t *testing.T) {
	uid := "123"

	sec := time.Duration(5) * time.Second
	c1 := mock.Command("SET", keyBannedInfo(uid), time.Now().Add(sec).Unix()).
		Expect("")
	c2 := mock.Command("HINCRBY", keyUidInfo(uid), hashStatusKey, -permission.Message).
		Expect("")
	c3 := mock.Command("EXPIRE", keyBannedInfo(uid), 5).
		Expect("")

	err := d.SetBanned(uid, 5)

	assert.Nil(t, err)
	assert.Nil(t, mock.ExpectationsWereMet())
	assert.Equal(t, 1, mock.Stats(c1))
	assert.Equal(t, 1, mock.Stats(c2))
	assert.Equal(t, 1, mock.Stats(c3))

	mock.Clear()
}

func TestGetBanned(t *testing.T) {
	uid := "123"
	unix := time.Now().Unix()

	c1 := mockGetBanned(uid, []byte(strconv.FormatInt(unix, 10)))
	ex, ok, err := d.GetBanned(uid)

	assert.Nil(t, err)
	assert.Nil(t, mock.ExpectationsWereMet())
	assert.Equal(t, 1, mock.Stats(c1))
	assert.True(t, ok)
	assert.Equal(t, unix, ex.Unix())

	mock.Clear()
}

func TestGetBannedEmpty(t *testing.T) {
	uid := "123"

	c1 := mockGetBanned(uid, nil)
	ex, ok, err := d.GetBanned(uid)

	assert.Nil(t, err)
	assert.Nil(t, mock.ExpectationsWereMet())
	assert.Equal(t, 1, mock.Stats(c1))
	assert.False(t, ok)
	assert.True(t, ex.IsZero())

	mock.Clear()
}

func TestDeleteBanned(t *testing.T) {
	uid := "123"

	c1 := mock.Command("DEL", keyBannedInfo(uid)).
		Expect("")
	c2 := mock.Command("HINCRBY", keyUidInfo(uid), hashStatusKey, permission.Message).
		Expect("")

	err := d.DelBanned(uid)

	assert.Nil(t, err)
	assert.Nil(t, mock.ExpectationsWereMet())
	assert.Equal(t, 1, mock.Stats(c1))
	assert.Equal(t, 1, mock.Stats(c2))

	mock.Clear()
}

func mockGetBanned(uid string, expect interface{}) *redigomock.Cmd {
	return mock.Command("GET", keyBannedInfo(uid)).
		Expect(expect)
}
