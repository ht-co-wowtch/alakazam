package store

const (
	// 訪客
	Guest = 0

	// 營銷
	Marketing = 1

	// 玩家
	Player = 2
)

type User struct {
	Uid        string
	Permission int
}

// 新增會員
func (s *Store) CreateUser(uid string, permission int) (int64, error) {
	sql := "INSERT INTO members (uid, permission, create_at) VALUES (?, ?, CURRENT_TIMESTAMP)"
	r, err := s.Exec(sql, uid, permission)
	if err != nil {
		return 0, err
	}
	return r.RowsAffected()
}

// 找會員
func (s *Store) FindUserPermission(uid string) (permission int, isBlockade bool, err error) {
	sql := "SELECT permission, is_blockade FROM members WHERE uid = ?"
	return permission, isBlockade, s.QueryRow(sql, uid).Scan(&permission, &isBlockade)
}
