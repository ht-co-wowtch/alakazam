package dao

import (
	"time"

	"github.com/gomodule/redigo/redis"
	"gitlab.com/jetfueltw/cpw/alakazam/server/logic/conf"
	kafka "gopkg.in/Shopify/sarama.v1"
)

type Dao struct {
	c        *conf.Config
	kafkaPub kafka.SyncProducer
	redis    *redis.Pool

	// redis 過期時間
	redisExpire int32
}

func New(c *conf.Config) *Dao {
	d := &Dao{
		c:           c,
		kafkaPub:    newKafkaPub(c.Kafka),
		redis:       newRedis(c.Redis),
		redisExpire: int32(c.Redis.Expire / time.Second),
	}
	return d
}

func newKafkaPub(c *conf.Kafka) kafka.SyncProducer {
	kc := kafka.NewConfig()
	kc.Producer.RequiredAcks = kafka.WaitForAll
	kc.Producer.Retry.Max = 10               
	kc.Producer.Return.Successes = true
	pub, err := kafka.NewSyncProducer(c.Brokers, kc)
	if err != nil {
		panic(err)
	}
	return pub
}

func newRedis(c *conf.Redis) *redis.Pool {
	return &redis.Pool{
		MaxIdle:     c.Idle,
		MaxActive:   c.Active,
		IdleTimeout: c.IdleTimeout,
		Dial: func() (redis.Conn, error) {
			conn, err := redis.Dial(c.Network, c.Addr,
				redis.DialConnectTimeout(c.DialTimeout),
				redis.DialReadTimeout(c.ReadTimeout),
				redis.DialWriteTimeout(c.WriteTimeout),
			)
			if err != nil {
				return nil, err
			}
			return conn, nil
		},
	}
}

// Close close the resource.
func (d *Dao) Close() error {
	return d.redis.Close()
}

// ping redis是否活著
func (d *Dao) Ping() error {
	return d.pingRedis()
}
