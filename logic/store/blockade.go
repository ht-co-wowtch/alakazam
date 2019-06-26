package store

func (s *Store) SetBlockade(uid, remark string) (int64, error) {
	return s.updateBlockade(uid, remark, true)
}

func (s *Store) DeleteBanned(uid string) (int64, error) {
	return s.updateBlockade(uid, "", false)
}

func (s *Store) updateBlockade(uid, remark string, is bool) (int64, error) {
	return s.d.Cols("is_blockade").
		Where("uid = ?", uid).
		Update(&Member{
			IsBlockade: is,
		})
}
