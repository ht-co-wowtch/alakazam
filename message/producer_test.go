package message

import (
	"github.com/stretchr/testify/assert"
	"gitlab.com/jetfueltw/cpw/alakazam/errors"
	"testing"
)

func TestCheckMessage(t *testing.T) {
	testCases := []string{
		`  `,
		`<dfkfdjkfdj`,
		`</iframe>`,
		`<iframe`,
		`<iframe</iframe>`,
		`<iframe /iframe>`,
		`<iframe src="dd"`,
		`<iframe src`,
		`< iframe src=http://m.647751.com/#/lottery/games/9 <iframe height='900'>`,
		`<iframe src="http://m.647751.com/#/lottery/games/9"`,
		`<iframe src="http://m.647751.com/#/lottery/games/9"</iframe>`,
		`<iframe src="http://m.647751.com/#/lottery/games/9" <="" iframe="" height="900px">wddwdw`,
		`<iframe src="http://m.647751.com/#/lottery/games/9" <="" iframe="" height="900px">wddwdw</iframe>`,
		`< iframe src="http://m.647751.com/#/lottery/games/9" <="" iframe="" height="900px">wddwdw</iframe>`,
		`<img src="https://www.imageshop.com.tw/pic/shop/women2/56137/56137_430_65.jpg">`,
		`<a`,
		`</a>`,
		`<a href=""`,
		`<img`,
		`<script src='http://tw.yahoo.com />`,
		`<a onblur="alert(secret)" href="http://www.google.com">Google</a>`,
		`<script`,
	}

	for _, v := range testCases {
		t.Run(v, func(t *testing.T) {
			err := checkMessage(v)
			assert.Equal(t, errors.ErrIllegal, err)
		})
	}
}

func TestCheckMessagePass(t *testing.T) {
	testCases := []string{
		`測試`,
		`test`,
		`qqwqwwq`,
		`iframe>`,
		`iframe`,
		`/iframe`,
		`<`,
		`><`,
		`framework`,
		`img`,
		`< img`,
		`a`,
		`< a`,
		`< script src='http://tw.yahoo.com />`,
		`< iframe src="http://m.647751.com/#/lottery/games/9" <="" iframe="" height="900px">wddwdw< /iframe>`,
	}

	for _, v := range testCases {
		t.Run(v, func(t *testing.T) {
			err := checkMessage(v)
			assert.Nil(t, err)
		})
	}
}
