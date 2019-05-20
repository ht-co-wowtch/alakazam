package cache

import (
	"github.com/rafaeljusto/redigomock"
	"github.com/stretchr/testify/assert"
	"gitlab.com/jetfueltw/cpw/alakazam/logic/business"
	"strconv"
	"testing"
	"time"
)

func TestSetBanned(t *testing.T) {
	expire := 5
	uid := "123"

	mockSetBanned(uid, expire)
	err := d.SetBanned(uid, expire)

	assert.Nil(t, err)
}

func TestGetBanned(t *testing.T) {
	uid := "123"
	unix := time.Now().Unix()

	mockGetBanned(uid, []byte(strconv.FormatInt(unix, 10)))
	ex, ok, err := d.GetBanned(uid)

	assert.Nil(t, err)
	assert.True(t, ok)
	assert.Equal(t, unix, ex.Unix())
}

func TestGetBannedEmpty(t *testing.T) {
	uid := "123"

	mockGetBanned(uid, nil)
	ex, ok, err := d.GetBanned(uid)

	assert.Nil(t, err)
	assert.False(t, ok)
	assert.True(t, ex.IsZero())
}

func TestDeleteBanned(t *testing.T) {
	uid := "123"

	mockDelBanned(uid)
	err := d.DelBanned(uid)

	assert.Nil(t, err)
}

func mockSetBanned(uid string, expire int) {
	sec := time.Duration(expire) * time.Second
	mock.Command("SET", keyBannedInfo(uid), time.Now().Add(sec).Unix()).
		Expect("")
	mock.Command("HINCRBY", keyUidInfo(uid), hashStatusKey, -business.Message).
		Expect("")
	mock.Command("EXPIRE", keyBannedInfo(uid), expire).
		Expect("")
}

func mockGetBanned(uid string, expect interface{}) *redigomock.Cmd {
	return mock.Command("GET", keyBannedInfo(uid)).
		Expect(expect)
}

func mockDelBanned(uid string) {
	mock.Command("DEL", keyBannedInfo(uid)).
		Expect("")
	mock.Command("HINCRBY", keyUidInfo(uid), hashStatusKey, business.Message).
		Expect("")
}
