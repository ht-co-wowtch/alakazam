package permission

import "gitlab.com/jetfueltw/cpw/alakazam/logic/store"

const (
	RoomDefaultPermission = Message + sendFollow + getFollow + getBonus
)

func ToRoomInt(room store.Room) int {
	i := 0
	if room.IsMessage {
		i += Message
	}
	if room.IsFollow {
		i += sendFollow + getFollow
	}
	if room.DayLimit > 0 && room.DmlLimit+room.DepositLimit > 0 {
		i += money
	}
	return i
}
