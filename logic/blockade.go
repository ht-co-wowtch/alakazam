package logic

import "gitlab.com/jetfueltw/cpw/alakazam/errors"

func (l *Logic) SetBlockade(uid, remark string) error {
	aff, err := l.db.SetBlockade(uid, remark)
	if err != nil {
		return err
	}
	if aff <= 0 {
		return errors.ErrNoRows
	}
	return nil
}

func (l *Logic) RemoveBlockade(uid string) (bool, error) {
	aff, err := l.db.DeleteBanned(uid)
	return aff >= 1, err
}
