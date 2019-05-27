package client

import (
	"net/http"
)

type transportFunc func(*http.Request) (*http.Response, error)

func (tf transportFunc) RoundTrip(req *http.Request) (*http.Response, error) {
	return tf(req)
}

func newMockClient(doer func(*http.Request) (*http.Response, error)) *Client {
	return &Client{
		client: &http.Client{
			Transport: transportFunc(doer),
		},
	}
}
