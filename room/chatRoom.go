package room

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"github.com/go-redis/redis"
	"gitlab.com/jetfueltw/cpw/alakazam/client"
	"gitlab.com/jetfueltw/cpw/alakazam/errors"
	"gitlab.com/jetfueltw/cpw/alakazam/member"
	"gitlab.com/jetfueltw/cpw/alakazam/models"
	"gitlab.com/jetfueltw/cpw/micro/errdefs"
	"gopkg.in/go-playground/validator.v8"
	"time"
)

var v *validator.Validate

func init() {
	v = validator.New(&validator.Config{TagName: "binding"})
}

type Chat interface {
	Connect(server string, token []byte) (*models.Member, string, int, error)
	Disconnect(uid, key string) (bool, error)
	Heartbeat(uid, key, name, server string) error
	RenewOnline(server string, roomCount map[int32]int32) (map[int32]int32, error)
	IsMessage(rid int, uid string) error
}

type chat struct {
	cache  Cache
	db     models.IChat
	member member.Chat
	cli    *client.Client
}

func NewChat(db models.IChat, cache *redis.Client, member member.Chat, cli *client.Client) Chat {
	return &chat{
		db:     db,
		cache:  newCache(cache),
		member: member,
		cli:    cli,
	}
}

func (c *chat) Connect(server string, token []byte) (*models.Member, string, int, error) {
	var params struct {
		// 帳務中心+版的認證token
		Token string `json:"token" binding:"required"`

		// client要進入的room
		RoomID int `json:"room_id" binding:"required"`
	}

	if err := json.Unmarshal(token, &params); err != nil {
		return nil, "", 0, err
	}
	if err := v.Struct(&params); err != nil {
		return nil, "", 0, err
	}

	room, err := c.getChat(params.RoomID)
	if err != nil {
		return nil, "", 0, err
	}
	if !room.Status {
		return nil, "", 0, errors.ErrRoomClose
	}

	user, key, err := c.member.Login(params.RoomID, params.Token, server)
	if err != nil {
		return nil, "", 0, err
	}
	return user, key, params.RoomID, nil
}

func (c *chat) getChat(id int) (models.Room, error) {
	room, err := c.cache.get(id)
	if err != nil {
		if err != redis.Nil {
			return models.Room{}, err
		}

		room, err = c.db.GetRoom(id)
		if err != nil {
			if err == sql.ErrNoRows {
				return models.Room{}, errors.ErrNoRoom
			}
			return models.Room{}, err
		}
		if err := c.cache.set(room); err != nil {
			return models.Room{}, err
		}
	}
	return room, nil
}

func (r *chat) Disconnect(uid, key string) (bool, error) {
	return r.member.Logout(uid, key)
}

func (c *chat) Heartbeat(uid, key, name, server string) error {
	return c.member.Heartbeat(uid)
}

func (c *chat) RenewOnline(server string, roomCount map[int32]int32) (map[int32]int32, error) {
	online := &Online{
		Server:    server,
		RoomCount: roomCount,
		Updated:   time.Now().Unix(),
	}
	err := c.cache.addOnline(server, online)
	return roomCount, err
}

func (c *chat) IsMessage(rid int, uid string) error {
	room, err := c.getChat(rid)
	if err != nil {
		return err
	}
	if !room.Status {
		return errors.ErrRoomClose
	}
	if !room.IsMessage {
		return errors.ErrRoomNoMessage
	}
	money, err := c.cli.GetDepositAndDml(room.DayLimit, uid)
	if err != nil {
		return err
	}
	if float64(room.DmlLimit) > money.Dml || float64(room.DepositLimit) > money.Deposit {
		msg := fmt.Sprintf(errors.ErrRoomLimit, room.DayLimit, room.DepositLimit, room.DmlLimit)
		return errdefs.Forbidden(errors.New(msg), 4035)
	}
	return nil
}
