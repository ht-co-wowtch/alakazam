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
	"gitlab.com/jetfueltw/cpw/alakazam/message/scheme"
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
	GetMessageSession(uid string, rid int) (*models.Member, models.Room, error)
	ChangeRoom(uid string, rid int, key string) (*pb.ConnectReply, error)
	GetTopMessage(rid int) (scheme.Message, error)
	GetOnline(server string) (*Online, error)
	GetManages(rid int) ([]memberList, error)
	GetBlockades(rid int) ([]memberList, error)
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

	user, key, err := c.member.Login(room, params.Token, server)
	if err != nil {
		return nil, err
	}

	return c.newConnectReply(user, room, key)
}

func (c *chat) ChangeRoom(uid string, rid int, key string) (*pb.ConnectReply, error) {
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

	if err := c.member.ChangeRoom(uid, key, rid); err != nil {
		return nil, err
	}

	return c.newConnectReply(user, room, key)
}

func (c *chat) newConnectReply(user *models.Member, room models.Room, key string) (*pb.ConnectReply, error) {
	if user.Blockade() {
		return nil, errors.ErrBlockade
	}

	connect := NewPbConnect(user, room, key, int32(room.Id))
	connect.Status = true

	return &pb.ConnectReply{
		Heartbeat:       c.heartbeatNanosec,
		TopMessage:      room.TopMessage,
		BulletinMessage: room.BulletinMessage,
		Connect:         connect,
		User: &pb.User{
			Id:     user.Id,
			Uid:    user.Uid,
			Name:   user.Name,
			Gender: user.Gender,
			Type:   int32(user.Type),
		},
		IsConnectSuccessReply: true,
	}, nil
}

func (c *chat) GetMessageSession(uid string, rid int) (*models.Member, models.Room, error) {
	user, err := c.member.GetMessageSession(uid, rid)
	if err != nil {
		return nil, models.Room{}, err
	}

	room, err := c.GetRoom(rid)
	if err != nil {
		return nil, models.Room{}, err
	}
	return user, room, nil
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

	for _, m := range msg {
		switch m.Type {
		case models.TOP_MESSAGE:
			if m.Type == models.TOP_MESSAGE {
				room.TopMessage, err = json.Marshal(message.RoomTopMessageToMessage(m))
				if err != nil {
					log.Error("json Marshal for room top message", zap.Error(err), zap.Int("rid", id))
				}
			}
			break
		case models.BULLETIN_MESSAGE:
			room.BulletinMessage, err = json.Marshal(message.RoomBulletinMessageToMessage(m))
			if err != nil {
				log.Error("json Marshal for room bulletin message", zap.Error(err), zap.Int("rid", id))
			}
			break
		}
	}

	if err := c.cache.set(room); err != nil {
		return models.Room{}, err
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
	for room, count := range roomCount {
		r, err := c.cache.get(int(room))
		if err == nil {
			roomCount[room] = int32(r.AudienceRatio * float64(count))
		} else {
			roomCount[room] = count
		}
	}

	online := &Online{
		Server:    server,
		RoomCount: roomCount,
		Updated:   time.Now().Unix(),
	}

	err := c.cache.addOnline(server, online)
	if err != nil {
		return nil, err
	}

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

func (c *chat) GetTopMessage(rid int) (scheme.Message, error) {
	msg, err := c.cache.getChatTopMessage(rid)
	if err != nil {
		if err == redis.Nil {
			return scheme.Message{}, errors.ErrNoRows
		}
		return scheme.Message{}, err
	}
	return scheme.ToMessage(msg)
}

func (c *chat) GetOnline(server string) (*Online, error) {
	return c.cache.getOnline(server)
}

type memberList struct {
	Uid    string `json:"uid"`
	Name   string `json:"name"`
	Avatar string `json:"avatar"`
}

func (c *chat) GetManages(rid int) ([]memberList, error) {
	ms, err := c.db.GetManages(rid)
	if err != nil {
		return nil, err
	}

	d := []memberList{}

	for _, v := range ms {
		d = append(d, memberList{
			Uid:    v.Uid,
			Name:   v.Name,
			Avatar: scheme.ToAvatarName(v.Gender),
		})
	}

	return d, nil
}

func (c *chat) GetBlockades(rid int) ([]memberList, error) {
	ms, err := c.db.GetBlockades(rid)
	if err != nil {
		return nil, err
	}

	d := []memberList{}

	for _, v := range ms {
		d = append(d, memberList{
			Uid:    v.Uid,
			Name:   v.Name,
			Avatar: scheme.ToAvatarName(v.Gender),
		})
	}

	return d, nil
}

func NewPbConnect(user *models.Member, room models.Room, key string, roomId int32) *pb.Connect {
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
			permissionMsg.IsMessage = errors.RoomBanned
			permission.IsMessage = false
		} else if user.Banned() {
			permissionMsg.IsMessage = errors.MemberBanned
		} else {
			permission.IsMessage = true
		}

		if room.IsBets {
			permission.IsBets = true
		} else {
			permissionMsg.IsBets = errors.NoLoginMessage
		}

		if user.Permission.IsManage {
			permission.IsManage = true
		}

		permission.IsRedEnvelope = true
	} else {
		permissionMsg.IsMessage = errors.NoLoginMessage
		permissionMsg.IsRedEnvelope = errors.NoLoginMessage
		permissionMsg.IsBets = errors.NoLoginMessage
	}
	connect.Permission = permission
	connect.PermissionMessage = permissionMsg
	return connect
}
