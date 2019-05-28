package client

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/google/uuid"
	"gitlab.com/jetfueltw/cpw/alakazam/logic/store"
	"io/ioutil"
	"net/http"
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
	default:
		response = &http.Response{
			StatusCode: http.StatusBadRequest,
		}
	}
	return response, err
}

func mockPost(request *http.Request) (response *http.Response, err error) {
	switch request.URL.Path {
	case "/authentication":
		response, err = mockAuth(request)
	default:
		response = &http.Response{
			StatusCode: http.StatusBadRequest,
		}
	}
	return response, err
}

func mockAuth(request *http.Request) (*http.Response, error) {
	b, err := ioutil.ReadAll(request.Body)
	if err != nil {
		return nil, err
	}

	var p ticket
	if err := json.Unmarshal(b, &p); err != nil {
		return nil, err
	}

	if p.Ticket == "" {
		return nil, fmt.Errorf("Ticket not found")
	}

	header := http.Header{}
	header.Set(contentType, jsonHeaderType)

	uuid, _ := uuid.New().MarshalBinary()

	u := User{
		Uid:      fmt.Sprintf("%x", uuid),
		Nickname: fmt.Sprintf("test%d", time.Now().Unix()),
		Type:     store.Player,
		Avatar:   "https://via.placeholder.com/30x30",
		Token:    "eyJ0eXAiOiJKV1QiLCJhbGciOiJIUzI1NiJ9.eyJpYXQiOjE1NTg2ODgwMTcsImlzcyI6ImNwdyIsImF1ZCI6ImNoYXQiLCJzZXNzaW9uX3Rva2VuIjoiY2MwZGEwNjMwMzg2NGFjNWJlZGJhMzViNWQ1NWNkZTEiLCJ1aWQiOiI5ODQxNjQyNmU0OTQ0ZWUyODhkOTQ3NWNkODBiYzUwMSJ9.sfIKY2nZ6b4pWGrAmNUV8ndkQRmnv2fKdg80cW3FS9Y",
	}

	b, err = json.Marshal(u)

	if err != nil {
		return nil, err
	}

	return &http.Response{
		StatusCode: http.StatusOK,
		Body:       ioutil.NopCloser(bytes.NewReader(b)),
		Header:     header,
	}, nil
}

func mockGetMoney(request *http.Request) (*http.Response, error) {
	header := http.Header{}
	header.Set(contentType, jsonHeaderType)
	m := Money{0, 0}
	b, _ := json.Marshal(m)
	return &http.Response{
		StatusCode: http.StatusOK,
		Body:       ioutil.NopCloser(bytes.NewReader(b)),
		Header:     header,
	}, nil
}
