package client

import (
	"encoding/json"
	"net/url"
	"time"
)

type Money struct {
	Dml     float64 `json:"total_dml"`
	Deposit float64 `json:"total_deposit_amount"`
}

func (c *Client) GetDepositAndDml(day int, uid, token string) (money Money, err error) {
	q := url.Values{}
	now := time.Now()
	q.Set("end_at", now.AddDate(0, 0, day).Format(time.RFC3339))
	q.Set("start_at", now.Format(time.RFC3339))
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
