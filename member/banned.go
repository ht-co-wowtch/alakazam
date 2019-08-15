package member

import (
	"gitlab.com/jetfueltw/cpw/alakazam/errors"
	"gitlab.com/jetfueltw/cpw/alakazam/models"
	"time"
)

func (l *Member) SetBanned(uid string, expired int) error {
	_, ok, err := l.db.Find(uid)
	if err != nil {
		return err
	}
	if !ok {
		return errors.ErrNoRows
	}
	return l.c.SetBanned(uid, time.Duration(expired)*time.Second)
}

func (l *Member) IsUserBanned(uid string, status int) (bool, error) {
	if !models.IsBanned(status) {
		return false, nil
	}
	_, ok, err := l.c.GetBanned(uid)
	if err != nil {
		return false, err
	}
	if ok {
		return true, nil
	}
	if err := l.c.DelBanned(uid); err != nil {
		return true, err
	}
	return false, nil
}

func (l *Member) RemoveBanned(uid string) error {
	return l.c.DelBanned(uid)
}
