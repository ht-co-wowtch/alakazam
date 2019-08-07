package logic

import "gitlab.com/jetfueltw/cpw/alakazam/errors"

// TODO 需要踢人，如果沒有該會員？
func (l *Logic) SetBlockade(uid string) error {
	aff, err := l.db.SetBlockade(uid)
	if err != nil {
		return err
	}
	if aff <= 0 {
		return errors.ErrNoRows
	}
	return nil
}

// TODO 如果沒有該會員？
func (l *Logic) RemoveBlockade(uid string) (bool, error) {
	aff, err := l.db.DeleteBanned(uid)
	return aff >= 1, err
}
