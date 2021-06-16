package client

import (
	"encoding/json"
	"errors"
	// "runtime/pprof"
)

type User struct {
	Uid    string `json:"uid"`
	Name   string `json:"name"`
	Avatar string `json:"avatar"`
	Type   int    `json:"type"`
	Gender int32  `json:"gender"`
}

var (
	errNoMember = errors.New("member not not found")
)

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
	if u.Uid == "" {
		return u, errNoMember
	}
	return u, nil
}
