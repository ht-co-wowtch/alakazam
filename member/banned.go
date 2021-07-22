package member

import (
	"time"

	"gitlab.com/jetfueltw/cpw/alakazam/errors"
	"gitlab.com/jetfueltw/cpw/micro/log"
	"go.uber.org/zap"
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

		if u.Banned() {
			return nil
		}

		u.Permission.RoomId = int64(rid)
		u.Permission.IsBanned = true

		if err := m.db.SetRoomPermission(*u); err != nil {
			return err
		}

		return m.c.set(u)
	}

	u, err := m.c.get(uid)
	if err != nil {
		return err
	}
	ok, err := m.db.SetBannedLog(u.Id, expire, isSystem)

	if err != nil || !ok {
		log.Error("set banned log", zap.Error(err), zap.Bool("action", ok))
		return err
	}
	return nil
}

func (m *Member) SetBannedAll(uid string, sec int) error {
	expire := time.Duration(sec) * time.Second
	if sec > 0 {
		//利用redis的時間進行控制
		log.Debug("member/banned.go 設定redis禁言快取時間", zap.String("uid", uid), zap.Int("sec", sec))
		if err := m.c.setBanned(uid, 0, expire); err != nil {

			return err
		}

	} else if sec == -1 {
		if _, err := m.db.SetUserBanned(uid, true); err != nil {
			return err
		}

		member, err := m.Get(uid)

		if err != nil {
			if err == errors.ErrLogin {
				return nil
			}
			return err
		}
		member.IsMessage = false
		return m.c.set(member)
	}
	return nil
}

func (m *Member) SetBannedForSystem(uid string, rid, sec int) (bool, error) {
	if err := m.SetBanned(uid, rid, sec, true); err != nil {
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

	u.Permission.RoomId = int64(rid)
	u.Permission.IsBanned = false

	// 更新db中會員於個別房間權限
	if err := m.db.SetRoomPermission(*u); err != nil {
		return err
	}

	return m.c.set(u)
}

//  對全站解除禁言
func (m *Member) RemoveBannedAll(uid string) error {
	var (
		err    error
		banned = false
	)

	//解Banned
	// db
	_, err = m.db.SetUserBanned(uid, banned)
	if err != nil {
		return err
	}

	// cahce
	if err = m.c.delAllBanned(uid); err != nil {
		log.Error("member/banned.go RemoveBannedAll.0", zap.Error(err))
		return err
	}

	member, err := m.Get(uid)

	if err != nil {
		log.Error("member/banned.go RemoveBannedAll.1", zap.Error(err))
		if err == errors.ErrLogin {
			return nil
		}
		return err
	}

	// 更新會員cache
	member.IsMessage = true
	return m.c.set(member)
}
