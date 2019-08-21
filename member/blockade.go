package member

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
	return aff == 1, err
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
