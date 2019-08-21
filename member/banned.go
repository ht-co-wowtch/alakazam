package member

import (
	"gitlab.com/jetfueltw/cpw/alakazam/errors"
	"gitlab.com/jetfueltw/cpw/micro/log"
	"go.uber.org/zap"
	"time"
)

func (m *Member) SetBanned(uid string, sec int) error {
	me, ok, err := m.db.Find(uid)
	if err != nil {
		return err
	}
	if !ok {
		return errors.ErrNoRows
	}
	expire := time.Duration(sec) * time.Second
	if sec > 0 {
		if err := m.c.setBanned(uid, expire); err != nil {
			return err
		}
	} else if sec == -1 {
		aff, err := m.db.UpdateIsMessage(me.Id, false)
		if err != nil {
			return err
		}
		if aff != 1 {
			return errors.ErrNoRows
		}
		expire = time.Duration(0)

		me.IsMessage = false
		if err := m.c.set(me); err != nil {
			return err
		}
	}

	aff, err := m.db.SetBannedLog(me.Id, expire, false)
	if err != nil || aff == 0 {
		log.Error("set banned log", zap.Error(err), zap.Int64("affected", aff))
	}
	return nil
}

func (m *Member) IsBanned(uid string) (bool, error) {
	ok, err := m.c.isBanned(uid)
	if err != nil {
		return false, err
	}
	if ok {
		return true, nil
	}
	if err := m.c.delBanned(uid); err != nil {
		return true, err
	}
	return false, nil
}

func (m *Member) RemoveBanned(uid string) error {
	return m.c.delBanned(uid)
}
