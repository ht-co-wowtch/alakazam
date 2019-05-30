package request

import (
	"encoding/json"
	"fmt"
	"gitlab.com/jetfueltw/cpw/alakazam/logic/conf"
)

func PushRoom(uid, key, message string) Response {
	j := map[string]interface{}{
		"uid":     uid,
		"key":     key,
		"message": message,
	}
	b, _ := json.Marshal(j)
	return PostJson(getHost()+"/push/room", b)
}

func PushRoomNotToken(uid, key, message string) Response {
	j := map[string]interface{}{
		"uid":     uid,
		"key":     key,
		"message": message,
	}
	b, _ := json.Marshal(j)
	return PostJsonNotToken(getHost()+"/push/room", b)
}

func PushBroadcast(roomId []string, message string) Response {
	j := map[string]interface{}{
		"room_id": roomId,
		"message": message,
	}
	b, _ := json.Marshal(j)
	return PostJson(getAdminHost()+"/push/all", b)
}

func getHost() string {
	return fmt.Sprintf("http://127.0.0.1%s", conf.Conf.HTTPServer.Addr)
}
