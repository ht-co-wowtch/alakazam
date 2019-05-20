package cache

import (
	"fmt"
	_ "github.com/go-sql-driver/mysql"
	"github.com/gomodule/redigo/redis"
	_ "github.com/mattn/go-sqlite3"
	"gitlab.com/jetfueltw/cpw/alakazam/server/logic/conf"
	"time"
)

const (
	// user id的前綴詞，用於存儲在redis當key
	prefixUidInfo = "uid_%s"

	// user 禁言key的前綴詞
	prefixBannedInfo = "b_%s"

	// server name的前綴詞，用於存儲在redis當key
	prefixServerOnline = "server_%s"

	// user hash table name key
	hashNameKey = "name"

	// user hash table status key
	hashStatusKey = "status"

	// user hash table server key
	hashServerKey = "server"
)

func keyUidInfo(uid string) string {
	return fmt.Sprintf(prefixUidInfo, uid)
}

func keyBannedInfo(uid string) string {
	return fmt.Sprintf(prefixBannedInfo, uid)
}

func keyServerOnline(key string) string {
	return fmt.Sprintf(prefixServerOnline, key)
}

type Cache struct {
	*redis.Pool

	expire int32
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

// ping redis是否活著
func (d *Cache) Ping() error {
	conn := d.Get()
	_, err := conn.Do("SET", "PING", "PONG")
	conn.Close()
	return err
}
