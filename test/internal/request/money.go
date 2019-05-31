package request

import (
	"encoding/json"
	"gitlab.com/jetfueltw/cpw/alakazam/logic/http/front"
)

func GiveLuckyMoney(money front.LuckyMoney) Response {
	b, _ := json.Marshal(money)
	return PostJson(getHost()+"/give-lucky-money", b)
}
