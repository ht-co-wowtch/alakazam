package request

import (
	"encoding/json"
	"fmt"
	"gitlab.com/jetfueltw/cpw/alakazam/logic"
	"net/url"
)

func CreateRoom(room logic.Room) Response {
	b, _ := json.Marshal(room)
	return PostJson(getAdminHost()+"/room", b)
}

func UpdateRoom(room logic.Room) Response {
	b, _ := json.Marshal(room)
	return PutJson(getAdminHost()+"/room", b)
}

func GetRoom(roomId string) Response {
	return Get(fmt.Sprintf(getAdminHost()+"/room/%s", roomId), url.Values{})
}
