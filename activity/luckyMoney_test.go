package activity

import (
	. "github.com/smartystreets/goconvey/convey"
	"gitlab.com/jetfueltw/cpw/alakazam/client"
	"gitlab.com/jetfueltw/cpw/alakazam/errors"
	"testing"
)

func TestGiveMoney(t *testing.T) {
	Convey("發紅包", t, func() {
		money := NewLuckyMoney(&client.Client{})

		Convey("輸入紅包金額", func() {
			err := money.Give(&GiveMoney{Amount: 100.001})

			Convey("不能小數點第三位", func() {
				So(err, ShouldResemble, errors.AmountError)
			})
		})
	})
}
