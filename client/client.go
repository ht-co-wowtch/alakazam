package client

import (
	"encoding/json"
	"gitlab.com/jetfueltw/cpw/micro/client"
	"gitlab.com/jetfueltw/cpw/micro/errdefs"
	"net/http"
)

type Client struct {
	c *client.Client
}

func New(conf *client.Conf) *Client {
	return &Client{
		c: client.New(conf),
	}
}

func checkResponse(resp *http.Response) error {
	if resp.StatusCode-http.StatusOK > 100 {
		e := new(errdefs.Error)
		if err := json.NewDecoder(resp.Body).Decode(e); err != nil {
			return err
		}
		return e
	}
	return nil
}
