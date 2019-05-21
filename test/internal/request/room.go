package request

import (
	"encoding/json"
	"fmt"
	"gitlab.com/jetfueltw/cpw/alakazam/logic/store"
	"net/url"
)

func SetRoom(room store.Room) Response {
	b, _ := json.Marshal(room)
	return PostJson(adminHost+"/room", b)
}

func GetRoom(roomId int) Response {
	return Get(fmt.Sprintf(adminHost+"/room/%d", roomId), url.Values{})
}
