package client

import (
	"encoding/json"
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
	b, err := c.post("/authentication", nil, ticket{Ticket: token}, nil)

	if err != nil {
		return auth, err
	}

	err = json.Unmarshal(b, &auth)

	return auth, err
}
