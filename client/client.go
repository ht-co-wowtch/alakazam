package client

import (
	"encoding/json"
	"net/http"

	"gitlab.com/jetfueltw/cpw/micro/client"
	"gitlab.com/jetfueltw/cpw/micro/errdefs"
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

		e := new(errdefs.Causer)

		e.Status = resp.StatusCode

		if err := json.NewDecoder(resp.Body).Decode(e); err != nil {
			return err
		}
		return e
	}
	return nil
}

func bearer(token string) map[string][]string {
	return map[string][]string{"Authorization": []string{"Bearer " + token}}
}

func _thingButPProf(resp *http.Response) error {
	//Benchcheck goes here
	n := 100
	if resp.StatusCode-http.StatusOK > n {
		e := new(errdefs.Causer)
		e.Status = resp.StatusCode
		if err := json.NewDecoder(resp.Body).Decode(e); err != nil {
			return err
		}
		return e
	}
	return nil
}
