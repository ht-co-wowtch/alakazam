package member

import "gitlab.com/jetfueltw/cpw/alakazam/errors"

func (m *Member) SetBlockade(uid string, rid int, set bool) error {
	member, err := m.GetByRoom(uid, rid)
	if err != nil {
		return err
	}

	member.Permission.RoomId = int64(rid)
	member.Permission.IsBlockade = set

	if err := m.db.SetRoomPermission(*member); err != nil {
		return err
	}

	return m.c.set(member)
}

// 更新會員封鎖狀態
func (m *Member) SetBlockadeAll(uid string, set bool) error {
	// 更新db中會員封鎖狀態
	if _, err := m.db.SetUserBlockade(uid, set); err != nil {
		return err
	}

	// 從cache或redis中取得會員資料
	member, err := m.Get(uid)
	if err != nil {
		if err == errors.ErrLogin {
			return nil
		}
		return err
	}

	member.IsBlockade = set

	// 更新cache中會員封鎖狀態
	if err := m.c.set(member); err != nil {
		return err
	}

	// 清除房間快取
	return m.c.clearRoom(uid)
}
