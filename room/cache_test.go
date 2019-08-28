package room

import (
	"encoding/json"
	"fmt"
	"github.com/alicebob/miniredis"
	goRedis "github.com/go-redis/redis"
	"github.com/stretchr/testify/assert"
	"gitlab.com/jetfueltw/cpw/alakazam/message"
	"gitlab.com/jetfueltw/cpw/alakazam/models"
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

	roomTest = models.Room{
		Id:           1,
		IsMessage:    true,
		DayLimit:     day,
		DmlLimit:     dml,
		DepositLimit: amount,
	}

	roomTopMessageTest = message.Message{
		Id:      1,
		Uid:     "123",
		Name:    "test",
		Message: "test",
		Time:    time.Now().Format(time.RFC3339),
	}
)

func TestSetAndGetRoom(t *testing.T) {
	c.c.FlushAll()

	err := c.set(roomTest)
	assert.Nil(t, err)

	room, err := c.get(roomTest.Id)
	assert.Nil(t, err)
	assert.Equal(t, roomTest, room)

	expire := r.TTL(keyRoom(roomTest.Id)).Val()
	assert.Equal(t, roomExpired, expire)
}

func TestGetNil(t *testing.T) {
	c.c.FlushAll()

	room, err := c.get(2)

	assert.Equal(t, goRedis.Nil, err)
	assert.Equal(t, models.Room{}, room)
}

func TestSetChatAndGetChatRoom(t *testing.T) {
	c.c.FlushAll()

	b, _ := json.Marshal(roomTopMessageTest)

	err := c.setChat(roomTest, string(b))
	assert.Nil(t, err)

	room, err := c.getChat(roomTest.Id)
	message := room.TopMessage
	room.TopMessage = []byte(``)
	assert.Nil(t, err)
	assert.Equal(t, roomTest, room)

	assert.Equal(t, string(b), message)

	expire := r.TTL(keyRoom(roomTest.Id)).Val()
	assert.Equal(t, roomExpired, expire)
}

func TestGetChatNil(t *testing.T) {
	c.c.FlushAll()

	room, err := c.getChat(roomTest.Id)

	assert.Equal(t, goRedis.Nil, err)
	assert.Equal(t, models.Room{}, room)
}

func TestGetChatMessageNil(t *testing.T) {
	c.c.FlushAll()

	_, err := c.c.HSet(keyRoom(roomTest.Id), roomDataKey, `{"id":1}`).Result()
	assert.Nil(t, err)

	room, err := c.getChat(roomTest.Id)

	assert.Equal(t, models.Room{Id: 1}, room)
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
