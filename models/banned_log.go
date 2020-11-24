package models

import (
	"time"
)

type BannedLog struct {
	Id       int `xorm:"pk autoincr"`
	MemberId int64
	Sec      int
	IsSystem bool
	ExpireAt time.Time
	CreateAt time.Time
}

func (r *BannedLog) TableName() string {
	return "banned_logs"
}

func (s *Store) SetBannedLog(mid int64, sec time.Duration, isSystem bool) (bool, error) {
	aff, err := s.d.InsertOne(&BannedLog{
		MemberId: mid,
		Sec:      int(sec.Seconds()),
		IsSystem: isSystem,
		ExpireAt: time.Now().Add(sec),
		CreateAt: time.Now(),
	})
	return aff == 1, err
}

func (s *Store) GetTodaySystemBannedLog(mid int64) ([]BannedLog, error) {
	log := make([]BannedLog, 0)
	err := s.d.Where("`member_id` = ?", mid).
		Where("is_system = ?", true).
		Limit(5).
		Desc("create_at").
		Find(&log)
	return log, err
}
