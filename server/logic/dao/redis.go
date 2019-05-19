package dao

import (
	"fmt"
	"github.com/gomodule/redigo/redis"
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

type Cache struct {
	*redis.Pool

	expire int32
}

func keyUidInfo(uid string) string {
	return fmt.Sprintf(prefixUidInfo, uid)
}

func keyBannedInfo(uid string) string {
	return fmt.Sprintf(prefixBannedInfo, uid)
}

func keyServerOnline(key string) string {
	return fmt.Sprintf(prefixServerOnline, key)
}

// ping redis是否活著
func (d *Cache) Ping() error {
	conn := d.Get()
	_, err := conn.Do("SET", "PING", "PONG")
	conn.Close()
	return err
}