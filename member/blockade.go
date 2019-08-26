package member

import "database/sql"

func (m *Member) SetBlockade(uid string) (bool, error) {
	me, err := m.db.Find(uid)
	if err != nil {
		if err == sql.ErrNoRows {
			return false, nil
		}
		return false, err
	}
	if me.IsBlockade {
		return true, nil
	}
	aff, err := m.db.SetBlockade(uid)
	return aff >= 1, err
}

func (m *Member) RemoveBlockade(uid string) (bool, error) {
	me, err := m.db.Find(uid)
	if err != nil {
		if err == sql.ErrNoRows {
			return false, nil
		}
		return false, err
	}
	if !me.IsBlockade {
		return true, nil
	}
	aff, err := m.db.DeleteBanned(uid)
	return aff == 1, err
}
