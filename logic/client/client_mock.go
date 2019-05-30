package client

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/google/uuid"
	"gitlab.com/jetfueltw/cpw/alakazam/logic/store"
	"io/ioutil"
	"net/http"
	"strings"
	"time"
)

func mockHTTPClient() *http.Client {
	return &http.Client{
		Transport: transportFunc(func(request *http.Request) (response *http.Response, err error) {
			switch request.Method {
			case "GET":
				response, err = mockGet(request)
			case "POST":
				response, err = mockPost(request)
			}
			return response, err
		}),
		// 整個Request timeout
		Timeout: 10 * time.Second,
	}
}

type transportFunc func(*http.Request) (*http.Response, error)

func (tf transportFunc) RoundTrip(req *http.Request) (*http.Response, error) {
	return tf(req)
}

func mockGet(request *http.Request) (response *http.Response, err error) {
	switch request.URL.Path {
	case "/user/money":
		response, err = mockGetMoney(request)
	case "/profile":
		response, err = mockAuth(request)
	default:
		response = &http.Response{
			StatusCode: http.StatusBadRequest,
		}
	}
	return response, err
}

func mockPost(request *http.Request) (response *http.Response, err error) {
	switch request.URL.Path {
	default:
		response = &http.Response{
			StatusCode: http.StatusBadRequest,
		}
	}
	return response, err
}

func mockAuth(request *http.Request) (*http.Response, error) {
	authorization := request.Header.Get("Authorization")
	token := strings.Split(authorization, " ")

	if token[0] != "Bearer" {
		return nil, fmt.Errorf("Authorization not Bearer")
	}

	if token[1] == "" {
		return nil, fmt.Errorf("Authorization not token")
	}

	header := http.Header{}
	header.Set(contentType, jsonHeaderType)

	uuid, _ := uuid.New().MarshalBinary()

	u := User{
		Uid:      fmt.Sprintf("%x", uuid),
		Nickname: fmt.Sprintf("test%d", time.Now().Unix()),
		Type:     store.Player,
		Avatar:   "https://via.placeholder.com/30x30",
	}

	b, err := json.Marshal(u)

	if err != nil {
		return nil, err
	}

	return toResponse(http.StatusOK, b), nil
}

func mockGetMoney(request *http.Request) (*http.Response, error) {
	header := http.Header{}
	header.Set(contentType, jsonHeaderType)
	m := Money{0, 0}
	b, _ := json.Marshal(m)
	return toResponse(http.StatusOK, b), nil
}

func toResponse(statusCode int, body []byte) *http.Response {
	header := http.Header{}
	header.Set(contentType, jsonHeaderType)

	return &http.Response{
		StatusCode: statusCode,
		Body:       ioutil.NopCloser(bytes.NewReader(body)),
		Header:     header,
	}
}
