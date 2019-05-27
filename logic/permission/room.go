package permission

import "gitlab.com/jetfueltw/cpw/alakazam/logic/store"

const (
	RoomDefaultPermission = Message + sendFollow + getFollow + sendBonus + getBonus
)

func ToRoomInt(room store.Room) int {
	i := 0
	if room.IsMessage {
		i += Message
	}
	if room.IsFollow {
		i += sendFollow + getFollow
	}
	if room.IsBonus {
		i += sendBonus + getBonus
	}
	if room.Limit.Day > 0 && room.Limit.Dml+room.Limit.Amount > 0 {
		i += money
	}
	return i
}
