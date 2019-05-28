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
		Uid: "82ea16cd2d6a49d887440066ef739669",
		Data: Claims{
			UserName: "test",
			Type:     store.Player,
			Avatar:   "/",
		},
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

	a, err := c.GetUser(token)

	assert.Nil(t, err)
	assert.Equal(t, user, a)
}

func TestGetUserNotFound(t *testing.T) {
	c := newMockClient(func(req *http.Request) (response *http.Response, e error) {
		header := http.Header{}
		header.Set(contentType, jsonHeaderType)

		Err := errors.Error{
			Code:    15024010,
			Message: "Invalid ticket",
		}

		b, err := json.Marshal(Err)
		if err != nil {
			return nil, err
		}

		return &http.Response{
			StatusCode: http.StatusNotFound,
			Body:       ioutil.NopCloser(bytes.NewReader(b)),
			Header:     header,
		}, nil
	})

	_, err := c.GetUser("")

	assert.Equal(t, fmt.Errorf(`response is error({"code":15024010,"message":"Invalid ticket"})`), err)
}
