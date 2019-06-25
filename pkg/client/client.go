package client

import (
	"net"
	"net/http"
	"time"
)

type Client struct {
	host string

	scheme string

	client *http.Client
}

func New(c *Conf) *Client {
	return Create(c, hTTPClient(c))
}

func Create(c *Conf, client *http.Client) *Client {
	return &Client{
		host:   c.Host,
		scheme: c.Scheme,
		client: client,
	}
}

func hTTPClient(c *Conf) *http.Client {
	return &http.Client{
		Transport: &http.Transport{
			Proxy: http.ProxyFromEnvironment,
			DialContext: (&net.Dialer{
				// 連線timeout
				Timeout: 5 * time.Second,
				// 心跳週期
				KeepAlive: time.Minute,
			}).DialContext,

			// 對每個host的最大連接數量
			MaxConnsPerHost: c.MaxConns,

			// 所有host的連接池最大空閒連接數量
			MaxIdleConns: c.MaxIdleConns,

			// 每個host的連接池最大空閒連接數
			MaxIdleConnsPerHost: c.MaxIdleConns,

			// 單一連線最多閒置多久
			IdleConnTimeout: c.IdleConnTimeout,

			// tls交握timeout
			TLSHandshakeTimeout:   10 * time.Second,
			ExpectContinueTimeout: 1 * time.Second,
		},

		// 整個Request timeout
		Timeout: 10 * time.Second,
	}
}
