package message

import (
	"encoding/json"
	"fmt"
	"github.com/go-redis/redis"
	"gitlab.com/jetfueltw/cpw/alakazam/app/logic/pb"
	"time"
)

const (
	roomTopKey = "room_top_msg_%d"

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

func keyRoomTopMsg(id int32) string {
	return fmt.Sprintf(roomTopKey, id)
}

func (c *cache) addTopMessage(msg *pb.PushMsg) error {
	sendAt := time.Unix(msg.SendAt, 0).Format("15:04:05")
	tx := c.c.Pipeline()
	for _, rid := range msg.Room {
		msg := Message{
			Id:      msg.Seq,
			Uid:     RootUid,
			Type:    topType,
			Name:    RootName,
			Message: msg.Message,
			Time:    sendAt,
		}

		b, err := json.Marshal(msg)
		if err != nil {
			return err
		}

		tx.Set(keyRoomTopMsg(rid), b, c.expiration)
	}
	_, err := tx.Exec()
	return err
}
