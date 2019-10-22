package client

import (
	"gitlab.com/jetfueltw/cpw/micro/client"
	"net/http"
)

type transportFunc func(req *http.Request) (resp *http.Response, err error)

func (tf transportFunc) RoundTrip(req *http.Request) (*http.Response, error) {
	return tf(req)
}

func NewMockClient(doer func(req *http.Request) (resp *http.Response, err error)) *Client {
	return &Client{
		c: client.Create(
			&client.Conf{
				Scheme: "http",
				Host:   "127.0.0.1",
			},
			&http.Client{
				Transport: transportFunc(doer),
			},
		),
	}
}
