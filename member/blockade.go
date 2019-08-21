package member

import "gitlab.com/jetfueltw/cpw/alakazam/errors"

func (m *Member) SetBlockade(uid string) (bool, error) {
	me, ok, err := m.db.Find(uid)
	if err != nil {
		return false, err
	}
	if !ok {
		return false, nil
	}
	if me.IsBlockade {
		return true, nil
	}
	aff, err := m.db.SetBlockade(uid)
	if err != nil {
		return false, err
	}
	if aff != 1 {
		return false, errors.ErrNoRows
	}
	return m.c.delete(uid)
}

func (m *Member) RemoveBlockade(uid string) (bool, error) {
	me, ok, err := m.db.Find(uid)
	if err != nil {
		return false, err
	}
	if !ok {
		return false, nil
	}
	if !me.IsBlockade {
		return true, nil
	}
	aff, err := m.db.DeleteBanned(uid)
	return aff == 1, err
}
