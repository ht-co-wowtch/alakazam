package store

func (s *Store) SetBlockade(uid, remark string) (int64, error) {
	return s.updateBlockade(uid, remark, true)
}

func (s *Store) DeleteBanned(uid string) (int64, error) {
	return s.updateBlockade(uid, "", false)
}

func (s *Store) updateBlockade(uid, remark string, is bool) (int64, error) {
	sql := "UPDATE members SET is_blockade = ? WHERE uid = ?"
	r, err := s.Exec(sql, is, uid)
	if err != nil {
		return 0, err
	}
	return r.RowsAffected()
}
