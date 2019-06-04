package front

import (
	"encoding/json"
	"gitlab.com/jetfueltw/cpw/alakazam/client"
	"gitlab.com/jetfueltw/cpw/alakazam/test/internal/request"
	"gitlab.com/jetfueltw/cpw/alakazam/test/internal/run"
	"net/http"
	"testing"
)

func mockDepositAndDmlApi(t *testing.T, a *request.Auth, deposit, dml int) {
	run.AddClient("/members/"+a.Uid+"/deposit-dml", func(res *http.Request) (response *http.Response, e error) {
		m := client.Money{
			Deposit: deposit,
			Dml:     dml,
		}

		b, err := json.Marshal(m)

		if err != nil {
			t.Fatalf("json.Marshal error(%v)", err)
		}

		return request.ToResponse(b, http.StatusOK)
	})
}
