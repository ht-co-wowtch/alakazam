package room

import (
	"encoding/json"
	"fmt"
	"github.com/go-redis/redis"
	"gitlab.com/jetfueltw/cpw/alakazam/client"
	"gitlab.com/jetfueltw/cpw/alakazam/errors"
	"gitlab.com/jetfueltw/cpw/alakazam/member"
	"gitlab.com/jetfueltw/cpw/alakazam/models"
	"gitlab.com/jetfueltw/cpw/micro/errdefs"
	"gopkg.in/go-playground/validator.v8"
	"strconv"
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
	cache  *Cache
	db     *models.Store
	member *member.Member
	cli    *client.Client
}

func NewChat(db *models.Store, cache *redis.Client, member *member.Member, cli *client.Client) Chat {
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

	room, err := c.cache.get(strconv.Itoa(params.RoomID))
	if err != nil {
		return nil, "", 0, err
	}
	if room == nil {
		r, ok, err := c.db.GetRoom(params.RoomID)
		if err != nil {
			return nil, "", 0, err
		}
		if !ok {
			// TODO error
			return nil, "", 0, errors.New("房間不存在")
		}
		if err := c.cache.set(r); err != nil {
			// TODO error
			return nil, "", 0, errors.New("讀取房間失敗")
		}
	}
	if ! room.Status {
		// TODO error
		return nil, "", 0, errors.New("房間目前關閉中")
	}

	user, key, err := c.member.Login(params.RoomID, params.Token, server)
	if err != nil {
		return nil, "", 0, err
	}
	return user, key, params.RoomID, nil
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
	if err := c.cache.addOnline(server, online); err != nil {
		return nil, err
	}
	return roomCount, nil
}

func (r *chat) IsMessage(rid int, uid string) error {
	room, err := r.cache.get(strconv.Itoa(rid))
	if err != nil {
		return err
	}
	if room == nil {
		// TODO error
		return errors.New("房間讀取錯誤")
	}
	money, err := r.cli.GetDepositAndDml(room.DayLimit, uid)
	if err != nil {
		return err
	}
	if float64(room.DmlLimit) > money.Dml || float64(room.DepositLimit) > money.Deposit {
		e := errors.New(fmt.Sprintf("您无法发言，当前发言条件：前%d天充值不少于%d元；打码量不少于%d元", room.DayLimit, room.DepositLimit, room.DmlLimit))
		return errdefs.Unauthorized(e, 4)
	}
	return nil
}
