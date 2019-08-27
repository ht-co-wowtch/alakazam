package room

import (
	"database/sql"
	"github.com/go-redis/redis"
	"gitlab.com/jetfueltw/cpw/alakazam/client"
	"gitlab.com/jetfueltw/cpw/alakazam/errors"
	"gitlab.com/jetfueltw/cpw/alakazam/member"
	"gitlab.com/jetfueltw/cpw/alakazam/models"
)

type Room interface {
	Create(status Status) (int, error)
	Update(id int, status Status) error
	Delete(id int) error
	Get(id int) (models.Room, error)
	GetOnline(server string) (*Online, error)
	GetTopMessage(msgId int64) ([]int32, models.Message, error)
	DeleteTopMessage(msgId int64) error
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

	// 儲值&打碼量發話限制
	Limit Limit `json:"limit"`
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
	room := models.Room{
		IsMessage:    status.IsMessage,
		DayLimit:     status.Limit.Day,
		DepositLimit: status.Limit.Deposit,
		DmlLimit:     status.Limit.Dml,
		Status:       true,
	}
	if _, err := r.db.CreateRoom(&room); err != nil {
		return 0, err
	}
	return room.Id, r.c.set(room)
}

func (r *room) Update(id int, status Status) error {
	room := models.Room{
		Id:           id,
		IsMessage:    status.IsMessage,
		DayLimit:     status.Limit.Day,
		DepositLimit: status.Limit.Deposit,
		DmlLimit:     status.Limit.Dml,
	}
	_, err := r.db.UpdateRoom(room)
	if err != nil {
		return err
	}
	if err := r.c.set(room); err != nil {
		return err
	}
	return nil
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

func (c *room) GetOnline(server string) (*Online, error) {
	return c.c.getOnline(server)
}

func (c *room) GetTopMessage(msgId int64) ([]int32, models.Message, error) {
	roomTopMsg, err := c.db.FindRoomTopMessage(msgId)
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

func (c *room) DeleteTopMessage(msgId int64) error {
	err := c.db.DeleteRoomTopMessage(msgId)
	if err != nil {
		if err == sql.ErrNoRows {
			return errors.ErrNoRows
		}
		return err
	}
	return nil
}
