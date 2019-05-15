package request

import (
	"fmt"
	"net/url"
)

func PushRoom(uid, key, message string) Response {
	data := url.Values{}
	data.Set("uid", uid)
	data.Set("key", key)
	data.Set("message", message)
	return Post(host+"/push/room", data)
}

func PushBroadcast(uid, key, message string, roomId []string, ) Response {
	data := url.Values{
		"room_id": roomId,
	}
	data.Set("uid", uid)
	data.Set("key", key)
	data.Set("message", message)
	return Post(fmt.Sprintf(adminHost+"/push/all"), data)
}
