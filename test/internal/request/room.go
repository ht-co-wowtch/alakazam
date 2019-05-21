package request

import (
	"encoding/json"
	"fmt"
	"gitlab.com/jetfueltw/cpw/alakazam/logic/store"
	"net/url"
)

func CreateRoom(room store.Room) Response {
	b, _ := json.Marshal(room)
	return PostJson(adminHost+"/room", b)
}

func UpdateRoom(roomId string, room store.Room) Response {
	b, _ := json.Marshal(room)
	return PutJson(fmt.Sprintf(adminHost+"/room/%s", roomId), b)
}

func GetRoom(roomId string) Response {
	return Get(fmt.Sprintf(adminHost+"/room/%s", roomId), url.Values{})
}
