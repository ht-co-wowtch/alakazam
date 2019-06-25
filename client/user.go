package client

import (
	"encoding/json"
)

type User struct {
	Uid    string `json:"uid"`
	Name   string `json:"name"`
	Avatar string `json:"avatar"`
	Type   int    `json:"type"`
}

func (c *Client) Auth(token string) (auth User, err error) {
	b, err := c.get("/profile", nil, bearer(&Params{Token: token}))

	if err != nil {
		return auth, err
	}

	err = json.Unmarshal(b, &auth)

	return auth, err
}
