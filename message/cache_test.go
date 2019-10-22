package message

import (
	"fmt"
	"github.com/alicebob/miniredis"
	goRedis "github.com/go-redis/redis"
	"github.com/stretchr/testify/assert"
	"gitlab.com/jetfueltw/cpw/alakazam/app/logic/pb"
	"gitlab.com/jetfueltw/cpw/micro/redis"
	"os"
	"testing"
	"time"
)

var (
	r *goRedis.Client
	c *cache

	pushMsgFake = []*pb.PushMsg{
		&pb.PushMsg{
			Room:   []int32{1},
			SendAt: time.Now().Add(-time.Minute).Unix(),
			Msg:    []byte(`{"id":1}`),
		},
		&pb.PushMsg{
			Room:   []int32{1},
			SendAt: time.Now().Unix(),
			Msg:    []byte(`{"id":2}`),
		},
	}
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

func TestAddMessage(t *testing.T) {
	r.FlushAll()
	for _, msg := range pushMsgFake {
		err := c.addMessage(msg)
		assert.Nil(t, err)
	}

	msg, err := r.ZRange(keyMessage(1), 0, 10).Result()

	assert.Nil(t, err)
	assert.Equal(t, []string{
		`{"id":1}`,
		`{"id":2}`,
	}, msg)
}

func TestGetMessage(t *testing.T) {
	r.FlushAll()
	for _, msg := range pushMsgFake {
		err := c.addMessage(msg)
		assert.Nil(t, err)
	}

	msg, err := c.getMessage(1, time.Now())

	assert.Nil(t, err)
	assert.Equal(t, []string{
		`{"id":1}`,
	}, msg)
}

func TestDelMessage(t *testing.T) {
	r.FlushAll()

	data := []*pb.PushMsg{
		&pb.PushMsg{
			Room:   []int32{1},
			SendAt: time.Now().Unix(),
			Msg:    []byte(`{"id":1}`),
		},
		&pb.PushMsg{
			Room:   []int32{1},
			SendAt: time.Now().Add(-time.Hour).Unix(),
			Msg:    []byte(`{"id":2}`),
		},
		&pb.PushMsg{
			Room:   []int32{1},
			SendAt: time.Now().Add(-2 * time.Hour).Unix(),
			Msg:    []byte(`{"id":3}`),
		},
		&pb.PushMsg{
			Room:   []int32{1},
			SendAt: time.Now().Add(-3 * time.Hour).Unix(),
			Msg:    []byte(`{"id":4}`),
		},
	}

	for _, msg := range data {
		err := c.addMessage(msg)
		assert.Nil(t, err)
	}

	err := c.delMessage([]string{"room_message_1"})
	assert.Nil(t, err)

	msg, err := r.ZRange(keyMessage(1), 0, 10).Result()
	assert.Nil(t, err)
	assert.Equal(t, []string{
		`{"id":2}`,
		`{"id":1}`,
	}, msg)
}
