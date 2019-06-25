package store

const (
	// 訪客
	Guest = "guest"

	// 營銷
	Marketing = "marketing"

	// 玩家
	Player = 2
)

type User struct {
	Uid        string
	Name       string
	Avatar     string
	IsBlockade bool

	// user權限權重
	Permission int
}

// 新增會員
func (s *Store) CreateUser(user *User) (int64, error) {
	sql := "INSERT INTO members (uid, name, avatar, permission, create_at) VALUES (?, ?, ?, ?, CURRENT_TIMESTAMP)"
	r, err := s.Exec(sql, user.Uid, user.Name, user.Avatar, user.Permission)

	if err != nil {
		return 0, err
	}

	return r.RowsAffected()
}

func (s *Store) UpdateUser(user *User) (int64, error) {
	sql := "UPDATE members SET name = ?, avatar = ? WHERE uid = ?"
	r, err := s.Exec(sql, user.Name, user.Avatar, user.Uid)
	if err != nil {
		return 0, err
	}
	return r.RowsAffected()
}

// 找會員
func (s *Store) Find(uid string) (*User, error) {
	user := &User{Uid: uid}

	sql := "SELECT name, avatar, permission, is_blockade FROM members WHERE uid = ?"
	err := s.QueryRow(sql, uid).Scan(&user.Name, &user.Avatar, &user.Permission, &user.IsBlockade)

	return user, err
}
