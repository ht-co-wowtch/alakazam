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
	CreateAt time.Time
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

func (s *Store) GetTodaySystemBannedLog(memberId int) ([]BannedLog, error) {
	log := make([]BannedLog, 0)
	s.d.Where("`member_id` = ?", memberId).
		Where("is_system = ?", true).
		Limit(5).
		Desc("create_at").
		Find(&log)
	return log, nil
}
