package models

import (
	"time"
)

type BannedLog struct {
	Id       int `xorm:"pk autoincr"`
	MemberId int
	Sec      int
	IsSystem bool
	ExpireAt time.Time
}

func (r *BannedLog) TableName() string {
	return "banned_logs"
}

func (s *Store) SetBannedLog(memberId int, sec time.Duration, isSystem bool) (int64, error) {
	return s.d.InsertOne(&BannedLog{
		MemberId: memberId,
		Sec:      int(sec.Seconds()),
		IsSystem: isSystem,
		ExpireAt: time.Now().Add(sec),
	})
}
