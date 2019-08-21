package member

import (
	"encoding/json"
	"fmt"
	"github.com/alicebob/miniredis"
	goRedis "github.com/go-redis/redis"
	"github.com/stretchr/testify/assert"
	"gitlab.com/jetfueltw/cpw/alakazam/models"
	"gitlab.com/jetfueltw/cpw/micro/errdefs"
	"gitlab.com/jetfueltw/cpw/micro/id"
	"gitlab.com/jetfueltw/cpw/micro/redis"
	"os"
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
	err := c.set(&models.Member{
		Uid: uid,
	})
	assert.Nil(t, err)

	err = c.setBanned(uid, time.Duration(10)*time.Second)
	assert.Nil(t, err)

	unix, err := r.Get(keyBanned(uid)).Int64()
	ex := time.Unix(unix, 0)
	assert.Nil(t, err)
	assert.False(t, ex.IsZero())

	ok, err := c.isBanned(uid)
	assert.Nil(t, err)
	assert.True(t, ok)

	m, err := c.get(uid)

	assert.Nil(t, err)
	assert.False(t, m.IsMessage)
}

func TestIsBannedByFalse(t *testing.T) {
	ok, err := c.isBanned("1")
	assert.Nil(t, err)
	assert.False(t, ok)
}

func TestDelBanned(t *testing.T) {
	uid := "1"
	err := c.set(&models.Member{Uid: uid, IsMessage: true})
	assert.Nil(t, err)

	err = c.setBanned(uid, time.Minute)
	assert.Nil(t, err)

	err = c.delBanned(uid)
	assert.Nil(t, err)

	ok, err := c.isBanned(uid)
	assert.Nil(t, err)
	assert.False(t, ok)
}

func TestSetLogin(t *testing.T) {
	uid := id.UUid32()
	key := id.UUid32()
	name := "test"
	server := "server"
	member := &models.Member{Id: 1, Uid: uid, Name: name, Type: models.Player, IsMessage: true}
	err := c.login(member, key, "1", server)

	u := r.HGetAll(keyUid(uid)).Val()

	b, err := json.Marshal(member)
	if err != nil {
		t.Fatal(err)
	}

	assert.Nil(t, err)
	assert.Equal(t, map[string]string{
		key:        "1",
		hJsonKey:   string(b),
		hServerKey: server,
	}, u)

	expire := r.TTL(keyUid(uid)).Val()

	assert.Equal(t, c.expire, expire)
}

func TestRefreshUserExpire(t *testing.T) {
	uid := id.UUid32()
	r.Set(keyUid(uid), 1, time.Hour)

	ok, err := c.refreshExpire(uid)

	assert.True(t, ok)
	assert.Nil(t, err)

	m := r.TTL(keyUid(uid)).Val()

	assert.Equal(t, c.expire, m)
}

func TestDeleteUser(t *testing.T) {
	uid := id.UUid32()
	r.HSet(keyUid(uid), "key", "test")

	ok, err := c.logout(uid, "key")

	assert.Nil(t, err)
	assert.True(t, ok)
}

func TestGetSession(t *testing.T) {
	uid := id.UUid32()
	key := id.UUid32()
	name := "test"
	member := &models.Member{Id: 1, Uid: uid, Name: name, Type: models.Player, IsMessage: true}

	_ = c.login(member, key, "1", "test")

	h, err := c.getSession(uid, key)

	assert.Nil(t, err)
	assert.Equal(t, 1, h.Mid)
	assert.Equal(t, 1, h.Room)
	assert.Equal(t, name, h.Name)
	assert.Equal(t, 1, h.Mid)
	assert.Equal(t, models.Player, h.Type)
	assert.True(t, h.IsMessage)
}

func TestGetSessionByNil(t *testing.T) {
	uid := id.UUid32()
	key := id.UUid32()
	name := "test"
	member := &models.Member{Uid: uid, Name: name, Type: models.Player}

	_ = c.login(member, key, "1", "test")

	_, err := c.getSession(uid, "123")

	assert.Equal(t, errdefs.InvalidParameter(errUserNil, 1), err)
}

func TestGetUserName(t *testing.T) {
	uid := []string{"1", "2", "3", "4"}
	for _, v := range uid {
		if err := c.login(&models.Member{Uid: v, Name: v}, v, v, v); err != nil {
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

	i := r.HGet(keyUid(uid), key).Val()

	assert.Equal(t, roomId, i)

	m := r.TTL(keyUid(uid)).Val()

	assert.Equal(t, c.expire, m)
}
