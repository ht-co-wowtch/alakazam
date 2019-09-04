package message

import (
	"fmt"
	"github.com/go-redis/redis"
	"gitlab.com/jetfueltw/cpw/alakazam/app/logic/pb"
	"time"
)

const (
	OK = "OK"

	messageKey = "room_message_%d_%d"
)

func keyMessage(rid int32, hour int) string {
	return fmt.Sprintf(messageKey, rid, hour)
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
	at := time.Unix(msg.SendAt, 0)
	return c.c.ZAdd(keyMessage(msg.Room[0], at.Hour()), redis.Z{Score: float64(msg.SendAt), Member: msg.Msg}).Err()
}
