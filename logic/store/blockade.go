package store

func (d *Store) SetBlockade(uid, remark string) (int64, error) {
	return d.updateBlockade(uid, remark, true)
}

func (d *Store) DeleteBanned(uid string) (int64, error) {
	return d.updateBlockade(uid, "", false)
}

func (d *Store) updateBlockade(uid, remark string, is bool) (int64, error) {
	sql := "UPDATE members SET is_blockade = ? WHERE uid = ?"
	r, err := d.Exec(sql, is, uid)
	if err != nil {
		return 0, err
	}
	return r.RowsAffected()
}
