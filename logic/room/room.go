package room

import (
	"encoding/json"
	"fmt"
	"gitlab.com/jetfueltw/cpw/alakazam/client"
	"gitlab.com/jetfueltw/cpw/alakazam/errors"
	"gitlab.com/jetfueltw/cpw/alakazam/logic/cache"
	"gitlab.com/jetfueltw/cpw/alakazam/logic/member"
	"gitlab.com/jetfueltw/cpw/alakazam/models"
	"gitlab.com/jetfueltw/cpw/micro/errdefs"
	"gopkg.in/go-playground/validator.v8"
	"time"
)

type Room struct {
	db        *models.Store
	cache     *cache.Cache
	member    *member.Member
	cli       *client.Client
	heartbeat int64
}

func New(db *models.Store, cache *cache.Cache, member *member.Member, cli *client.Client, heartbeat int64) *Room {
	return &Room{
		db:        db,
		cache:     cache,
		member:    member,
		cli:       cli,
		heartbeat: heartbeat,
	}
}

type Status struct {
	// 要設定的房間id
	Id string `json:"id" binding:"required,len=32"`

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

func (l *Room) CreateRoom(r Status) (string, error) {
	room := models.Room{
		Id:           r.Id,
		IsMessage:    r.IsMessage,
		DayLimit:     r.Limit.Day,
		DepositLimit: r.Limit.Deposit,
		DmlLimit:     r.Limit.Dml,
		Status:       true,
	}
	dbRoom, err := l.GetRoom(r.Id)
	if err == errors.ErrNoRows {
		_, err = l.db.CreateRoom(room)
	}
	if err != nil {
		return "", err
	}
	if dbRoom.Id != "" {
		return room.Id, l.updateRoom(room)
	}
	return r.Id, l.cache.SetRoom(room)
}

func (l *Room) UpdateRoom(r Status) error {
	room := models.Room{
		Id:           r.Id,
		IsMessage:    r.IsMessage,
		DayLimit:     r.Limit.Day,
		DepositLimit: r.Limit.Deposit,
		DmlLimit:     r.Limit.Dml,
	}
	return l.updateRoom(room)
}

func (l *Room) DeleteRoom(roomId string) error {
	r, err := l.GetRoom(roomId)
	if err != nil {
		return err
	}
	if r.Status == false {
		return nil
	}
	aff, err := l.db.DeleteRoom(roomId)
	if err != nil {
		return err
	}
	if aff <= 0 {
		return errors.ErrNoRows
	}
	return nil
}

func (l *Room) GetRoom(roomId string) (models.Room, error) {
	r, ok, err := l.db.GetRoom(roomId)
	if err != nil {
		return models.Room{}, err
	}
	if !ok {
		return models.Room{}, errors.ErrNoRows
	}
	return r, nil
}

func (l *Room) updateRoom(room models.Room) error {
	_, err := l.db.UpdateRoom(room)
	if err != nil {
		return err
	}
	if err := l.cache.SetRoom(room); err != nil {
		return err
	}
	return nil
}

func (l *Room) IsMessage(rid string, status int, uid, token string) error {
	if !models.IsMoney(status) {
		return nil
	}
	day, dml, deposit, err := l.cache.GetRoomByMoney(rid)
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
	return l.cache.DeleteUser(uid, key)
}

// user key更換房間
func (l *Room) ChangeRoom(uid, key, roomId string) error {
	return l.cache.ChangeRoom(uid, key, roomId)
}

// 更新某人redis資訊的過期時間
func (l *Room) Heartbeat(uid, key, roomId, name, server string) error {
	_, err := l.cache.RefreshUserExpire(uid)
	if err != nil {
		return err
	}
	return nil
}

// restart redis內存的每個房間總人數
func (l *Room) RenewOnline(server string, roomCount map[string]int32) (map[string]int32, error) {
	online := &cache.Online{
		Server:    server,
		RoomCount: roomCount,
		Updated:   time.Now().Unix(),
	}
	if err := l.cache.AddServerOnline(server, online); err != nil {
		return nil, err
	}
	return roomCount, nil
}
