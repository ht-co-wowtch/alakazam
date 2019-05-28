package client

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/stretchr/testify/assert"
	"gitlab.com/jetfueltw/cpw/alakazam/errors"
	"gitlab.com/jetfueltw/cpw/alakazam/logic/store"
	"io/ioutil"
	"net/http"
	"strings"
	"testing"
)

func TestGetUser(t *testing.T) {
	user := User{
		Uid:      "82ea16cd2d6a49d887440066ef739669",
		Nickname: "test",
		Type:     store.Player,
		Avatar:   "/",
		Token:    "eyJ0eXAiOiJKV1QiLCJhbGciOiJIUzI1NiJ9.eyJpYXQiOjE1NTg2ODgwMTcsImlzcyI6ImNwdyIsImF1ZCI6ImNoYXQiLCJzZXNzaW9uX3Rva2VuIjoiY2MwZGEwNjMwMzg2NGFjNWJlZGJhMzViNWQ1NWNkZTEiLCJ1aWQiOiI5ODQxNjQyNmU0OTQ0ZWUyODhkOTQ3NWNkODBiYzUwMSJ9.sfIKY2nZ6b4pWGrAmNUV8ndkQRmnv2fKdg80cW3FS9Y",
	}
	token := "ec2fa7acc9d443489531b156077c09a1"
	expectedURL := "/authentication"

	c := newMockClient(func(req *http.Request) (response *http.Response, e error) {
		if req.Method != "POST" {
			return nil, fmt.Errorf("expected POST method, got %s", req.Method)
		}
		if !strings.HasPrefix(req.URL.Path, expectedURL) {
			return nil, fmt.Errorf("Expected URL '%s', got '%s'", expectedURL, req.URL)
		}

		b, err := ioutil.ReadAll(req.Body)

		if err != nil {
			return nil, err
		}

		var p ticket
		if err := json.Unmarshal(b, &p); err != nil {
			return nil, err
		}

		if p.Ticket != token {
			return nil, fmt.Errorf("Body Ticket Not is %s", token)
		}

		header := http.Header{}
		header.Set(contentType, jsonHeaderType)

		b, err = json.Marshal(user)
		if err != nil {
			return nil, err
		}

		return &http.Response{
			StatusCode: http.StatusOK,
			Body:       ioutil.NopCloser(bytes.NewReader(b)),
			Header:     header,
		}, nil
	})

	a, err := c.Auth(token)

	assert.Nil(t, err)
	assert.Equal(t, user, a)
}

func TestGetUserNotFound(t *testing.T) {
	expected := errors.FailureError

	c := newMockClient(func(req *http.Request) (response *http.Response, e error) {
		header := http.Header{}
		header.Set(contentType, jsonHeaderType)

		b, err := json.Marshal(expected)
		if err != nil {
			return nil, err
		}

		return &http.Response{
			StatusCode: expected.Status,
			Body:       ioutil.NopCloser(bytes.NewReader(b)),
			Header:     header,
		}, nil
	})

	_, err := c.Auth("")

	assert.Equal(t, expected, err)
}
