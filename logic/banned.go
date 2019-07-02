package logic

import (
	log "github.com/golang/glog"
	"gitlab.com/jetfueltw/cpw/alakazam/models"
)

func (l *Logic) SetBanned(uid, remark string, expired int) error {
	return l.cache.SetBanned(uid, expired)
}

func (l *Logic) isUserBanned(uid string, status int) (bool, error) {
	if !models.IsBanned(status) {
		return false, nil
	}
	_, ok, err := l.cache.GetBanned(uid)
	if err != nil {
		return false, err
	}
	if ok {
		return true, nil
	}
	if err := l.cache.DelBanned(uid); err != nil {
		return true, err
	}
	return false, nil
}

func (l *Logic) RemoveBanned(uid string) error {
	if err := l.cache.DelBanned(uid); err != nil {
		log.Errorf("dao.DelBanned(uid: %s) error(%v)", uid, err)
		return err
	}
	return nil
}
