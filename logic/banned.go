package logic

import (
	log "github.com/golang/glog"
	"gitlab.com/jetfueltw/cpw/alakazam/logic/business"
)

func (l *Logic) SetBanned(uid, remark string, expired int) error {
	return l.cache.SetBanned(uid, expired)
}

func (l *Logic) isUserBanned(uid string, status int) bool {
	if !business.IsBanned(status) {
		return false
	}
	_, ok, err := l.cache.GetBanned(uid)
	if err != nil {
		log.Errorf("dao.GetBanned(uid: %s) error(%v)", uid, err)
		return false
	}
	if ok {
		return true
	}
	if err := l.cache.DelBanned(uid); err != nil {
		log.Errorf("dao.DelBanned(uid: %s) error(%v)", uid, err)
		return true
	}
	return false
}

func (l *Logic) RemoveBanned(uid string) error {
	if err := l.cache.DelBanned(uid); err != nil {
		log.Errorf("dao.DelBanned(uid: %s) error(%v)", uid, err)
		return err
	}
	return nil
}
