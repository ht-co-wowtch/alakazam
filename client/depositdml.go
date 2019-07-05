package client

import (
	"encoding/json"
	"net/url"
	"time"
)

type Money struct {
	Dml     float64 `json:"dml"`
	Deposit float64 `json:"deposit"`
}

func (c *Client) GetDepositAndDml(day int, uid, token string) (money Money, err error) {
	q := url.Values{}
	now := time.Now()
	q.Set("start_at", now.AddDate(0, 0, -day).Format("2006-01-02T00:00:00Z07:00"))
	q.Set("end_at", now.Format(time.RFC3339))
	resp, err := c.c.Get("/deposit-dml", q, bearer(token))
	if err != nil {
		return Money{}, err
	}

	defer resp.Body.Close()
	if err := checkResponse(resp); err != nil {
		return Money{}, err
	}
	var u Money
	if err := json.NewDecoder(resp.Body).Decode(&u); err != nil {
		return Money{}, err
	}
	return u, nil
}
