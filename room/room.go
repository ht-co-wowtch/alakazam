package room

import (
	"database/sql"
	"encoding/json"
	"github.com/go-redis/redis"
	"github.com/go-sql-driver/mysql"
	"gitlab.com/jetfueltw/cpw/alakazam/client"
	"gitlab.com/jetfueltw/cpw/alakazam/errors"
	"gitlab.com/jetfueltw/cpw/alakazam/member"
	"gitlab.com/jetfueltw/cpw/alakazam/message"
	"gitlab.com/jetfueltw/cpw/alakazam/models"
	"time"
)

const (
	// 彩票房間
	LOTTERY_TYPE = "lottery"

	// 直播房間
	LIVE_TYPE    = "live"
)

type Room interface {
	Create(status Status) (int, error)
	Update(id int, status Status) error
	Delete(id int) error
	Get(id int) (models.Room, error)
	GetTopMessage(msgId int64) ([]int32, models.Message, error)
	AddTopMessage(rids []int32, message message.Message) error
	DeleteTopMessage(rids []int32, msgId int64) error
}

type room struct {
	db     *models.Store
	c      Cache
	member *member.Member
	cli    *client.Client
}

func New(db *models.Store, cache *redis.Client, member *member.Member, cli *client.Client) Room {
	return &room{
		db:     db,
		c:      newCache(cache),
		member: member,
		cli:    cli,
	}
}

type Status struct {
	// 是否禁言
	IsMessage bool `json:"is_message"`

	// 是否打開跟注
	IsBets bool `json:"is_bets"`

	// 儲值&打碼量發話限制
	Limit Limit `json:"limit"`

	// 房間種類
	Type string `json:"type"`

	// 房間屬於誰
	Uid string `json:"uid"`

	// 房間狀態
	Status bool `json:"status"`
}

type Limit struct {
	// 限制範圍
	Day int `json:"day" binding:"max=31"`

	// 儲值金額
	Deposit int `json:"deposit"`

	// 打碼量
	Dml int `json:"dml"`
}

func (r *room) Create(status Status) (int, error) {
	model, err := r.newRoomModel(status)
	if err != nil {
		return 0, err
	}

	model.Status = true
	if _, err := r.db.CreateRoom(&model); err != nil {
		return 0, err
	}
	return model.Id, r.c.set(model)
}

func (r *room) Update(id int, status Status) error {
	model, err := r.newRoomModel(status)
	if err != nil {
		return err
	}

	model.Id = id
	model.Status = status.Status
	if _, err = r.db.UpdateRoom(model); err != nil {
		return err
	}
	if err = r.c.set(model); err != nil {
		return err
	}
	return nil
}

func (r *room) newRoomModel(status Status) (models.Room, error) {
	room := models.Room{
		IsMessage:    status.IsMessage,
		IsBets:       status.IsBets,
		DayLimit:     status.Limit.Day,
		DepositLimit: status.Limit.Deposit,
		DmlLimit:     status.Limit.Dml,
	}

	switch status.Type {
	case LOTTERY_TYPE:
		room.Type = models.LOTTERY_TYPE
		room.MemberId = sql.NullInt64{Valid: false}
	case LIVE_TYPE:
		room.Type = models.LIVE_TYPE

		m, err := r.member.Fetch(status.Uid)
		if err != nil {
			return models.Room{}, err
		}

		room.MemberId = sql.NullInt64{
			Int64: int64(m.Id),
			Valid: true,
		}
	default:
		return models.Room{}, errors.ErrRoomType
	}
	return room, nil
}

func (r *room) Delete(id int) error {
	room, err := r.Get(id)
	if err != nil {
		return err
	}
	if room.Status == false {
		return nil
	}
	aff, err := r.db.DeleteRoom(id)
	if err != nil {
		return err
	}
	if aff <= 0 {
		return errors.ErrNoRows
	}

	room.Status = false
	if err := r.c.set(room); err != nil {
		return err
	}
	return nil
}

func (r *room) Get(id int) (models.Room, error) {
	room, err := r.db.GetRoom(id)
	if err != nil {
		if err == sql.ErrNoRows {
			return models.Room{}, errors.ErrNoRows
		}
		return models.Room{}, err
	}
	return room, nil
}

func (r *room) GetTopMessage(msgId int64) ([]int32, models.Message, error) {
	roomTopMsg, err := r.db.FindRoomTopMessage(msgId)
	if err != nil {
		return nil, models.Message{}, err
	}
	if len(roomTopMsg) == 0 {
		return nil, models.Message{}, errors.ErrNoRows
	}

	rid := make([]int32, 0, len(roomTopMsg))
	for _, v := range roomTopMsg {
		rid = append(rid, v.RoomId)
	}
	return rid, models.Message{
		MsgId:   msgId,
		Message: roomTopMsg[0].Message,
		SendAt:  roomTopMsg[0].SendAt,
	}, nil
}

func (r *room) AddTopMessage(rids []int32, msg message.Message) error {
	model := models.Message{
		MsgId:   msg.Id,
		Message: msg.Message,
		SendAt:  time.Now(),
	}
	if err := r.db.AddRoomTopMessage(rids, model); err != nil {
		if e, ok := err.(*mysql.MySQLError); ok {
			if e.Number == 1452 {
				return errors.ErrNoRoom
			}
		}
		return err
	}

	b, err := json.Marshal(msg)
	if err != nil {
		return err
	}
	return r.c.setChatTopMessage(rids, b)
}

func (r *room) DeleteTopMessage(rids []int32, msgId int64) error {
	err := r.db.DeleteRoomTopMessage(msgId)
	if err != nil {
		if err == sql.ErrNoRows {
			return errors.ErrNoRows
		}
		return err
	}
	return r.c.deleteChatTopMessage(rids)
}
