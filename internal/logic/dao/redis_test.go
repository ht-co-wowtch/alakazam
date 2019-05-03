package dao

import (
	"context"
	"gitlab.com/jetfueltw/cpw/alakazam/internal/logic/conf"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"gitlab.com/jetfueltw/cpw/alakazam/internal/logic/model"
)

var d *Dao

func TestMain(m *testing.M) {
	if err := conf.Read("../../../cmd/logic/logic-example.yml"); err != nil {
		panic(err)
	}
	d = &Dao{
		c:           conf.Conf,
		kafkaPub:    nil,
		redis:       newRedis(conf.Conf.Redis),
		redisExpire: int32(conf.Conf.Redis.Expire / time.Second),
	}
}

func TestDaopingRedis(t *testing.T) {
	err := d.pingRedis(context.Background())
	assert.Nil(t, err)
}

func TestDaoAddMapping(t *testing.T) {
	var (
		c      = context.Background()
		key    = "test_key"
		server = "test_server"
	)
	err := d.AddMapping(c, "test", server)
	assert.Nil(t, err)
	err = d.AddMapping(c, key, server)
	assert.Nil(t, err)

	has, err := d.ExpireMapping(c, "test")
	assert.Nil(t, err)
	assert.NotEqual(t, false, has)
	has, err = d.ExpireMapping(c, key)
	assert.Nil(t, err)
	assert.NotEqual(t, false, has)

	res, err := d.ServersByKeys(c, []string{key})
	assert.Nil(t, err)
	assert.Equal(t, server, res[0])

	has, err = d.DelMapping(c, "test", server)
	assert.Nil(t, err)
	assert.NotEqual(t, false, has)
	has, err = d.DelMapping(c, key, server)
	assert.Nil(t, err)
	assert.NotEqual(t, false, has)
}

func TestDaoAddServerOnline(t *testing.T) {
	var (
		c      = context.Background()
		server = "test_server"
		online = &model.Online{
			RoomCount: map[string]int32{"room": 10},
		}
	)
	err := d.AddServerOnline(c, server, online)
	assert.Nil(t, err)

	r, err := d.ServerOnline(c, server)
	assert.Nil(t, err)
	assert.Equal(t, online.RoomCount["room"], r.RoomCount["room"])

	err = d.DelServerOnline(c, server)
	assert.Nil(t, err)
}
