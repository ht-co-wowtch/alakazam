package room

import (
	"database/sql"
	"encoding/json"
	"github.com/go-redis/redis"
	"gitlab.com/jetfueltw/cpw/alakazam/app/logic/pb"
	"gitlab.com/jetfueltw/cpw/alakazam/client"
	"gitlab.com/jetfueltw/cpw/alakazam/errors"
	"gitlab.com/jetfueltw/cpw/alakazam/member"
	"gitlab.com/jetfueltw/cpw/alakazam/message"
	"gitlab.com/jetfueltw/cpw/alakazam/models"
	"gitlab.com/jetfueltw/cpw/micro/log"
	"go.uber.org/zap"
	"gopkg.in/go-playground/validator.v8"
	"time"
)

var v *validator.Validate

func init() {
	v = validator.New(&validator.Config{TagName: "binding"})
}

type Chat interface {
	Connect(server string, token []byte) (*pb.ConnectReply, error)
	Disconnect(uid, key string) (bool, error)
	Heartbeat(uid, key, name, server string) error
	RenewOnline(server string, roomCount map[int32]int32) (map[int32]int32, error)
	GetRoom(rid int) (models.Room, error)
	GetUserMessageSession(uid string, rid int) (*models.Member, models.Room, error)
	ChangeRoom(uid string, rid int) (*pb.ChangeRoomReply, error)
	GetTopMessage(rid int) (message.Message, error)
	GetOnline(server string) (*Online, error)
}

type chat struct {
	cache            Cache
	db               models.IChat
	member           member.Chat
	cli              *client.Client
	heartbeatNanosec int64
}

func NewChat(db models.IChat, cache *redis.Client, member member.Chat, cli *client.Client, heartbeat int64) Chat {
	return &chat{
		db:               db,
		cache:            newCache(cache),
		member:           member,
		cli:              cli,
		heartbeatNanosec: heartbeat,
	}
}

func (c *chat) Connect(server string, token []byte) (*pb.ConnectReply, error) {
	var params struct {
		// 帳務中心+版的認證token
		Token string `json:"token" binding:"required"`

		// client要進入的room
		RoomID int `json:"room_id" binding:"required"`
	}

	if err := json.Unmarshal(token, &params); err != nil {
		return nil, err
	}
	if err := v.Struct(&params); err != nil {
		return nil, err
	}

	room, err := c.getChat(params.RoomID)
	if err != nil {
		return nil, err
	}
	if !room.Status {
		return nil, errors.ErrRoomClose
	}

	user, key, err := c.member.Login(params.RoomID, params.Token, server)
	if err != nil {
		return nil, err
	}
	if user.IsBlockade {
		return nil, errors.ErrBlockade
	}

	connect := newPbConnect(user, room, key, int32(params.RoomID))
	connect.Status = true

	return &pb.ConnectReply{
		Name:          user.Name,
		Heartbeat:     c.heartbeatNanosec,
		HeaderMessage: room.HeaderMessage,
		Connect:       connect,
	}, nil
}

func (c *chat) GetUserMessageSession(uid string, rid int) (*models.Member, models.Room, error) {
	user, err := c.member.GetMessageSession(uid)
	if err != nil {
		return nil, models.Room{}, err
	}
	room, err := c.GetRoom(rid)
	if err != nil {
		return nil, models.Room{}, err
	}
	return user, room, nil
}

func (c *chat) ChangeRoom(uid string, rid int) (*pb.ChangeRoomReply, error) {
	room, err := c.getChat(rid)
	if err != nil {
		return nil, err
	}
	if !room.Status {
		return nil, errors.ErrRoomClose
	}

	user, err := c.member.Get(uid)
	if err != nil {
		return nil, err
	}

	if user.IsBlockade {
		return nil, errors.ErrBlockade
	}
	connect := newPbConnect(user, room, "", int32(rid))
	connect.Status = true
	return &pb.ChangeRoomReply{
		HeaderMessage: room.HeaderMessage,
		Connect:       connect,
	}, nil
}

func (c *chat) get(id int) (models.Room, error) {
	room, err := c.cache.get(id)
	if err != nil {
		if err != redis.Nil {
			return models.Room{}, err
		}
		if room, err = c.reloadChat(id); err != nil {
			return models.Room{}, err
		}
	}
	return room, nil
}

func (c *chat) getChat(id int) (models.Room, error) {
	room, err := c.cache.getChat(id)
	if err != nil {
		if err != redis.Nil {
			return models.Room{}, err
		}
		if room, err = c.reloadChat(id); err != nil {
			return models.Room{}, err
		}
	}
	return room, nil
}

func (c *chat) reloadChat(id int) (models.Room, error) {
	room, msg, err := c.db.GetChat(id)
	if err != nil {
		if err == sql.ErrNoRows {
			return models.Room{}, errors.ErrNoRoom
		}
		return models.Room{}, err
	}

	if msg.RoomId == 0 {
		if err = c.cache.set(room); err != nil {
			return models.Room{}, err
		}
	} else {
		b, err := json.Marshal(message.RoomTopMessageToMessage(msg))
		if err != nil {
			log.Error("json Marshal for room top message", zap.Error(err), zap.Int("rid", id))
		}
		if err := c.cache.setChat(room, b); err != nil {
			return models.Room{}, err
		}
		room.HeaderMessage = b
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

func (c *chat) GetRoom(rid int) (models.Room, error) {
	room, err := c.get(rid)
	if err != nil {
		return room, err
	}
	if !room.Status {
		return room, errors.ErrRoomClose
	}
	if !room.IsMessage {
		return room, errors.ErrRoomNoMessage
	}
	return room, nil
}

func (c *chat) GetTopMessage(rid int) (message.Message, error) {
	msg, err := c.cache.getChatTopMessage(rid)
	if err != nil {
		if err == redis.Nil {
			return message.Message{}, errors.ErrNoRows
		}
		return message.Message{}, err
	}
	return message.ToMessage(msg)
}

func (c *chat) GetOnline(server string) (*Online, error) {
	return c.cache.getOnline(server)
}

func newPbConnect(user *models.Member, room models.Room, key string, roomId int32) *pb.Connect {
	connect := &pb.Connect{
		Uid:    user.Uid,
		Key:    key,
		Status: true,
		RoomID: roomId,
	}

	permission := new(pb.Permission)
	permissionMsg := new(pb.PermissionMessage)

	if user.Type != models.Guest {
		if !room.IsMessage {
			permissionMsg.IsMessage = "聊天室目前禁言状态，无法发言"
			permission.IsMessage = false
		} else if user.IsMessage {
			permission.IsMessage = true
		} else {
			permissionMsg.IsMessage = "您在永久禁言状态，无法发言"
		}

		permission.IsRedEnvelope = true
	} else {
		permissionMsg.IsMessage = "请先登入会员"
		permissionMsg.IsRedEnvelope = "请先登入会员"
	}
	connect.Permission = permission
	connect.PermissionMessage = permissionMsg
	return connect
}
