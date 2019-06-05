package front

import (
	"encoding/json"
	. "github.com/smartystreets/goconvey/convey"
	"gitlab.com/jetfueltw/cpw/alakazam/activity"
	"gitlab.com/jetfueltw/cpw/alakazam/client"
	"gitlab.com/jetfueltw/cpw/alakazam/errors"
	"gitlab.com/jetfueltw/cpw/alakazam/logic"
	"gitlab.com/jetfueltw/cpw/alakazam/logic/http/front"
	pd "gitlab.com/jetfueltw/cpw/alakazam/protocol"
	"gitlab.com/jetfueltw/cpw/alakazam/test/internal/protocol"
	"gitlab.com/jetfueltw/cpw/alakazam/test/internal/request"
	"gitlab.com/jetfueltw/cpw/alakazam/test/internal/run"
	"io/ioutil"
	"net/http"
	"testing"
	"time"
)

func TestGiveLuckyMoney(t *testing.T) {
	Convey("假設要發送紅包", t, func() {
		model := activity.Money
		message := "test"

		mockGiveLuckyMoneyApi(t)

		Convey("紅包金額最低0.01", func() {
			r := giveLuckyMoney(request.Auth{}, 0.001, 1, message, model)
			shouldBeGiveLuckyMoneyError(t, r, "红包金额最低0.01")
		})

		Convey("紅包數量最大500包", func() {
			r := giveLuckyMoney(request.Auth{}, 1, 501, message, model)

			shouldBeGiveLuckyMoneyError(t, r, "红包最大数量是500")
		})

		Convey("文案限制在1~20字元", func() {
			s := ""
			for i := 0; i <= 20; i++ {
				s += "1"
			}

			r := giveLuckyMoney(request.Auth{}, 1, 1, s, model)

			shouldBeGiveLuckyMoneyError(t, r, "限制文字长度为1到20个字")
		})

		Convey("紅包種類必須是普通或拼手氣兩種", func() {
			r := giveLuckyMoney(request.Auth{}, 1, 1, message, 3)

			shouldBeGiveLuckyMoneyError(t, r, errors.DataError.Message)
		})

		Convey("發普通紅包", func() {
			auth, err := request.DialAuth("dkflkfldfklfd")

			model = activity.Money

			So(err, ShouldBeNil)

			Convey("餘額足夠", func() {
				r := giveLuckyMoney(auth, 1, 1, message, model)

				So(r.StatusCode, ShouldEqual, http.StatusNoContent)
			})

			Convey("餘額不足夠", func() {
				r := giveLuckyMoney(auth, 2, 3, message, model)

				e := request.ToError(t, r.Body)
				e.Status = r.StatusCode
				So(e, ShouldResemble, errors.BalanceError)
			})
		})

		Convey("發拼手氣紅包", func() {
			auth, err := request.DialAuth("e3jdhffdgjf")

			model = activity.LuckMoney

			So(err, ShouldBeNil)

			Convey("餘額足夠", func() {
				r := giveLuckyMoney(auth, 2, 3, message, model)

				So(r.StatusCode, ShouldEqual, http.StatusNoContent)
			})

			Convey("餘額不足夠", func() {
				r := giveLuckyMoney(auth, 5, 2, message, model)

				e := request.ToError(t, r.Body)
				e.Status = r.StatusCode
				So(e, ShouldResemble, errors.BalanceError)
			})
		})
	})
}

func TestGiveLuckyMoneyMessage(t *testing.T) {
	var auth request.Auth

	Convey("假設要發送紅包", t, func() {
		mockGiveLuckyMoneyApi(t)

		var err error
		var msg = "test"

		roomId := "10000"

		Convey("先進入聊天室", func() {
			auth, err = request.DialAuth(roomId)

			So(err, ShouldBeNil)
		})

		Convey("點擊發送紅包", func() {
			r := giveLuckyMoney(auth, 1, 3, msg, activity.Money)

			So(r.StatusCode, ShouldEqual, http.StatusNoContent)
		})

		Convey("聊天室出現紅包訊息", func() {
			time.Sleep(time.Second)

			p, err := protocol.ReadMessage(auth.Rd, auth.Proto)

			So(err, ShouldBeNil)

			var message = new(logic.Money)
			err = json.Unmarshal(p[0].Body, &message)

			So(err, ShouldBeNil)
			So(p[0].Op, ShouldEqual, pd.OpMoney)
			So(auth.Uid, ShouldEqual, message.Uid)
			So(message.Message.Message, ShouldEqual, msg)
			So(message.Name, ShouldNotBeEmpty)
			So(message.Token, ShouldNotBeEmpty)
			So(time.Unix(message.Expired, 0).IsZero(), ShouldBeFalse)
		})
	})
}

func TestTakeLuckyMoney(t *testing.T) {
	var auth request.Auth
	var message = new(logic.Money)

	mockGiveLuckyMoneyApi(t)

	Convey("假設要搶紅包", t, func() {
		var err error

		roomId := "10000"
		msg := "測試"

		Convey("先進入聊天室", func() {
			auth, err = request.DialAuth(roomId)

			So(err, ShouldBeNil)
		})

		Convey("看到紅包訊息", func() {

			r := giveLuckyMoney(auth, 1, 3, msg, activity.Money)

			if r.StatusCode != http.StatusNoContent {
				t.Fatalf("giveLuckyMoney error(%s)", string(r.Body))
			}

			time.Sleep(time.Second)

			p, err := protocol.ReadMessage(auth.Rd, auth.Proto)

			if err != nil || p == nil {
				t.Fatalf("Read Message error(%v) body(%v)", err, p)
			}

			err = json.Unmarshal(p[0].Body, &message)

			if err != nil {
				t.Fatalf("json.Unmarshal error(%v)", err)
			}
		})

		Convey("點擊紅包", func() {
			r := request.TakeLuckyMoney(message.Token)

			Convey("搶到一元紅包", func() {

				var p struct {
					Id         string `json:"id"`
					Name       string `json:"name"`
					Avatar     string `json:"avatar"`
					LuckyMoney struct {
						Message string  `json:"message"`
						Amount  float64 `json:"amount"`
					}
				}

				err := json.Unmarshal(r.Body, &p)

				if err != nil {
					t.Fatalf("json.Unmarshal error(%v)", err)
				}

				So(http.StatusOK, ShouldEqual, r.StatusCode)
				So(p.Id, ShouldNotBeEmpty)
				So(p.Name, ShouldNotBeEmpty)
				So(p.Avatar, ShouldNotBeEmpty)
				So(msg, ShouldEqual, p.LuckyMoney.Message)
				So(1, ShouldEqual, p.LuckyMoney.Amount)
			})
		})
	})
}

func mockGiveLuckyMoneyApi(t *testing.T) {
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
}

func giveLuckyMoney(auth request.Auth, amount float64, count int, message string, model int) request.Response {
	return request.GiveLuckyMoney(front.LuckyMoney{
		User: logic.User{
			Uid: auth.Uid,
			Key: auth.Key,
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
	expected := errors.DataError
	expected.Message = message

	So(e, ShouldResemble, expected)
}
