package client

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/stretchr/testify/assert"
	"gitlab.com/jetfueltw/cpw/alakazam/logic/store"
	"io/ioutil"
	"net/http"
	"strings"
	"testing"
)

func TestGetUser(t *testing.T) {
	user := User{
		Uid:    "82ea16cd2d6a49d887440066ef739669",
		Name:   "test",
		Gender: store.Female,
		Type:   store.Player,
	}
	token := "ec2fa7acc9d443489531b156077c09a1"
	expectedURL := "/tripartite/user/" + user.Uid + "/token/" + token

	c := newMockClient(func(req *http.Request) (response *http.Response, e error) {
		if req.Method != "GET" {
			return nil, fmt.Errorf("expected POST method, got %s", req.Method)
		}
		if !strings.HasPrefix(req.URL.Path, expectedURL) {
			return nil, fmt.Errorf("Expected URL '%s', got '%s'", expectedURL, req.URL)
		}

		header := http.Header{}
		header.Set(contentType, jsonHeaderType)

		b, err := json.Marshal(user)
		if err != nil {
			return nil, err
		}

		return &http.Response{
			StatusCode: http.StatusOK,
			Body:       ioutil.NopCloser(bytes.NewReader(b)),
			Header:     header,
		}, nil
	})

	a, err := c.GetUser(user.Uid, token)

	assert.Nil(t, err)
	assert.Equal(t, user, a)
}
