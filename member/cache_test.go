package member

import (
	"encoding/json"
	"fmt"
	"github.com/alicebob/miniredis"
	goRedis "github.com/go-redis/redis"
	"github.com/stretchr/testify/assert"
	"gitlab.com/jetfueltw/cpw/alakazam/models"
	"gitlab.com/jetfueltw/cpw/micro/id"
	"gitlab.com/jetfueltw/cpw/micro/redis"
	"os"
	"testing"
	"time"
)

var (
	r *goRedis.Client
	c *cache
)

func TestMain(m *testing.M) {
	s, err := miniredis.Run()
	if err != nil {
		fatalTestError("Error creating redis test : %v\n", err)
	}
	r = redis.New(&redis.Conf{
		Addr: s.Addr(),
	})
	c = &cache{
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
	ok, err := c.set(&models.Member{
		Uid: uid,
	})
	assert.Nil(t, err)
	assert.True(t, ok)

	ok, err = c.setBanned(uid, time.Duration(10)*time.Second)
	assert.Nil(t, err)
	assert.True(t, ok)

	unix, err := r.Get(keyBanned(uid)).Int64()
	ex := time.Unix(unix, 0)
	assert.Nil(t, err)
	assert.False(t, ex.IsZero())

	ok, err = c.isBanned(uid)
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
	ok, err := c.set(&models.Member{Uid: uid, IsMessage: true})
	assert.Nil(t, err)
	assert.True(t, ok)

	ok, err = c.setBanned(uid, time.Minute)
	assert.Nil(t, err)
	assert.True(t, ok)

	ok, err = c.delBanned(uid)
	assert.Nil(t, err)
	assert.True(t, ok)

	ok, err = c.isBanned(uid)
	assert.Nil(t, err)
	assert.False(t, ok)
}

func TestLogin(t *testing.T) {
	uid := id.UUid32()
	key := id.UUid32()
	name := "test"
	server := "server"
	member := &models.Member{Id: 1, Uid: uid, Name: name, Type: models.Player, IsMessage: true}
	err := c.login(member, key, server)

	u := r.HMGet(keyUid(uid), uidJsonHKey, uidNameHKey).Val()

	b, err := json.Marshal(member)
	if err != nil {
		t.Fatal(err)
	}

	assert.Nil(t, err)
	assert.Equal(t, string(b), u[0])
	assert.Equal(t, name, u[1])

	keys, err := r.HGetAll(keyUidWs(member.Uid)).Result()

	assert.Nil(t, err)
	assert.Equal(t, map[string]string{
		key: server,
	}, keys)

	expire := r.TTL(keyUid(uid)).Val()

	assert.Equal(t, c.expire, expire)

	expire = r.TTL(keyUidWs(uid)).Val()

	assert.Equal(t, c.expire, expire)
}

func TestRefreshUserExpire(t *testing.T) {
	uid := id.UUid32()
	r.Set(keyUid(uid), 1, time.Hour)
	r.Set(keyUidWs(uid), 1, time.Hour)

	err := c.refreshExpire(uid)

	assert.Nil(t, err)

	uidT := r.TTL(keyUid(uid)).Val()
	wsT := r.TTL(keyUidWs(uid)).Val()

	assert.Equal(t, c.expire, uidT)
	assert.Equal(t, c.expire, wsT)
}

func TestDeleteUser(t *testing.T) {
	uid := id.UUid32()
	r.HSet(keyUidWs(uid), "key1", "test")
	r.HSet(keyUidWs(uid), "key2", "test")

	ok, err := c.logout(uid, "key1")

	assert.Nil(t, err)
	assert.True(t, ok)

	keys, err := r.HGetAll(keyUidWs(uid)).Result()

	assert.Nil(t, err)
	assert.Equal(t, map[string]string{
		"key2": "test",
	}, keys)
}

func TestSetAndGet(t *testing.T) {
	uid := id.UUid32()
	name := "test"
	member := &models.Member{Id: 1, Uid: uid, Name: name, Type: models.Player, IsMessage: true}

	ok, err := c.set(member)
	assert.Nil(t, err)
	assert.True(t, ok)

	m, err := c.get(member.Uid)

	assert.Nil(t, err)
	assert.Equal(t, member.Id, m.Id)
	assert.Equal(t, member.Uid, m.Uid)
	assert.Equal(t, member.Name, m.Name)
	assert.Equal(t, member.Type, m.Type)
	assert.Equal(t, member.IsMessage, m.IsMessage)
}

func TestGetNil(t *testing.T) {
	_, err := c.get("test")

	assert.Equal(t, goRedis.Nil, err)
}

func TestCache_GetUserName(t *testing.T) {
	uid := []string{"1", "2", "3", "4"}
	for _, v := range uid {
		if err := c.login(&models.Member{Uid: v, Name: v}, v, v); err != nil {
			t.Fatal(err)
		}
	}

	name, err := c.getName(uid)

	assert.Nil(t, err)
	assert.Equal(t, map[string]string{"1": "1", "2": "2", "3": "3", "4": "4"}, name)
}
