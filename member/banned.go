package member

import (
	"gitlab.com/jetfueltw/cpw/alakazam/errors"
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
	return l.c.setBanned(uid, time.Duration(expired)*time.Second)
}

func (l *Member) IsBanned(uid string) (bool, error) {
	ok, err := l.c.isBanned(uid)
	if err != nil {
		return false, err
	}
	if ok {
		return true, nil
	}
	if err := l.c.delBanned(uid); err != nil {
		return true, err
	}
	return false, nil
}

func (l *Member) RemoveBanned(uid string) error {
	return l.c.delBanned(uid)
}
