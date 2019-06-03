package front

import (
	"github.com/stretchr/testify/assert"
	"gitlab.com/jetfueltw/cpw/alakazam/activity"
	"gitlab.com/jetfueltw/cpw/alakazam/errors"
	"gitlab.com/jetfueltw/cpw/alakazam/logic"
	"gitlab.com/jetfueltw/cpw/alakazam/logic/http/front"
	"gitlab.com/jetfueltw/cpw/alakazam/test/internal/request"
	"net/http"
	"testing"
)

func TestGiveLuckyMoney(t *testing.T) {
	r := giveLuckyMoney(1, 1, "test", activity.Money)

	assert.Equal(t, http.StatusNoContent, r.StatusCode)
}

func TestGiveLuckyMoneyMinAmountBy0_01(t *testing.T) {
	r := giveLuckyMoney(0.001, 1, "test", activity.Money)

	shouldBeGiveLuckyMoneyError(t, r, "红包金额最低0.01")
}

func TestGiveLuckyMoneyMaxCountBy500(t *testing.T) {
	r := giveLuckyMoney(1, 501, "test", activity.Money)

	shouldBeGiveLuckyMoneyError(t, r, "红包最大数量是500")
}

func TestGiveLuckyMoneyMaxMessageChatBy20(t *testing.T) {
	s := ""
	for i := 0; i <= 20; i++ {
		s += "1"
	}

	r := giveLuckyMoney(1, 1, s, activity.Money)

	shouldBeGiveLuckyMoneyError(t, r, "限制文字长度为1到20个字")
}

func TestGiveLuckyMoneyTypeError(t *testing.T) {
	r := giveLuckyMoney(1, 1, "test", 3)

	shouldBeGiveLuckyMoneyError(t, r, errors.DataError.Message)
}

func giveLuckyMoney(amount float32, count int, message string, model int) request.Response {
	return request.GiveLuckyMoney(front.LuckyMoney{
		User: logic.User{
			Uid: "82ea16cd2d6a49d887440066ef739669",
			Key: "f0962f33-b444-4ac0-8be9-2a8423178212",
		},
		GiveMoney: activity.GiveMoney{
			Amount:  amount,
			Count:   count,
			Message: message,
			Type:    model,
		},
	})
}

func shouldBeGiveLuckyMoneyError(t *testing.T, r request.Response, message string) {
	e := request.ToError(t, r.Body)
	e.Status = r.StatusCode
	assert.Equal(t, errors.DataError.Mes(message), e)
}
