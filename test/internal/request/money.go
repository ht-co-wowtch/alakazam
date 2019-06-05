package request

import (
	"encoding/json"
	"gitlab.com/jetfueltw/cpw/alakazam/logic/http/front"
)

func GiveLuckyMoney(money front.LuckyMoney) Response {
	b, _ := json.Marshal(money)
	return PostJson(getHost()+"/give-lucky-money", b)
}

func TakeLuckyMoney(token string) Response {
	var p struct {
		Token string `json:"token"`
	}
	p.Token = token
	b, _ := json.Marshal(p)
	return PostJson(getHost()+"/take-lucky-money", b)
}
