package logic

import (
	log "github.com/golang/glog"
)

func (l *Logic) SetBlockade(uid, remark string) bool {
	if aff, err := l.db.SetBlockade(uid, remark); err != nil || aff <= 0 {
		log.Errorf("logic.db.SetBlockade uid:%s aff:%d error(%v)", uid, aff, err)
		return false
	}
	return true
}

func (l *Logic) RemoveBlockade(uid string) bool {
	if aff, err := l.db.DeleteBanned(uid); err != nil || aff <= 0 {
		log.Errorf("logic.db.DeleteBanned uid:%s aff:%d error(%v)", uid, aff, err)
		return false
	}
	return true
}
