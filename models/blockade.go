package models

func (s *Store) SetBlockade(uid string) (int64, error) {
	return s.updateBlockade(uid, true)
}

func (s *Store) DeleteBanned(uid string) (int64, error) {
	return s.updateBlockade(uid, false)
}

func (s *Store) updateBlockade(uid string, is bool) (int64, error) {
	return s.d.Cols("is_blockade").
		Where("uid = ?", uid).
		Update(&Member{
			IsBlockade: is,
		})
}
