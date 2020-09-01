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

	// 直播主
	STREAMER = 3

	// 房管
	MANAGE = 4
)

type Member struct {
	Id       int64     `xorm:"pk autoincr"`
	Uid      string    `json:"uid"`
	Name     string    `json:"name"`
	Type     int       `json:"type"`
	Gender   int32     `json:"gender"`
	CreateAt time.Time `json:"-"`

	RoomId     int  `json:"-" xorm:"-"`
	IsBanned   bool `json:"-" xorm:"-"`
	IsBlockade bool `json:"-" xorm:"-"`
	IsManage   bool `json:"-" xorm:"-"`
}

func (r *Member) TableName() string {
	return "members"
}

type Permission struct {
	RoomId     int64
	MemberId   int64
	IsBanned   bool
	IsBlockade bool
	IsManage   bool
}

func (r *Permission) TableName() string {
	return "room_user_permissions"
}

// 新增會員
func (s *Store) CreateUser(member *Member) (bool, error) {
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

func (s *Store) Permission(id int64, rid int) (Permission, error) {
	p := Permission{}
	_, err := s.d.Where("room_id = ?", rid).Where("member_id = ?", id).Get(&p)
	if err != nil {
		if err == sql.ErrNoRows {
			return Permission{}, nil
		}
		return Permission{}, err
	}

	return p, nil
}

func (s *Store) SetPermission(member Member) error {
	data := &Permission{
		IsBanned:   member.IsBanned,
		IsBlockade: member.IsBlockade,
		IsManage:   member.IsManage,
	}

	ok, err := s.d.Where("room_id = ?", member.RoomId).
		Where("member_id = ?", member.Id).
		Exist(data)
	if err != nil {
		return err
	}

	if ok {
		_, err = s.d.Cols("is_banned", "is_blockade", "is_manage").
			Where("room_id = ?", member.RoomId).
			Where("member_id = ?", member.Id).
			Update(data)
	} else {
		data.RoomId = int64(member.RoomId)
		data.MemberId = member.Id
		_, err = s.d.InsertOne(data)
	}

	return err
}

func (s *Store) GetMembers(ids []int64) ([]Member, error) {
	m := make([]Member, 0)
	err := s.d.Table(&Member{}).In("id", ids).Find(&m)
	return m, err
}

func (s *Store) GetMembersByUid(uid []string) ([]Member, error) {
	m := make([]Member, 0)
	err := s.d.In("uid", uid).Find(&m)
	return m, err
}
