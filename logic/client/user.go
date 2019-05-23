package client

import (
	"encoding/json"
	"io/ioutil"
	"net/http"
)

type User struct {
	Uid    string `json:"uid"`
	Name   string `json:"name"`
	Gender string `json:"gender"`
	Type   string `json:"type"`
}

func (c *Client) GetUser(token string) (auth User, err error) {
	req, err := http.NewRequest("GET", "/game/user/"+token, nil)
	if err != nil {
		return auth, err
	}

	resp, err := c.client.Do(req)

	if err != nil {
		return auth, err
	}

	defer resp.Body.Close()

	b, _ := ioutil.ReadAll(resp.Body)

	json.Unmarshal(b, &auth)

	return auth, nil
}
