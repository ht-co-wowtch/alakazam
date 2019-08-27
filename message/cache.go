package message

import (
	"github.com/go-redis/redis"
	"time"
)

const (
	OK = "OK"
)

type cache struct {
	c          *redis.Client
	expiration time.Duration
}

func newCache(c *redis.Client) *cache {
	return &cache{
		c:          c,
		expiration: time.Hour,
	}
}
