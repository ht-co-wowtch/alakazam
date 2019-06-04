package store

type Room struct {
	// 要設定的房間id
	Id string `json:"id"`

	// 是否禁言
	IsMessage bool `json:"is_message"`

	// 是否可發/跟注
	IsFollow bool `json:"is_follow"`

	// 儲值&打碼量發話限制
	Limit Limit `json:"limit"`

	// 紅包多久過期
	LuckyMoneyExpire int `json:"lucky_money_expire"`
}

type Limit struct {
	// 限制範圍
	Day int `json:"day"`

	// 儲值金額
	Amount int `json:"amount"`

	// 打碼量
	Dml int `json:"dml"`
}

func (s *Store) CreateRoom(room Room) (int64, error) {
	sql := "INSERT INTO rooms (room_id, is_message, is_follow, day_limit, amount_limit, dml_limit) VALUES (?, ?, ?, ?, ?, ?)"
	r, err := s.Exec(sql, room.Id, room.IsMessage, room.IsFollow, room.Limit.Day, room.Limit.Amount, room.Limit.Dml)
	if err != nil {
		return 0, err
	}
	return r.RowsAffected()
}

func (s *Store) UpdateRoom(room Room) (int64, error) {
	sql := "UPDATE rooms SET is_message = ?, is_follow = ?, day_limit = ?, amount_limit = ?, dml_limit = ? WHERE room_id = ? "
	r, err := s.Exec(sql, room.IsMessage, room.IsFollow, room.Limit.Day, room.Limit.Amount, room.Limit.Dml, room.Id)
	if err != nil {
		return 0, err
	}
	return r.RowsAffected()
}

func (s *Store) GetRoom(roomId string) (Room, error) {
	room := Room{}
	sql := "SELECT * FROM rooms WHERE room_id = ?"
	return room, s.QueryRow(sql, roomId).
		Scan(&room.Id, &room.IsMessage, &room.IsFollow, &room.Limit.Day, &room.Limit.Amount, &room.Limit.Dml)
}
