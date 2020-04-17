package message

import (
	"fmt"
	"github.com/go-redis/redis"
	"gitlab.com/jetfueltw/cpw/alakazam/app/logic/pb"
	"gitlab.com/jetfueltw/cpw/alakazam/errors"
	"strconv"
	"time"
)

const (
	OK = "OK"

	prefix = "ala"

	messageKey = prefix + ":room_message_%d"
)

var (
	messageExpire = 2 * time.Hour
)

func keyMessage(rid int32) string {
	return fmt.Sprintf(messageKey, rid)
}

type Cache interface {
	getMessage(rid int32, at time.Time) ([]string, error)
	addMessages(rid int32, msg []interface{}) error
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

type ZM interface {
	Score() float64
}

func (c *cache) addMessages(rid int32, msg []interface{}) error {
	z := make([]redis.Z, len(msg))
	for i, v := range msg {
		data, ok := v.(ZM)
		if !ok {
			return errors.New("not implementation message.ZM")
		}
		z[i] = redis.Z{
			Score:  data.Score(),
			Member: v,
		}
	}
	return c.c.ZAdd(keyMessage(rid), z...).Err()
}

func (c *cache) addMessage(msg *pb.PushMsg) error {
	for _, rid := range msg.Room {
		if err := c.c.ZAdd(keyMessage(rid), redis.Z{Score: float64(msg.SendAt), Member: msg.Msg}).Err(); err != nil {
			return err
		}
	}
	return nil
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
