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
	expectedToken := "eyJ0eXAiOiJKV1QiLCJhbGciOiJIUzI1NiJ9.eyJpYXQiOjE1NTg2ODgwMTcsImlzcyI6ImNwdyIsImF1ZCI6ImNoYXQiLCJzZXNzaW9uX3Rva2VuIjoiY2MwZGEwNjMwMzg2NGFjNWJlZGJhMzViNWQ1NWNkZTEiLCJ1aWQiOiI5ODQxNjQyNmU0OTQ0ZWUyODhkOTQ3NWNkODBiYzUwMSJ9.sfIKY2nZ6b4pWGrAmNUV8ndkQRmnv2fKdg80cW3FS9Y"

	c := newMockClient(func(request *http.Request) (response *http.Response, e error) {
		if request.Method != "GET" {
			return nil, fmt.Errorf("expected GET method, got %s", request.Method)
		}

		if request.URL.Path != expectedURL {
			return nil, fmt.Errorf("Expected URL '%s', got '%s'", expectedURL, request.URL.Path)
		}

		if query.Encode() != request.URL.RawQuery {
			return nil, fmt.Errorf("Expected URL query '%s', got '%s'", query.Encode(), request.URL.RawQuery)
		}

		authorization := request.Header.Get("Authorization")
		token := strings.Split(authorization, " ")

		if token[0] != "Bearer" {
			return nil, fmt.Errorf("Authorization not Bearer")
		}

		if token[1] != expectedToken {
			return nil, fmt.Errorf("Expected Authorization Bearer token '%s', got '%s'", expectedToken, token[1])
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

	actual, err := c.GetDepositAndDml(day, &Params{Uid: uid, Token: expectedToken})

	assert.Nil(t, err)
	assert.Equal(t, expected, actual)
}
