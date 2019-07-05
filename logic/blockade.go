package logic

func (l *Logic) SetBlockade(uid, remark string) (bool, error) {
	aff, err := l.db.SetBlockade(uid, remark)
	return aff >= 1, err
}

func (l *Logic) RemoveBlockade(uid string) (bool, error) {
	aff, err := l.db.DeleteBanned(uid)
	return aff >= 1, err
}
