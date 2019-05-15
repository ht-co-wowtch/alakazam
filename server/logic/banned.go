package logic

import "time"

func (l *Logic) Banned(uid, remark string, expired int) error {
	return l.dao.SetBanned(uid, time.Duration(expired)*time.Second)
}
