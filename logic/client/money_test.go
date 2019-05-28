package client

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/stretchr/testify/assert"
	"io/ioutil"
	"net/http"
	"net/url"
	"strings"
	"testing"
	"time"
)

func TestGetMoney(t *testing.T) {
	expected := Money{
		Dml:     100,
		Deposit: 101,
	}

	now := time.Now()
	day := 1
	uid := "1"

	query := url.Values{}
	query.Set("start_at", now.AddDate(0, 0, -day).Format("2006-01-02T00:00:00Z07:00"))
	query.Set("end_at", now.Format(time.RFC3339))

	expectedURL := fmt.Sprintf("/members/%s/deposit-dml", uid)

	c := newMockClient(func(request *http.Request) (response *http.Response, e error) {
		if request.Method != "GET" {
			return nil, fmt.Errorf("expected GET method, got %s", request.Method)
		}

		if !strings.HasPrefix(request.URL.Path, expectedURL) {
			return nil, fmt.Errorf("Expected URL '%s', got '%s'", expectedURL, request.URL.Path)
		}

		if query.Encode() != request.URL.RawQuery {
			return nil, fmt.Errorf("Expected URL query '%s', got '%s'", query.Encode(), request.URL.RawQuery)
		}

		b, err := json.Marshal(expected)

		if err != nil {
			return nil, err
		}

		header := http.Header{}
		header.Set(contentType, jsonHeaderType)

		return &http.Response{
			StatusCode: http.StatusOK,
			Body:       ioutil.NopCloser(bytes.NewReader(b)),
			Header:     header,
		}, nil
	})

	actual, err := c.GetMoney(uid, day)

	assert.Nil(t, err)
	assert.Equal(t, expected, actual)
}
