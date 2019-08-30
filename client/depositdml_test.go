package client

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/stretchr/testify/assert"
	"gitlab.com/jetfueltw/cpw/alakazam/errors"
	"io/ioutil"
	"net/http"
	"testing"
	"time"
)

func TestGetDepositAndDmlByTime(t *testing.T) {
	c := newMockClient(func(req *http.Request) (resp *http.Response, err error) {
		query := req.URL.Query()

		start, err := time.Parse(time.RFC3339, query.Get("start_at"))
		if err != nil {
			return nil, fmt.Errorf("start_at time parse %s", err.Error())
		}

		end, err := time.Parse(time.RFC3339, query.Get("end_at"))
		if err != nil {
			return nil, fmt.Errorf("end_at time parse %s", err.Error())
		}

		if end.Sub(start).Hours() != 24 {
			return nil, errors.New("not one day")
		}

		body, err := json.Marshal(Money{})
		if err != nil {
			return nil, err
		}
		return &http.Response{
			StatusCode: http.StatusOK,
			Body:       ioutil.NopCloser(bytes.NewBuffer(body)),
		}, nil
	})

	_, err := c.GetDepositAndDml(1, "", "")
	assert.Nil(t, err)
}
