package logic

import (
	log "github.com/golang/glog"
	"gitlab.com/jetfueltw/cpw/alakazam/server/logic/business"
)

func (l *Logic) SetBanned(uid, remark string, expired int) error {
	return l.dao.SetBanned(uid, expired)
}

func (l *Logic) IsBanned(uid string, status int) bool {
	if !business.IsBanned(status) {
		return false
	}
	_, ok, err := l.dao.GetBanned(uid)
	if err != nil {
		log.Errorf("dao.GetBanned(uid: %s) error(%v)", uid, err)
		return false
	}
	if ok {
		return true
	}
	if err := l.dao.DelBanned(uid); err != nil {
		log.Errorf("dao.DelBanned(uid: %s) error(%v)", uid, err)
		return true
	}
	return false
}

func (l *Logic) RemoveBanned(uid string) bool {
	// TODO 待實作 從redis刪除禁言資料並對此user增加發話權限
	// 用DelBanned method
	return true
}
