package request

import (
	"encoding/json"
	"fmt"
	"net/url"
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

func PushBroadcast(uid, key, message string, roomId []string) Response {
	data := url.Values{
		"room_id": roomId,
	}
	data.Set("uid", uid)
	data.Set("key", key)
	data.Set("message", message)
	return Post(fmt.Sprintf(adminHost+"/push/all"), data)
}
