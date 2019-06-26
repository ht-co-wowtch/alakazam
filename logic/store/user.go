package store

import "time"

const (
	// 訪客
	Guest = "guest"

	// 營銷
	Marketing = "marketing"

	// 玩家
	Player = 2
)

type Member struct {
	Id int `xorm:"pk autoincr"`

	Uid string `xorm:"char(32) not null unique index"`

	Name string `xorm:"varchar(30) not null"`

	Avatar string `xorm:"varchar(255) not null"`

	Permission int `xorm:"not null"`

	// 是否被封鎖
	IsBlockade bool `xorm:"default(0) not null"`

	// 建立時間
	CreateAt time.Time `xorm:"not null"`
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
