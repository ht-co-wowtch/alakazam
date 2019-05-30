package client

import (
	"encoding/json"
)

type User struct {
	Uid      string `json:"uid"`
	Nickname string `json:"nickname"`
	Avatar   string `json:"avatar"`
	Type     string `json:"type"`
}

func (c *Client) Auth(token string) (auth User, err error) {
	b, err := c.get("/profile", nil, bearer(&Params{Token: token}))

	if err != nil {
		return auth, err
	}

	err = json.Unmarshal(b, &auth)

	return auth, err
}
