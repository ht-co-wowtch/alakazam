package request

import (
	"encoding/json"
	"net/url"
)

func SetBanned(uid, remark string, sec int) Response {
	j := map[string]interface{}{
		"uid":     uid,
		"expired": sec,
		"remark":  remark,
	}
	b, _ := json.Marshal(j)
	return PostJson(adminHost+"/banned", b)
}

func DeleteBanned(uid string) Response {
	d := url.Values{}
	d.Set("uid", uid)
	return Delete(adminHost+"/banned", d)
}

func SetBlockade(uid, remark string) Response {
	j := map[string]interface{}{
		"uid":    uid,
		"remark": remark,
	}
	b, _ := json.Marshal(j)
	return PostJson(adminHost+"/blockade", b)
}
