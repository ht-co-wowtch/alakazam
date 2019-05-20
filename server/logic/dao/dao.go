package dao

import (
	_ "github.com/go-sql-driver/mysql"
	"github.com/gomodule/redigo/redis"
	_ "github.com/mattn/go-sqlite3"
	"gitlab.com/jetfueltw/cpw/alakazam/server/logic/conf"
	kafka "gopkg.in/Shopify/sarama.v1"
	"time"
)

func NewKafkaPub(c *conf.Kafka) *Stream {
	kc := kafka.NewConfig()
	kc.Producer.RequiredAcks = kafka.WaitForAll
	kc.Producer.Retry.Max = 10
	kc.Producer.Return.Successes = true
	pub, err := kafka.NewSyncProducer(c.Brokers, kc)
	if err != nil {
		panic(err)
	}
	return &Stream{c: c, SyncProducer: pub}
}

func NewRedis(c *conf.Redis) *Cache {
	p := &redis.Pool{
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
	return &Cache{
		Pool:   p,
		expire: int32(c.Expire / time.Second),
	}
}