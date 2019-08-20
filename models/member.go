package models

import "time"

const (
	// 封鎖
	Blockade = 0

	// 聊天
	MessageStatus = 1

	// 搶紅包
	redEnvelope = 2

	PlayStatus = MessageStatus + redEnvelope
)

// 是否禁言
func IsBanned(status int) bool {
	return (MessageStatus & status) != MessageStatus
}

// 是否可搶/發紅包
func IsRedEnvelope(status int) bool {
	return (redEnvelope & status) == redEnvelope
}

const (
	// 訪客
	Guest = 0

	// 營銷
	Marketing = 1

	// 玩家
	Player = 2
)

type Member struct {
	Id int `xorm:"pk autoincr"`

	Uid string `xorm:"char(32) not null unique index"`

	Name string `xorm:"varchar(30) not null"`

	Avatar string `xorm:"varchar(255) not null"`

	Type int `xorm:"tinyint(3) not null"`

	// 是否被封鎖
	IsBlockade bool `xorm:"default(0) not null"`

	// 建立時間
	CreateAt time.Time `xorm:"not null"`
}

func (r *Member) Status() int {
	if r.IsBlockade {
		return Blockade
	}
	var status int
	switch r.Type {
	case Guest:
		status = MessageStatus
	case Player, Marketing:
		status = PlayStatus
	}
	return status
}

func (r *Member) TableName() string {
	return "members"
}

// 新增會員
func (s *Store) CreateUser(member *Member) (int64, error) {
	member.CreateAt = time.Now()
	return s.d.InsertOne(member)
}

func (s *Store) UpdateUser(member *Member) (int64, error) {
	return s.d.Cols("name", "avatar").
		Where("uid = ?", member.Uid).
		Update(member)
}

// 找會員
func (s *Store) Find(uid string) (*Member, bool, error) {
	m := new(Member)
	ok, err := s.d.Where("uid = ?", uid).
		Get(m)
	return m, ok, err
}

func (s *Store) GetMembers(ids []int) ([]Member, error) {
	m := make([]Member, 0)
	err := s.d.Table(&Member{}).Find(&m)
	return m, err
}
