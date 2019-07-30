package logic

import (
	"gitlab.com/jetfueltw/cpw/alakazam/errors"
	"gitlab.com/jetfueltw/cpw/alakazam/models"
	"time"
)

func (l *Logic) SetBanned(uid, remark string, expired int) error {
	_, ok, err := l.db.Find(uid)
	if err != nil {
		return err
	}
	if !ok {
		return errors.ErrNoRows
	}
	return l.cache.SetBanned(uid, time.Duration(expired)*time.Second)
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
	return l.cache.DelBanned(uid)
}
