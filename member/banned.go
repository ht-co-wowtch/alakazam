package member

import (
	"database/sql"
	"gitlab.com/jetfueltw/cpw/alakazam/errors"
	"gitlab.com/jetfueltw/cpw/micro/log"
	"go.uber.org/zap"
	"time"
)

var (
	errFailure = errors.New("失敗")
)

func (m *Member) SetBanned(uid string, sec int, isSystem bool) error {
	me, err := m.db.Find(uid)
	if err != nil {
		if err == sql.ErrNoRows {
			return errors.ErrNoMember
		}
		return err
	}

	expire := time.Duration(sec) * time.Second
	if sec > 0 {
		ok, err := m.c.setBanned(uid, expire)
		if err != nil {
			return err
		}
		if !ok {
			return errFailure
		}
	} else if sec == -1 {
		if !me.IsMessage {
			return nil
		}

		ok, err := m.db.UpdateIsMessage(me.Id, false)
		if err != nil {
			return err
		}
		if !ok {
			return errFailure
		}

		expire = time.Duration(0)

		me.IsMessage = false
		ok, err = m.c.set(me)
		if err != nil {
			return err
		}
		if !ok {
			return errFailure
		}
	}

	ok, err := m.db.SetBannedLog(me.Id, expire, isSystem)
	if err != nil || !ok {
		log.Error("set banned log", zap.Error(err), zap.Bool("action", ok))
	}
	return nil
}

func (m *Member) SetBannedForSystem(uid string, sec int) (bool, error) {
	if err := m.SetBanned(uid, sec, true); err != nil {
		return false, err
	}

	me, err := m.db.Find(uid)
	if err != nil {
		if err == sql.ErrNoRows {
			return false, errors.ErrNoMember
		}
		return false, err
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

func (m *Member) RemoveBanned(uid string) error {
	me, err := m.db.Find(uid)
	if err != nil {
		if err == sql.ErrNoRows {
			return errors.ErrNoMember
		}
		return err
	}

	ok, err := m.c.delBanned(uid)
	if err != nil {
		return err
	}

	// 如果redis內沒有禁言時效資料且用戶發言狀態為true
	if me.IsMessage && !ok {
		return nil
	}

	if !me.IsMessage {
		ok, err := m.db.UpdateIsMessage(me.Id, true)
		if err != nil {
			return err
		}
		if ok {
			me.IsMessage = true
			ok, err = m.c.set(me)
			if err != nil {
				return err
			}
			if !ok {
				return errors.New("remove banned in set redis for admin api")
			}
		}
	}
	return nil
}
