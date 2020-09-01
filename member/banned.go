package member

import (
	"gitlab.com/jetfueltw/cpw/micro/log"
	"go.uber.org/zap"
	"time"
)

func (m *Member) SetBanned(uid string, rid, sec int, isSystem bool) error {
	expire := time.Duration(sec) * time.Second
	if sec > 0 {
		if err := m.c.setBanned(uid, rid, expire); err != nil {
			return err
		}
	} else if sec == -1 {
		u, err := m.GetByRoom(uid, rid)
		if err != nil {
			return err
		}

		if u.IsBanned {
			return nil
		}

		u.IsBanned = true

		if err := m.db.SetPermission(*u); err != nil {
			return err
		}

		return m.c.set(u)
	}

	u, _ := m.c.get(uid)

	ok, err := m.db.SetBannedLog(u.Id, expire, isSystem)
	if err != nil || !ok {
		log.Error("set banned log", zap.Error(err), zap.Bool("action", ok))
	}
	return nil
}

func (m *Member) SetBannedForSystem(uid string, rid, sec int) (bool, error) {
	if err := m.SetBanned(uid, sec, rid, true); err != nil {
		return false, err
	}

	me, err := m.c.get(uid)
	if err != nil {
		return false, err
	}

	l, err := m.db.GetTodaySystemBannedLog(me.Id)
	if err != nil {
		log.Error("automatically banned up to 5 times for set banned", zap.Error(err), zap.Int64("mid", me.Id))
	} else {
		now := time.Now()
		nowUnix := now.Unix()
		zeroUnix, err := time.ParseInLocation("2006-01-02 15:04:05", now.Format("2006-01-02 0:00:00"), time.Local)
		if err != nil {
			log.Error("parse time layout for set banned", zap.Error(err), zap.Int64("mid", me.Id))
		} else if len(l) >= 5 {
			for _, v := range l {
				cat := v.CreateAt.Unix()
				if !(zeroUnix.Unix() <= cat && cat <= nowUnix) {
					return false, nil
				}
			}

			if err := m.SetBlockade(uid, rid, true); err != nil {
				log.Error("set blockade for set banned", zap.Error(err), zap.Int64("mid", me.Id))
			}
			return true, nil
		}
	}
	return false, nil
}

func (m *Member) RemoveBanned(uid string, rid int) error {
	u, err := m.c.getByRoom(uid, rid)
	if err != nil {
		return err
	}

	if err = m.c.delBanned(uid, rid); err != nil {
		return err
	}

	u.IsBanned = false

	if err := m.db.SetPermission(*u); err != nil {
		return err
	}

	return m.c.set(u)
}
