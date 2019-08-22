package member

import (
	"gitlab.com/jetfueltw/cpw/alakazam/errors"
	"gitlab.com/jetfueltw/cpw/micro/log"
	"go.uber.org/zap"
	"time"
)

func (m *Member) SetBanned(uid string, sec int, isSystem bool) error {
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

	aff, err := m.db.SetBannedLog(me.Id, expire, isSystem)
	if err != nil || aff == 0 {
		log.Error("set banned log", zap.Error(err), zap.Int64("affected", aff))
	}
	return nil
}

func (m *Member) SetBannedForSystem(uid string, sec int) (bool, error) {
	if err := m.SetBanned(uid, sec, true); err != nil {
		return false, err
	}

	me, ok, err := m.db.Find(uid)
	if err != nil {
		return false, err
	}
	if !ok {
		return false, nil
	}

	l, err := m.db.GetTodaySystemBannedLog(me.Id)
	if err != nil {
		log.Error("automatically banned up to 5 times for set banned", zap.Error(err), zap.Int("mid", me.Id))
	} else {
		now := time.Now()
		nowUnix := now.Unix()
		zeroUnix, err := time.ParseInLocation("2006-01-02 15:04:05", now.Format("2006-01-02 0:00:00"), time.Local)
		if err != nil {
			log.Error("parse time layout for set banned", zap.Error(err), zap.Int("mid", me.Id))
		} else if len(l) >= 5 {
			for _, v := range l {
				cat := v.CreateAt.Unix()
				if !(zeroUnix.Unix() <= cat && cat <= nowUnix) {
					return false, nil
				}
			}

			ok, err := m.SetBlockade(uid)
			if err != nil || !ok {
				log.Error("set blockade for set banned", zap.Error(err), zap.Bool("action", ok), zap.Int("mid", me.Id))
			}
			return true, nil
		}
	}
	return false, nil
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
	me, ok, err := m.db.Find(uid)
	if err != nil {
		return err
	}
	if !ok {
		return errors.ErrNoRows
	}
	if err := m.c.delBanned(uid); err != nil {
		return err
	}
	if !me.IsMessage {
		aff, err := m.db.UpdateIsMessage(me.Id, true)
		if err != nil {
			return err
		}
		if aff != 1 {
			return errors.ErrNoRows
		}

		me.IsMessage = true
		if err := m.c.set(me); err != nil {
			return err
		}
	}
	return nil
}
