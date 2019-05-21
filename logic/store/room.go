package store

type Room struct {
	// 要設定的房間id
	RoomId int `json:"room_id" binding:"required"`

	// 是否禁言
	IsMessage bool `json:"is_message"`

	// 是否可發/搶紅包
	IsBonus bool `json:"is_bonus"`

	// 是否可發/跟注
	IsFollow bool `json:"is_follow"`

	// 儲值&打碼量發話限制
	Limit Limit `json:"limit"`
}

type Limit struct {
	// 限制範圍
	Day int `json:"day"`

	// 儲值金額
	Amount int `json:"amount"`

	// 打碼量
	Dml int `json:"dml"`
}

func (d *Store) SetRoom(room Room) (int64, error) {
	sql := "INSERT INTO rooms (room_id, is_message, is_bonus, is_follow, day_limit, amount_limit, dml_limit) VALUES (?, ?, ?, ?, ?, ?, ?)"
	r, err := d.Exec(sql, room.RoomId, room.IsMessage, room.IsBonus, room.IsFollow, room.Limit.Day, room.Limit.Amount, room.Limit.Dml)
	if err != nil {
		return 0, err
	}
	return r.RowsAffected()
}

func (d *Store) GetRoom(roomId int) (Room, error) {
	room := Room{}
	sql := "SELECT * FROM rooms WHERE room_id = ?"
	return room, d.QueryRow(sql, roomId).
		Scan(&room.RoomId, &room.IsMessage, &room.IsBonus, &room.IsFollow, &room.Limit.Day, &room.Limit.Amount, &room.Limit.Dml)
}
