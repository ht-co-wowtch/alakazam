package client

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
)

type User struct {
	Uid  string `json:"uid"`
	Data Claims `json:"Claims"`
}

type Claims struct {
	UserName string `json:"username"`
	Type     int    `json:"type"`
	Avatar   string `json:"avatar"`
}

type ticket struct {
	Ticket string `json:"ticket"`
}

func (c *Client) GetUser(token string) (auth User, err error) {
	var t ticket
	t.Ticket = token

	b, err := json.Marshal(t)
	if err != nil {
		return auth, err
	}

	req, err := http.NewRequest("POST", c.host+"/authentication", bytes.NewBuffer(b))
	if err != nil {
		return auth, err
	}

	req.Header.Set("Content-Type", "application/json")
	resp, err := c.client.Do(req)

	if err != nil {
		return auth, err
	}

	defer resp.Body.Close()

	b, err = ioutil.ReadAll(resp.Body)

	if err != nil {
		return auth, err
	}

	if resp.StatusCode != http.StatusOK {
		return auth, fmt.Errorf("response is error(%s)", string(b))
	}

	return auth, json.Unmarshal(b, &auth)
}
