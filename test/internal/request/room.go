package request

import (
	"encoding/json"
	"gitlab.com/jetfueltw/cpw/alakazam/logic/store"
)

func SetRoom(room store.Room) Response {
	b, _ := json.Marshal(room)
	return PostJson(adminHost+"/room", b)
}
