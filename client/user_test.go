package client

import (
	"encoding/json"
	"fmt"
	"github.com/stretchr/testify/assert"
	"gitlab.com/jetfueltw/cpw/alakazam/errors"
	"gitlab.com/jetfueltw/cpw/alakazam/logic/store"
	"net/http"
	"strings"
	"testing"
)

func TestAuth(t *testing.T) {
	user := User{
		Uid:    "82ea16cd2d6a49d887440066ef739669",
		Name:   "test",
		Type:   store.Player,
		Avatar: "/",
	}
	expectedPath := "/profile"

	c := newMockClient(func(req *http.Request) (response *http.Response, e error) {
		if req.Method != "GET" {
			return nil, fmt.Errorf("expected GET method, got %s", req.Method)
		}

		if req.URL.Path != expectedPath {
			return nil, fmt.Errorf("Expected url path '%s', got '%s'", expectedPath, req.URL.Path)
		}

		header := http.Header{}
		header.Set(contentType, jsonHeaderType)

		b, err := json.Marshal(user)

		if err != nil {
			return nil, err
		}

		return toResponse(http.StatusOK, b), nil
	})

	a, err := c.Auth("")

	assert.Nil(t, err)
	assert.Equal(t, user, a)
}

func TestAuthToken(t *testing.T) {
	expectedToken := "ec2fa7acc9d443489531b156077c09a1"

	c := newMockClient(func(req *http.Request) (response *http.Response, e error) {
		if err := checkAuthorization(req, expectedToken); err != nil {
			return nil, err
		}

		header := http.Header{}
		header.Set(contentType, jsonHeaderType)

		return toResponse(http.StatusOK, []byte(`{}`)), nil
	})

	_, err := c.Auth(expectedToken)

	assert.Nil(t, err)
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

		return toResponse(errors.FailureError.Status, b), nil
	})

	_, err := c.Auth("")

	assert.Equal(t, expected, err)
}

func checkAuthorization(request *http.Request, jwt string) error {
	authorization := request.Header.Get("Authorization")
	token := strings.Split(authorization, " ")

	if token[0] != "Bearer" {
		return fmt.Errorf("Authorization not Bearer")
	}

	if token[1] != jwt {
		return fmt.Errorf("Authorization not token")
	}

	return nil
}
