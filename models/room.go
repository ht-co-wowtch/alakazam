package models

import (
	"github.com/go-xorm/xorm"
	"time"
)

const (
	money = 4
)

// 是否有充值&打碼量限制
func IsMoney(status int) bool {
	return (money & status) == money
}

type Room struct {
	// 要設定的房間id
	Id int `xorm:"pk autoincr"`

	Uuid string

	// 是否禁言
	IsMessage bool `xorm:"default(0) not null"`

	// 是否可發/跟注
	IsFollow bool `xorm:"default(0) not null"`

	// 聊天打碼與充值量天數限制
	DayLimit int `xorm:"tinyint(4) default(0)"`

	// 充值量限制
	DepositLimit int `xorm:"default(0)"`

	// 打碼量限制
	DmlLimit int `xorm:"default(0)"`

	// 房間狀態(開:1 關:0)
	Status bool `xorm:"default(1)"`

	// 更新時間
	UpdateAt time.Time `xorm:"not null"`

	// 建立時間
	CreateAt time.Time `xorm:"not null"`
}

func (r *Room) Permission() int {
	if r.Uuid == "" {
		return Message
	}
	var permission int
	if r.IsMessage {
		permission += Message
	}
	if r.DayLimit > 0 && r.DmlLimit+r.DepositLimit > 0 {
		permission += money
	}
	return permission
}

func (r *Room) TableName() string {
	return "rooms"
}

func (s *Store) CreateRoom(room Room) (int64, error) {
	room.UpdateAt = time.Now()
	room.CreateAt = time.Now()
	room.Status = true
	tx := s.d.Master().Prepare()
	defer tx.Rollback()

	aff, err := tx.InsertOne(&room)
	if err != nil || aff != 1 {
		return aff, err
	}

	aff, err = tx.InsertOne(&Seq{
		RoomId: room.Id,
		Batch:  200,
	})
	if err != nil || aff != 1 {
		return aff, err
	}
	return 1, tx.Commit()
}

func (s *Store) UpdateRoom(room Room) (int64, error) {
	var u *xorm.Session
	if room.Status {
		u = s.d.Cols("is_message", "day_limit", "deposit_limit", "dml_limit", "status")
	} else {
		u = s.d.Cols("is_message", "day_limit", "deposit_limit", "dml_limit")
	}
	return u.Where("uuid = ?", room.Uuid).
		Update(&room)
}

func (s *Store) GetRoom(roomId string) (Room, bool, error) {
	r := Room{}
	ok, err := s.d.Where("id = ?", roomId).Get(&r)
	return r, ok, err
}

func (s *Store) DeleteRoom(roomId string) (int64, error) {
	r := Room{
		Status: false,
	}
	return s.d.Cols("status").
		Where("id = ?", roomId).
		Update(&r)
}
