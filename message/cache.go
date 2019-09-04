package message

import (
	"fmt"
	"github.com/go-redis/redis"
	"gitlab.com/jetfueltw/cpw/alakazam/app/logic/pb"
	"strconv"
	"time"
)

const (
	OK = "OK"

	messageKey = "room_message_%d"
)

var (
	messageExpire = 2 * time.Hour
)

func keyMessage(rid int32) string {
	return fmt.Sprintf(messageKey, rid)
}

type Cache interface {
	getMessage(rid int32, at time.Time) ([]string, error)
}

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

func (c *cache) addMessage(msg *pb.PushMsg) error {
	return c.c.ZAdd(keyMessage(msg.Room[0]), redis.Z{Score: float64(msg.SendAt), Member: msg.Msg}).Err()
}

func (c *cache) getMessage(rid int32, at time.Time) ([]string, error) {
	return c.c.ZRevRangeByScore(keyMessage(rid), redis.ZRangeBy{
		Max:    "(" + strconv.FormatInt(at.Unix(), 10),
		Min:    strconv.FormatInt(at.Add(-messageExpire).Unix(), 10),
		Offset: 0,
		Count:  20,
	}).Result()
}

func (c *cache) delMessage(keys []string) error {
	tx := c.c.Pipeline()
	for _, k := range keys {
		tx.ZRemRangeByScore(k, "-inf", strconv.FormatInt(time.Now().Add(-messageExpire).Unix(), 10))
	}
	_, err := tx.Exec()
	return err
}

func (c *cache) getMessageExistsKey() ([]string, error) {
	return c.c.Keys("room_message_*").Result()
}
