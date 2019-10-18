package models

import (
	"database/sql"
	"time"
)

const (
	// 封鎖
	Blockade = 0

	// 聊天
	MessageStatus = 1

	// 搶紅包
	redEnvelope = 2

	PlayStatus = MessageStatus + redEnvelope
)

const (
	// 訪客
	Guest = 0

	// 營銷
	Market = 1

	// 玩家
	Player = 2
)

type Member struct {
	Id         int       `xorm:"pk autoincr"`
	Uid        string    `json:"uid"`
	Name       string    `json:"name"`
	Type       int       `json:"type"`
	Gender     int       `json:"gender"`
	IsMessage  bool      `json:"is_message"`
	IsBlockade bool      `json:"is_blockade"`
	CreateAt   time.Time `json:"-"`
}

func (r *Member) TableName() string {
	return "members"
}

// 新增會員
func (s *Store) CreateUser(member *Member) (bool, error) {
	member.IsMessage = true
	member.CreateAt = time.Now()
	aff, err := s.d.InsertOne(member)
	return aff == 1, err
}

func (s *Store) UpdateUser(member *Member) (bool, error) {
	aff, err := s.d.Cols("name", "gender").
		Where("uid = ?", member.Uid).
		Update(member)
	return aff == 1, err
}

// 找會員
func (s *Store) Find(uid string) (*Member, error) {
	m := new(Member)
	ok, err := s.d.Where("uid = ?", uid).
		Get(m)
	if !ok {
		return nil, sql.ErrNoRows
	}
	return m, err
}

func (s *Store) GetMembers(ids []int) ([]Member, error) {
	m := make([]Member, 0)
	err := s.d.Table(&Member{}).Find(&m)
	return m, err
}

func (s *Store) GetMembersByUid(uid []string) ([]Member, error) {
	m := make([]Member, 0)
	err := s.d.In("uid", uid).Find(&m)
	return m, err
}

func (s *Store) UpdateIsMessage(memberId int, isMessage bool) (bool, error) {
	aff, err := s.d.Cols("is_message").
		Where("id = ?", memberId).
		Update(&Member{
			IsMessage: isMessage,
		})
	return aff == 1, err
}
