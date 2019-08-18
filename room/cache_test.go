package room

import (
	"fmt"
	"github.com/alicebob/miniredis"
	goRedis "github.com/go-redis/redis"
	"github.com/stretchr/testify/assert"
	"gitlab.com/jetfueltw/cpw/alakazam/models"
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
		c: r,
	}
	exitStatus := m.Run()
	s.Close()
	os.Exit(exitStatus)
}

func fatalTestError(fmtStr string, args ...interface{}) {
	fmt.Fprintf(os.Stderr, fmtStr, args...)
	os.Exit(1)
}

var (
	day    = 1
	dml    = 100
	amount = 500

	room = models.Room{
		Uuid:         id.UUid32(),
		IsMessage:    true,
		DayLimit:     day,
		DmlLimit:     dml,
		DepositLimit: amount,
	}
)

func TestSetRoom(t *testing.T) {
	err := c.set(room)

	assert.Nil(t, err)

	m := r.HGetAll(keyRoom(room.Uuid)).Val()

	assert.Equal(t, map[string]string{
		hashPermissionKey:  "5",
		hashLimitDayKey:    strconv.Itoa(day),
		hashLimitDmlKey:    strconv.Itoa(dml),
		hashLimitAmountKey: strconv.Itoa(amount),
	}, m)

	expire := r.TTL(keyRoom(room.Uuid)).Val()

	assert.Equal(t, time.Hour, expire)
}

func TestGetRoomByMoney(t *testing.T) {
	_ = c.set(room)

	dy, dl, a, err := c.getMoney(room.Uuid)

	assert.Nil(t, err)
	assert.Equal(t, day, dy)
	assert.Equal(t, dml, dl)
	assert.Equal(t, amount, a)
}

func TestGetRoom(t *testing.T) {
	_ = c.set(room)

	s, err := c.get(room.Uuid)

	assert.Nil(t, err)
	assert.Equal(t, s, s)
}

func TestAddServerOnline(t *testing.T) {
	unix := time.Now().Unix()
	server := &Online{
		Server:    "123",
		RoomCount: map[string]int32{"1": 1, "2": 2},
		Updated:   unix,
	}
	err := c.addOnline("123", server)

	assert.Nil(t, err)

	o, err := c.getOnline("123")

	assert.Equal(t, server, o)
}
