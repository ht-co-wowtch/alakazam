package dao

import (
	_ "github.com/go-sql-driver/mysql"
	"github.com/gomodule/redigo/redis"
	_ "github.com/mattn/go-sqlite3"
	"gitlab.com/jetfueltw/cpw/alakazam/server/logic/conf"
	"time"
)

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
