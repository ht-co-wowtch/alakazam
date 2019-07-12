package client

import (
	"encoding/json"
)

type User struct {
	Uid    string `json:"uid"`
	Name   string `json:"name"`
	Avatar string `json:"avatar"`
	Type   string `json:"type"`
}

func (c *Client) Auth(token string) (User, error) {
	resp, err := c.c.Get("/profile", nil, bearer(token))
	if err != nil {
		return User{}, err
	}

	defer resp.Body.Close()
	if err := checkResponse(resp); err != nil {
		return User{}, err
	}
	var u User
	if err := json.NewDecoder(resp.Body).Decode(&u); err != nil {
		return User{}, err
	}
	return u, nil
}
