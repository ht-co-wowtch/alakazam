package request

import (
	"encoding/json"
	"fmt"
	"gitlab.com/jetfueltw/cpw/alakazam/logic/conf"
	"net/url"
)

func SetBanned(uid, remark string, sec int) Response {
	j := map[string]interface{}{
		"uid":     uid,
		"expired": sec,
		"remark":  remark,
	}
	b, _ := json.Marshal(j)
	return PostJson(getAdminHost()+"/banned", b)
}

func DeleteBanned(uid string) Response {
	d := url.Values{}
	d.Set("uid", uid)
	return Delete(getAdminHost()+"/banned", d)
}

func SetBlockade(uid, remark string) Response {
	j := map[string]interface{}{
		"uid":    uid,
		"remark": remark,
	}
	b, _ := json.Marshal(j)
	return PostJson(getAdminHost()+"/blockade", b)
}

func DeleteBlockade(uid string) Response {
	d := url.Values{}
	d.Set("uid", uid)
	return Delete(getAdminHost()+"/blockade", d)
}

func getAdminHost() string {
	return fmt.Sprintf("http://127.0.0.1%s", conf.Conf.HTTPAdminServer.Addr)
}
