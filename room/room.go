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

func (l *Room) IsMessage(rid int, status int, uid, token string) error {
	if !models.IsMoney(status) {
		return nil
	}
	day, dml, deposit, err := l.c.getMoney(rid)
	if err != nil {
		return err
	}
	money, err := l.cli.GetDepositAndDml(day, uid, token)
	if err != nil {
		return err
	}
	if float64(dml) > money.Dml || float64(deposit) > money.Deposit {
		e := errors.New(fmt.Sprintf("您无法发言，当前发言条件：前%d天充值不少于%d元；打码量不少于%d元", day, deposit, dml))
		return errdefs.Unauthorized(e, 4)
	}
	return nil
}

type ConnectReply struct {
	// user uid
	Uid string

	// websocket connection key
	Key string

	// user name
	Name string

	// key所在的房間id
	RoomId string

	// 前台心跳週期時間
	Hb int64

	// 操作權限
	Permission int
}

var v *validator.Validate

func init() {
	v = validator.New(&validator.Config{TagName: "binding"})
}

// redis紀錄某人連線資訊
func (l *Room) Connect(server string, token []byte) (*ConnectReply, error) {
	var params struct {
		// 帳務中心+版的認證token
		Token string `json:"token" binding:"required"`

		// client要進入的room
		RoomID string `json:"room_id" binding:"required"`
	}

	if err := json.Unmarshal(token, &params); err != nil {
		return nil, err
	}
	if err := v.Struct(&params); err != nil {
		return nil, err
	}

	rid, err := strconv.Atoi(params.RoomID)
	if err != nil {
		return nil, err
	}
	if _, err := l.Get(rid); err != nil {
		return nil, err
	}

	r := new(ConnectReply)
	user, key, err := l.member.Login(params.Token, params.RoomID, server)
	if err != nil {
		return nil, err
	}
	r.Uid = user.Uid
	r.Name = user.Name
	r.Permission = user.Status()
	r.RoomId = params.RoomID
	r.Key = key
	// 告知comet連線多久沒心跳就直接close
	r.Hb = l.heartbeat
	return r, nil
}

// redis清除某人連線資訊
func (l *Room) Disconnect(uid, key, server string) (has bool, err error) {
	return l.member.Logout(uid, key)
}

// user key更換房間
func (l *Room) ChangeRoom(uid, key, roomId string) error {
	return l.member.ChangeRoom(uid, key, roomId)
}

// 更新某人redis資訊的過期時間
func (l *Room) Heartbeat(uid, key, roomId, name, server string) error {
	return l.member.Heartbeat(uid)
}

// restart redis內存的每個房間總人數
func (l *Room) RenewOnline(server string, roomCount map[string]int32) (map[string]int32, error) {
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
