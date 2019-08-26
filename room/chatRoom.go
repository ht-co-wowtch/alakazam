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
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
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

var (
	errNoRoom       = status.Error(codes.NotFound, "room not found")
	errSetRoomCache = status.Error(codes.Internal, "set room cache")
	errRoomClose    = status.Error(codes.NotFound, "room is close")
)

func (c *chat) Connect(server string, token []byte) (*models.Member, string, int, error) {
	var params struct {
		// 帳務中心+版的認證token
		Token string `json:"token" binding:"required"`

		// client要進入的room
		RoomID int `json:"room_id" binding:"required"`
	}

	if err := json.Unmarshal(token, &params); err != nil {
		return nil, "", 0, status.Error(codes.InvalidArgument, err.Error())
	}
	if err := v.Struct(&params); err != nil {
		return nil, "", 0, status.Error(codes.InvalidArgument, err.Error())
	}

	room, err := c.cache.get(params.RoomID)

	if err != nil {
		if err != redis.Nil {
			return nil, "", 0, status.Error(codes.Internal, err.Error())
		}

		room, err = c.db.GetRoom(params.RoomID)
		if err != nil {
			if err == sql.ErrNoRows {
				return nil, "", 0, errNoRoom
			}
			return nil, "", 0, status.Error(codes.Internal, err.Error())
		}
		if err := c.cache.set(room); err != nil {
			return nil, "", 0, errSetRoomCache
		}
	}
	if !room.Status {
		return nil, "", 0, errRoomClose
	}

	user, key, err := c.member.Login(params.RoomID, params.Token, server)
	if err != nil {
		return nil, "", 0, status.Error(codes.Internal, err.Error())
	}
	return user, key, params.RoomID, nil
}

func (r *chat) Disconnect(uid, key string) (bool, error) {
	ok, err := r.member.Logout(uid, key)
	if err != nil {
		return ok, status.Error(codes.Internal, err.Error())
	}
	return ok, nil
}

func (c *chat) Heartbeat(uid, key, name, server string) error {
	if err := c.member.Heartbeat(uid); err != nil {
		return status.Error(codes.Internal, err.Error())
	}
	return nil
}

func (c *chat) RenewOnline(server string, roomCount map[int32]int32) (map[int32]int32, error) {
	online := &Online{
		Server:    server,
		RoomCount: roomCount,
		Updated:   time.Now().Unix(),
	}
	if err := c.cache.addOnline(server, online); err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}
	return roomCount, nil
}

func (r *chat) IsMessage(rid int, uid string) error {
	room, err := r.cache.get(rid)
	if err != nil {
		if err == redis.Nil {
			// TODO error
			return errors.New("房間讀取錯誤")
		}
		return err
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
