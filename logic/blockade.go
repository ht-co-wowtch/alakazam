package logic

// TODO 需要踢人
func (l *Logic) SetBlockade(uid string) (bool, error) {
	m, ok, err := l.db.Find(uid)
	if err != nil {
		return false, err
	}
	if !ok {
		return false, nil
	}
	if m.IsBlockade {
		return true, nil
	}
	aff, err := l.db.SetBlockade(uid)
	return aff == 1, err
}

func (l *Logic) RemoveBlockade(uid string) (bool, error) {
	m, ok, err := l.db.Find(uid)
	if err != nil {
		return false, err
	}
	if !ok {
		return false, nil
	}
	if !m.IsBlockade {
		return true, nil
	}
	aff, err := l.db.DeleteBanned(uid)
	return aff == 1, err
}
