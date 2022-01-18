package client

import (
	"gitlab.com/ht-co/micro/client"
	"net/http"
)

type TransportFunc func(req *http.Request) (resp *http.Response, err error)

func (tf TransportFunc) RoundTrip(req *http.Request) (*http.Response, error) {
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
				Transport: TransportFunc(doer),
			},
		),
	}
}
