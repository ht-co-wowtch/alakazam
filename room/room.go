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
	"time"
)

type Room struct {
	db        *models.Store
	c         *Cache
	member    *member.Member
	cli       *client.Client
	heartbeat int64
}

func New(db *models.Store, cache *redis.Client, member *member.Member, cli *client.Client, heartbeat int64) *Room {
	return &Room{
		db:        db,
		c:         newCache(cache),
		member:    member,
		cli:       cli,
		heartbeat: heartbeat,
	}
}

type ConnectReply struct {
	// user uid
	Uid string

	// websocket connection key
	Key string

	// user name
	Name string

	// key所在的房間id
	RoomId int

	// 前台心跳週期時間
	Hb int64

	// 操作權限
	Permission int
}

var v *validator.Validate

func init() {
	v = validator.New(&validator.Config{TagName: "binding"})
}

func (r *Room) Connect(server string, token []byte) (*ConnectReply, error) {
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
	if _, err := r.Get(params.RoomID); err != nil {
		return nil, err
	}

	connectReply := new(ConnectReply)
	user, key, err := r.member.Login(params.RoomID, params.Token, server)
	if err != nil {
		return nil, err
	}

	connectReply.Uid = user.Uid
	connectReply.Name = user.Name
	connectReply.Permission = user.Status()
	connectReply.RoomId = params.RoomID
	connectReply.Key = key
	// 告知comet連線多久沒心跳就直接close
	connectReply.Hb = r.heartbeat
	return connectReply, nil
}

// redis清除某人連線資訊
func (r *Room) Disconnect(uid, key string) (has bool, err error) {
	return r.member.Logout(uid, key)
}

// 更新某人redis資訊的過期時間
func (l *Room) Heartbeat(uid, key, name, server string) error {
	return l.member.Heartbeat(uid)
}

// restart redis內存的每個房間總人數
func (l *Room) RenewOnline(server string, roomCount map[int32]int32) (map[int32]int32, error) {
	online := &Online{
		Server:    server,
		RoomCount: roomCount,
		Updated:   time.Now().Unix(),
	}
	if err := l.c.addOnline(server, online); err != nil {
		return nil, err
	}
	return roomCount, nil
}

func (r *Room) GetOnline(server string) (*Online, error) {
	return r.c.getOnline(server)
}

type Status struct {
	// 是否禁言
	IsMessage bool `json:"is_message"`

	// 儲值&打碼量發話限制
	Limit Limit `json:"limit"`

	// 房間狀態
	status bool
}

type Limit struct {
	// 限制範圍
	Day int `json:"day" binding:"max=31"`

	// 儲值金額
	Deposit int `json:"deposit"`

	// 打碼量
	Dml int `json:"dml"`
}

func (l *Room) Create(r Status) (int, error) {
	room := models.Room{
		IsMessage:    r.IsMessage,
		DayLimit:     r.Limit.Day,
		DepositLimit: r.Limit.Deposit,
		DmlLimit:     r.Limit.Dml,
		Status:       true,
	}
	if _, err := l.db.CreateRoom(&room); err != nil {
		return 0, err
	}
	return room.Id, l.c.set(room)
}

func (l *Room) Update(id int, r Status) error {
	room := models.Room{
		Id:           id,
		IsMessage:    r.IsMessage,
		DayLimit:     r.Limit.Day,
		DepositLimit: r.Limit.Deposit,
		DmlLimit:     r.Limit.Dml,
	}
	return l.update(room)
}

func (l *Room) Delete(id int) error {
	r, err := l.Get(id)
	if err != nil {
		return err
	}
	if r.Status == false {
		return nil
	}
	aff, err := l.db.DeleteRoom(id)
	if err != nil {
		return err
	}
	if aff <= 0 {
		return errors.ErrNoRows
	}
	return nil
}

func (l *Room) Get(id int) (models.Room, error) {
	r, ok, err := l.db.GetRoom(id)
	if err != nil {
		return models.Room{}, err
	}
	if !ok {
		return models.Room{}, errors.ErrNoRows
	}
	return r, nil
}

func (l *Room) update(room models.Room) error {
	_, err := l.db.UpdateRoom(room)
	if err != nil {
		return err
	}
	if err := l.c.set(room); err != nil {
		return err
	}
	return nil
}

func (r *Room) IsMessage(rid int, uid string) error {
	room, err := r.Get(rid)
	if err != nil {
		return err
	}
	// TODO 三方需改不需要token
	money, err := r.cli.GetDepositAndDml(room.DayLimit, uid, "token")
	if err != nil {
		return err
	}
	if float64(room.DmlLimit) > money.Dml || float64(room.DepositLimit) > money.Deposit {
		e := errors.New(fmt.Sprintf("您无法发言，当前发言条件：前%d天充值不少于%d元；打码量不少于%d元", room.DayLimit, room.DepositLimit, room.DmlLimit))
		return errdefs.Unauthorized(e, 4)
	}
	return nil
}
