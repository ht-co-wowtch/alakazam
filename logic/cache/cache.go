package cache

import (
	"fmt"
	"github.com/go-redis/redis"
	_ "github.com/go-sql-driver/mysql"
	_ "github.com/mattn/go-sqlite3"
	redism "gitlab.com/jetfueltw/cpw/micro/redis"
	"time"
)

const (
	// user id的前綴詞，用於存儲在redis當key
	prefixUidInfo = "uid_%s"

	// user 禁言key的前綴詞
	prefixBannedInfo = "b_%s"

	// server name的前綴詞，用於存儲在redis當key
	prefixServerOnline = "server_%s"

	// 房間的前綴詞，用於存儲在redis當key
	prefixRoom = "room_%s"

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

func keyRoom(key string) string {
	return fmt.Sprintf(prefixRoom, key)
}

type Cache struct {
	c *redis.Client

	expire time.Duration
}

func NewRedis(c *redism.Conf) *Cache {
	return &Cache{
		c:      redism.New(c),
		expire: time.Minute * 30,
	}
}

// ping redis是否活著
func (c *Cache) Ping() error {
	return c.c.Ping().Err()
}

func (c *Cache) Close() error {
	return c.c.Close()
}
