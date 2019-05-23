package client

import (
	"gitlab.com/jetfueltw/cpw/alakazam/logic/conf"
	"net"
	"net/http"
	"time"
)

var (
	contentType = "Content-Type"

	jsonHeaderType = "application/json"
)

type Client struct {
	host string

	client *http.Client
}

func New(c *conf.Api) *Client {
	return Create(c, hTTPClient())
}

func Create(c *conf.Api, client *http.Client) *Client {
	return &Client{
		host:   c.Host,
		client: client,
	}
}

func hTTPClient() *http.Client {
	return &http.Client{
		Transport: &http.Transport{
			Proxy: http.ProxyFromEnvironment,
			DialContext: (&net.Dialer{
				// 連線timeout
				Timeout: 5 * time.Second,
				// 心跳週期
				KeepAlive: time.Minute,
			}).DialContext,
			// 全部host連線上限
			MaxConnsPerHost: 400,
			// 單一host連線閒置數
			MaxIdleConns: 200,
			// 全部host連線閒置數
			MaxIdleConnsPerHost: 200,
			// 單一連線最多閒置多久
			IdleConnTimeout: 3 * time.Minute,
			// tls交握timeout
			TLSHandshakeTimeout:   10 * time.Second,
			ExpectContinueTimeout: 1 * time.Second,
		},
		// 整個Request timeout
		Timeout: 10 * time.Second,
	}
}
