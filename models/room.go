package models

import (
	"database/sql"
	"time"
)

const (
	LOTTERY_TYPE = 1
	LIVE_TYPE    = 2
)

type IChat interface {
	GetChat(id int) (Room, []RoomTopMessage, error)
}

type Room struct {
	// 要設定的房間id
	Id int `xorm:"pk autoincr"`

	// 是否禁言
	IsMessage bool

	// 是否可發/跟注
	IsBets bool

	// 聊天打碼與充值量天數限制
	DayLimit int

	// 充值量限制
	DepositLimit int

	// 打碼量限制
	DmlLimit int

	// 觀眾數倍率
	AudienceRatio float64

	// 房間狀態(開:1 關:0)
	Status bool

	// 更新時間
	UpdateAt time.Time `json:"-"`

	// 建立時間
	CreateAt time.Time `json:"-"`

	TopMessage []byte `xorm:"-"`

	BulletinMessage []byte `xorm:"-"`
}

func (r *Room) TableName() string {
	return "rooms"
}

func (s *Store) CreateRoom(room *Room) (int64, error) {
	room.UpdateAt = time.Now()
	room.CreateAt = time.Now()
	room.Status = true
	return s.d.Master().InsertOne(room)
}

func (s *Store) UpdateRoom(room Room) (int64, error) {
	return s.d.Cols("type", "member_id", "is_message", "is_bets", "day_limit", "deposit_limit", "dml_limit", "audience_ratio", "status", "update_at").
		Where("id = ?", room.Id).
		Update(&room)
}

func (s *Store) GetRoom(roomId int) (Room, error) {
	r := Room{}
	ok, err := s.d.Where("id = ?", roomId).Get(&r)
	if !ok {
		return r, sql.ErrNoRows
	}
	return r, err
}

func (s *Store) GetRoomTopMessage(id int) (RoomTopMessage, error) {
	top := RoomTopMessage{}
	ok, err := s.d.Where("`room_id` = ?", id).
		Where("status = ?", true).
		Get(&top)
	if err != nil {
		return top, err
	}
	if !ok {
		return top, sql.ErrNoRows
	}
	return top, nil
}

func (s *Store) GetChat(id int) (Room, []RoomTopMessage, error) {
	tx := s.d.Prepare()
	defer tx.Rollback()
	room := Room{}

	ok, _ := tx.Where("id = ?", id).
		Where("status = ?", true).
		Get(&room)

	top := make([]RoomTopMessage, 0)
	tx.Where("`room_id` = ?", id).Find(&top)

	if err := tx.Commit(); err != nil {
		return room, top, err
	}
	if !ok {
		return room, top, sql.ErrNoRows
	}
	return room, top, nil

}

func (s *Store) DeleteRoom(id int) (int64, error) {
	r := Room{
		Status: false,
	}
	return s.d.Cols("status").
		Where("id = ?", id).
		Update(&r)
}
