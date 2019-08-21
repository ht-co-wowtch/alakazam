package member

import (
	"fmt"
	"github.com/alicebob/miniredis"
	goRedis "github.com/go-redis/redis"
	"github.com/stretchr/testify/assert"
	"gitlab.com/jetfueltw/cpw/alakazam/models"
	"gitlab.com/jetfueltw/cpw/micro/errdefs"
	"gitlab.com/jetfueltw/cpw/micro/id"
	"gitlab.com/jetfueltw/cpw/micro/redis"
	"os"
	"strconv"
	"testing"
	"time"
)

var (
	r *goRedis.Client
	c *Cache
)

func TestMain(m *testing.M) {
	s, err := miniredis.Run()
	if err != nil {
		fatalTestError("Error creating redis test : %v\n", err)
	}
	r = redis.New(&redis.Conf{
		Addr: s.Addr(),
	})
	c = &Cache{
		c:      r,
		expire: time.Second * 10,
	}
	exitStatus := m.Run()
	s.Close()
	os.Exit(exitStatus)
}

func fatalTestError(fmtStr string, args ...interface{}) {
	fmt.Fprintf(os.Stderr, fmtStr, args...)
	os.Exit(1)
}

func TestSetBanned(t *testing.T) {
	uid := id.UUid32()
	r.HSet(keyUidInfo(uid), hashStatusKey, 0)

	err := c.setBanned(uid, time.Duration(10)*time.Second)
	assert.Nil(t, err)

	unix, err := r.Get(keyBannedInfo(uid)).Int64()
	ex := time.Unix(unix, 0)
	assert.Nil(t, err)
	assert.False(t, ex.IsZero())

	status, err := r.HGet(keyUidInfo(uid), hashStatusKey).Int()
	assert.Equal(t, -1, status)
	assert.Nil(t, err)
}

func TestSetBannedByExist(t *testing.T) {
	uid := id.UUid32()
	r.HSet(keyUidInfo(uid), hashStatusKey, 0)

	c.setBanned(uid, time.Duration(10)*time.Second)
	err := c.setBanned(uid, time.Duration(10)*time.Second)
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

	ex, ok, err := c.getBanned(uid)

	assert.Nil(t, err)
	assert.True(t, ok)
	assert.Equal(t, unix, ex.Unix())
}

func TestGetBannedEmpty(t *testing.T) {
	ex, ok, err := c.getBanned(id.UUid32())

	assert.Nil(t, err)
	assert.False(t, ok)
	assert.True(t, ex.IsZero())
}

func TestDelBanned(t *testing.T) {
	uid := id.UUid32()
	r.HSet(keyUidInfo(uid), hashStatusKey, 2)
	c.setBanned(uid, time.Duration(10)*time.Second)
	err := c.delBanned(uid)

	assert.Nil(t, err)

	bi := r.Exists(keyBannedInfo(uid)).Val()
	status, _ := r.HGet(keyUidInfo(uid), hashStatusKey).Int()

	assert.Equal(t, int64(0), bi)
	assert.Equal(t, 2, status)
}

func TestSetUser(t *testing.T) {
	uid := id.UUid32()
	key := id.UUid32()
	roomId := id.UUid32()
	name := "test"
	server := "server"
	member := &models.Member{Id: 1, Uid: uid, Name: name, Type: models.Player}
	err := c.set(member, key, roomId, server)

	u := r.HGetAll(keyUidInfo(uid)).Val()

	assert.Nil(t, err)
	assert.Equal(t, map[string]string{
		key:           roomId,
		hMidKey:       "1",
		hashNameKey:   name,
		hashStatusKey: strconv.Itoa(models.PlayStatus),
		hashServerKey: server,
	}, u)

	expire := r.TTL(keyUidInfo(uid)).Val()

	assert.Equal(t, c.expire, expire)
}

func TestRefreshUserExpire(t *testing.T) {
	uid := id.UUid32()
	r.Set(keyUidInfo(uid), 1, time.Hour)

	ok, err := c.refreshUserExpire(uid)

	assert.True(t, ok)
	assert.Nil(t, err)

	m := r.TTL(keyUidInfo(uid)).Val()

	assert.Equal(t, c.expire, m)
}

func TestDeleteUser(t *testing.T) {
	uid := id.UUid32()
	r.HSet(keyUidInfo(uid), "key", "test")

	ok, err := c.deleteUser(uid, "key")

	assert.Nil(t, err)
	assert.True(t, ok)
}

func TestGetUser(t *testing.T) {
	uid := id.UUid32()
	key := id.UUid32()
	roomId := id.UUid32()
	name := "test"
	member := &models.Member{Id: 1, Uid: uid, Name: name, Type: models.Player}

	_ = c.set(member, key, roomId, "test")

	h, err := c.get(uid, key)

	assert.Nil(t, err)
	assert.Equal(t, roomId, h.Room)
	assert.Equal(t, name, h.Name)
	assert.Equal(t, models.PlayStatus, h.Status)
	assert.Equal(t, 1, h.Mid)
}

func TestGetUserBuNil(t *testing.T) {
	uid := id.UUid32()
	key := id.UUid32()
	roomId := id.UUid32()
	name := "test"
	member := &models.Member{Uid: uid, Name: name, Type: models.Player}

	_ = c.set(member, key, roomId, "test")

	_, err := c.get(uid, "123")

	assert.Equal(t, errdefs.InvalidParameter(errUserNil, 1), err)
}

func TestGetUserName(t *testing.T) {
	uid := []string{"1", "2", "3", "4"}
	for _, v := range uid {
		if err := r.HSet(keyUidInfo(v), hashNameKey, v).Err(); err != nil {
			t.Fatal(err)
		}
	}

	name, err := c.getName(uid)

	assert.Nil(t, err)
	assert.Equal(t, uid, name)
}

func TestChangeRoom(t *testing.T) {
	uid := id.UUid32()
	key := id.UUid32()
	roomId := id.UUid32()

	err := c.changeRoom(uid, key, roomId)

	assert.Nil(t, err)

	i := r.HGet(keyUidInfo(uid), key).Val()

	assert.Equal(t, roomId, i)

	m := r.TTL(keyUidInfo(uid)).Val()

	assert.Equal(t, c.expire, m)
}

// BenchmarkGetUserName-4   	   10000	    174115 ns/op
func BenchmarkGetUserName(b *testing.B) {
	uid := []string{"1", "2", "3", "4", "5", "6", "7", "8", "9", "10"}
	for _, v := range uid {
		r.HSet(keyUidInfo(v), hashNameKey, v)
	}
	for i := 0; i < b.N; i++ {
		c.getName(uid)
	}
}
