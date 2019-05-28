package client

import (
	"net/http"
)

func newMockClient(doer func(*http.Request) (*http.Response, error)) *Client {
	return &Client{
		client: &http.Client{
			Transport: transportFunc(doer),
		},
	}
}
