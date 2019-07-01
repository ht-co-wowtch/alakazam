package cache

import (
	"github.com/stretchr/testify/assert"
	"gitlab.com/jetfueltw/cpw/alakazam/logic/permission"
	"gitlab.com/jetfueltw/cpw/micro/id"
	"strconv"
	"testing"
	"time"
)

func TestSetUser(t *testing.T) {
	uid := id.UUid32()
	key := id.UUid32()
	roomId := id.UUid32()
	name := "test"
	server := "server"
	err := c.SetUser(uid, key, roomId, name, server, permission.PlayDefaultPermission)

	u := r.HGetAll(keyUidInfo(uid)).Val()

	assert.Nil(t, err)
	assert.Equal(t, map[string]string{
		key:           roomId,
		hashNameKey:   name,
		hashStatusKey: strconv.Itoa(permission.PlayDefaultPermission),
		hashServerKey: server,
	}, u)

	expire := r.TTL(keyUidInfo(uid)).Val()

	assert.Equal(t, c.expire, expire)
}

func TestRefreshUserExpire(t *testing.T) {
	uid := id.UUid32()
	r.Set(keyUidInfo(uid), 1, time.Hour)

	ok, err := c.RefreshUserExpire(uid)

	assert.True(t, ok)
	assert.Nil(t, err)

	m := r.TTL(keyUidInfo(uid)).Val()

	assert.Equal(t, c.expire, m)
}

func TestDeleteUser(t *testing.T) {
	uid := id.UUid32()
	r.HSet(keyUidInfo(uid), "key", "test")

	ok, err := c.DeleteUser(uid, "key")

	assert.Nil(t, err)
	assert.True(t, ok)
}

func TestGetUser(t *testing.T) {
	uid := id.UUid32()
	key := id.UUid32()
	roomId := id.UUid32()
	name := "test"

	_ = c.SetUser(uid, key, roomId, name, "test", permission.PlayDefaultPermission)

	r, n, s, err := c.GetUser(uid, key)

	assert.Nil(t, err)
	assert.Equal(t, roomId, r)
	assert.Equal(t, name, n)
	assert.Equal(t, permission.PlayDefaultPermission, s)
}

func TestGetUserBuNil(t *testing.T) {
	uid := id.UUid32()
	key := id.UUid32()
	roomId := id.UUid32()
	name := "test"

	_ = c.SetUser(uid, key, roomId, name, "test", permission.PlayDefaultPermission)

	_, _, _, err := c.GetUser(uid, "123")

	assert.Equal(t, errUserNil, err)
}

func TestChangeRoom(t *testing.T) {
	uid := id.UUid32()
	key := id.UUid32()
	roomId := id.UUid32()

	err := c.ChangeRoom(uid, key, roomId)

	assert.Nil(t, err)

	i := r.HGet(keyUidInfo(uid), key).Val()

	assert.Equal(t, roomId, i)

	m := r.TTL(keyUidInfo(uid)).Val()

	assert.Equal(t, c.expire, m)
}

func TestGetUserName(t *testing.T) {
	uid := []string{"1", "2", "3", "4"}
	for _, v := range uid {
		if err := r.HSet(keyUidInfo(v), hashNameKey, v).Err(); err != nil {
			t.Fatal(err)
		}
	}

	name, err := c.GetUserName(uid)

	assert.Nil(t, err)
	assert.Equal(t, uid, name)
}

// BenchmarkGetUserName-4   	   10000	    174115 ns/op
func BenchmarkGetUserName(b *testing.B) {
	uid := []string{"1", "2", "3", "4", "5", "6", "7", "8", "9", "10"}
	for _, v := range uid {
		r.HSet(keyUidInfo(v), hashNameKey, v)
	}
	for i := 0; i < b.N; i++ {
		c.GetUserName(uid)
	}
}
