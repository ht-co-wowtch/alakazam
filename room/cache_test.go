package room

import (
	"encoding/json"
	"fmt"
	"github.com/alicebob/miniredis"
	goRedis "github.com/go-redis/redis"
	"github.com/stretchr/testify/assert"
	"gitlab.com/jetfueltw/cpw/alakazam/models"
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
		Id:           1,
		IsMessage:    true,
		DayLimit:     day,
		DmlLimit:     dml,
		DepositLimit: amount,
	}
)

func TestSetRoom(t *testing.T) {
	err := c.set(room)

	assert.Nil(t, err)

	m := r.Get(keyRoom("1")).Val()

	b, err := json.Marshal(room)

	assert.Nil(t, err)
	assert.Equal(t, string(b), m)

	expire := r.TTL(keyRoom("1")).Val()

	assert.Equal(t, time.Hour, expire)
}

func TestGetRoom(t *testing.T) {
	_ = c.set(room)

	s, err := c.get("1")

	assert.Nil(t, err)
	assert.Equal(t, room, *s)
}

func TestAddServerOnline(t *testing.T) {
	unix := time.Now().Unix()
	server := &Online{
		Server:    "123",
		RoomCount: map[int32]int32{1: 1, 2: 2},
		Updated:   unix,
	}
	err := c.addOnline("123", server)

	assert.Nil(t, err)

	o, err := c.getOnline("123")

	assert.Equal(t, server, o)
}
