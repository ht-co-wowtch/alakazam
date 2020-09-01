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
	"gitlab.com/jetfueltw/cpw/alakazam/message/scheme"
	"gitlab.com/jetfueltw/cpw/alakazam/models"
	"time"
)

type Room interface {
	Create(status Status) (int, error)
	Update(id int, status Status) error
	Delete(id int) error
	Get(id int) (models.Room, error)
	GetTopMessage(msgId int64, t int) ([]int32, models.Message, error)
	AddTopMessage(rids []int32, seq int64, message string, ts []int) error
	DeleteTopMessage(rids []int32, msgId int64, t int) error
	Online() (map[int32]int32, error)
}

type room struct {
	db     *models.Store
	c      Cache
	member *member.Member
	cli    *client.Client
}

func New(db *models.Store, cache *redis.Client, m *member.Member, cli *client.Client) Room {
	return &room{
		db:     db,
		c:      newCache(cache),
		member: m,
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

	// 房間屬於誰
	Uid string `json:"uid"`

	// 觀眾數倍率
	AudienceRatio float64 `json:"audience_ratio"`

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
	model := r.newRoomModel(status)
	model.Status = true

	if _, err := r.db.CreateRoom(&model); err != nil {
		return 0, err
	}
	return model.Id, r.c.set(model)
}

func (r *room) Update(id int, status Status) error {
	model := r.newRoomModel(status)
	model.Id = id
	model.Status = status.Status

	if _, err := r.db.UpdateRoom(model); err != nil {
		return err
	}
	if err := r.c.set(model); err != nil {
		return err
	}
	return nil
}

func (r *room) newRoomModel(status Status) models.Room {
	if status.AudienceRatio < 1 {
		status.AudienceRatio = 1
	}

	return models.Room{
		IsMessage:     status.IsMessage,
		IsBets:        status.IsBets,
		DayLimit:      status.Limit.Day,
		DepositLimit:  status.Limit.Deposit,
		DmlLimit:      status.Limit.Dml,
		AudienceRatio: status.AudienceRatio,
	}
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

func (r *room) GetTopMessage(msgId int64, t int) ([]int32, models.Message, error) {
	roomTopMsg, err := r.db.FindRoomTopMessage(msgId)
	if err != nil {
		return nil, models.Message{}, err
	}
	if len(roomTopMsg) == 0 {
		return nil, models.Message{}, errors.ErrNoRows
	}

	var index int
	rid := make([]int32, 0, len(roomTopMsg))
	for i, v := range roomTopMsg {
		if v.Type == t {
			index = i
			rid = append(rid, v.RoomId)
		}
	}

	return rid, models.Message{
		MsgId:   msgId,
		Message: roomTopMsg[index].Message,
		SendAt:  roomTopMsg[index].SendAt,
	}, nil
}

func (r *room) AddTopMessage(rids []int32, seq int64, msg string, ts []int) error {
	var roomTopMessage scheme.Message
	model := models.RoomTopMessage{
		MsgId:   seq,
		Message: msg,
		SendAt:  time.Now(),
	}

	for _, t := range ts {
		model.Type = t
		if err := r.db.AddRoomTopMessage(rids, model); err != nil {
			if e, ok := err.(*mysql.MySQLError); ok {
				if e.Number == 1452 {
					return errors.ErrNoRoom
				}
			}
			return err
		}

		if t == models.TOP_MESSAGE {
			roomTopMessage = message.RoomTopMessageToMessage(model)
		} else {
			roomTopMessage = message.RoomBulletinMessageToMessage(model)
		}

		b, err := json.Marshal(roomTopMessage)
		if err != nil {
			return err
		}

		if t == models.TOP_MESSAGE {
			err = r.c.setChatTopMessage(rids, b)
		} else {
			err = r.c.setChatBulletinMessage(rids, b)
		}

		if err != nil {
			return err
		}
	}

	return nil
}

func (r *room) DeleteTopMessage(rids []int32, msgId int64, t int) error {
	err := r.db.DeleteRoomTopMessage(rids, msgId, t)
	if err != nil {
		if err == sql.ErrNoRows {
			return errors.ErrNoRows
		}
		return err
	}

	if t == models.TOP_MESSAGE {
		return r.c.deleteChatTopMessage(rids)
	}

	return r.c.deleteChatBulletinMessage(rids)
}

func (r *room) Online() (map[int32]int32, error) {
	online, err := r.c.getOnline("hostname")
	if err != nil {
		return nil, err
	}
	return online.RoomCount, nil
}
