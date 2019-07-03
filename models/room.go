package models

import "time"

const (
	money = 4
)

// 是否有充值&打碼量限制
func IsMoney(status int) bool {
	return (money & status) == money
}

const (
	RoomStatus = Message
)

type Room struct {
	// 要設定的房間id
	Id string `xorm:"pk"`

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

	// 更新時間
	UpdateAt time.Time `xorm:"not null"`

	// 建立時間
	CreateAt time.Time `xorm:"not null"`
}

func (r *Room) Status() int {
	if r.Id == "" {
		return RoomStatus
	}
	var status int
	if r.IsMessage {
		status += Message
	}
	if r.DayLimit > 0 && r.DmlLimit+r.DepositLimit > 0 {
		status += money
	}
	return status
}

func (r *Room) TableName() string {
	return "rooms"
}

func (s *Store) CreateRoom(room Room) (int64, error) {
	room.UpdateAt = time.Now()
	room.CreateAt = time.Now()
	return s.d.InsertOne(&room)
}

func (s *Store) UpdateRoom(room Room) (int64, error) {
	return s.d.Cols("is_message", "is_follow", "day_limit", "deposit_limit", "dml_limit").
		Where("id = ?", room.Id).
		Update(&room)
}

func (s *Store) GetRoom(roomId string) (Room, bool, error) {
	r := Room{}
	ok, err := s.d.Where("id = ?", roomId).Get(&r)
	return r, ok, err
}
