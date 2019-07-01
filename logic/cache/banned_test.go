package cache

import (
	"github.com/stretchr/testify/assert"
	"gitlab.com/jetfueltw/cpw/micro/id"
	"testing"
	"time"
)

func TestSetBanned(t *testing.T) {
	uid := id.UUid32()
	r.HSet(keyUidInfo(uid), hashStatusKey, 0)

	err := c.SetBanned(uid, 10)
	assert.Nil(t, err)

	unix, err := r.Get(keyBannedInfo(uid)).Int64()
	ex := time.Unix(unix, 0)
	assert.Nil(t, err)
	assert.False(t, ex.IsZero())

	status, err := r.HGet(keyUidInfo(uid), hashStatusKey).Int()
	assert.Equal(t, -1, status)
	assert.Nil(t, err)
}

func TestGetBanned(t *testing.T) {
	uid := id.UUid32()
	unix := time.Now().Unix()
	r.Set(keyBannedInfo(uid), unix, 10*time.Second)

	ex, ok, err := c.GetBanned(uid)

	assert.Nil(t, err)
	assert.True(t, ok)
	assert.Equal(t, unix, ex.Unix())
}

func TestGetBannedEmpty(t *testing.T) {
	ex, ok, err := c.GetBanned(id.UUid32())

	assert.Nil(t, err)
	assert.False(t, ok)
	assert.True(t, ex.IsZero())
}

func TestDelBanned(t *testing.T) {
	uid := id.UUid32()
	r.HSet(keyUidInfo(uid), hashStatusKey, 2)
	c.SetBanned(uid, 10)
	err := c.DelBanned(uid)

	assert.Nil(t, err)

	bi := r.Exists(keyBannedInfo(uid)).Val()
	status, _ := r.HGet(keyUidInfo(uid), hashStatusKey).Int()

	assert.Equal(t, int64(0), bi)
	assert.Equal(t, 2, status)
}
