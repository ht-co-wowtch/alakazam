package request

import (
	"encoding/json"
)

func PushRoom(uid, key, message string) Response {
	j := map[string]interface{}{
		"uid":     uid,
		"key":     key,
		"message": message,
	}
	b, _ := json.Marshal(j)
	return PostJson(host+"/push/room", b)
}

func PushBroadcast(roomId []string, message string) Response {
	j := map[string]interface{}{
		"room_id": roomId,
		"message": message,
	}
	b, _ := json.Marshal(j)
	return PostJson(adminHost+"/push/all", b)
}
