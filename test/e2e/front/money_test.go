package front

import (
	"encoding/json"
	. "github.com/smartystreets/goconvey/convey"
	"gitlab.com/jetfueltw/cpw/alakazam/activity"
	"gitlab.com/jetfueltw/cpw/alakazam/client"
	"gitlab.com/jetfueltw/cpw/alakazam/errors"
	"gitlab.com/jetfueltw/cpw/alakazam/logic"
	"gitlab.com/jetfueltw/cpw/alakazam/logic/http/front"
	"gitlab.com/jetfueltw/cpw/alakazam/test/internal/request"
	"gitlab.com/jetfueltw/cpw/alakazam/test/internal/run"
	"io/ioutil"
	"net/http"
	"testing"
)

func TestGiveLuckyMoney(t *testing.T) {
	Convey("發紅包", t, func() {
		model := activity.Money
		message := "test"

		run.AddClient("/give-lucky-money", func(req *http.Request) (response *http.Response, e error) {
			var p client.Older

			b, err := ioutil.ReadAll(req.Body)

			if err != nil {
				t.Fatalf("Request body ioutil ReadAll error(%v)", err)
			}

			if err := json.Unmarshal(b, &p); err != nil {
				t.Fatalf("Request body json Unmarshal error(%v)", err)
			}

			var j interface{}
			var statusCode int

			if p.Amount <= 4 {
				var r struct {
					Balance float32 `json:"balance"`
				}

				r.Balance = 1000
				j = r

				statusCode = http.StatusOK

			} else {
				j = errors.Error{
					Code:    15024020,
					Message: "Insufficient balance",
				}

				statusCode = http.StatusPaymentRequired
			}

			b, err = json.Marshal(j)

			if err != nil {
				t.Fatalf("json Marshal error(%v)", err)
			}

			return request.ToResponse(b, statusCode)
		})

		Convey("資料驗證", func() {
			Convey("紅包金額最低0.01", func() {
				r := giveLuckyMoney(0.001, 1, message, model)

				shouldBeGiveLuckyMoneyError(t, r, "红包金额最低0.01")
			})

			Convey("紅包數量最大500包", func() {
				r := giveLuckyMoney(1, 501, message, model)

				shouldBeGiveLuckyMoneyError(t, r, "红包最大数量是500")
			})

			Convey("文案限制在1~20字元", func() {
				s := ""
				for i := 0; i <= 20; i++ {
					s += "1"
				}

				r := giveLuckyMoney(1, 1, s, model)

				shouldBeGiveLuckyMoneyError(t, r, "限制文字长度为1到20个字")
			})

			Convey("紅包種類錯誤", func() {
				r := giveLuckyMoney(1, 1, message, 3)

				shouldBeGiveLuckyMoneyError(t, r, errors.DataError.Message)
			})
		})

		Convey("普通紅包", func() {
			model = activity.Money

			Convey("餘額足夠", func() {
				r := giveLuckyMoney(1, 1, message, model)

				So(r.StatusCode, ShouldEqual, http.StatusNoContent)
			})

			Convey("餘額不足夠", func() {
				r := giveLuckyMoney(2, 3, message, model)

				e := request.ToError(t, r.Body)
				e.Status = r.StatusCode
				So(e, ShouldResemble, errors.BalanceError)
			})
		})

		Convey("拼手氣紅包", func() {
			model = activity.LuckMoney

			Convey("餘額足夠", func() {
				r := giveLuckyMoney(2, 3, message, model)

				So(r.StatusCode, ShouldEqual, http.StatusNoContent)
			})

			Convey("餘額不足夠", func() {
				r := giveLuckyMoney(5, 2, message, model)

				e := request.ToError(t, r.Body)
				e.Status = r.StatusCode
				So(e, ShouldResemble, errors.BalanceError)
			})
		})
	})
}

func giveLuckyMoney(amount float64, count int, message string, model int) request.Response {
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
	So(e, ShouldResemble, errors.DataError.Mes(message))
}
