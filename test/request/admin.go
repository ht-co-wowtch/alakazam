package request

import "encoding/json"

func SetBanned(uid, remark string, sec int) Response {
	j := map[string]interface{}{
		"uid":     uid,
		"expired": sec,
		"remark":  remark,
	}
	b, _ := json.Marshal(j)
	return PostJson(adminHost+"/banned", b)
}
