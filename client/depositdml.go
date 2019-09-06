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
	end := time.Now()
	start, err := time.Parse(time.RFC3339, end.Format("2006-01-02")+"T00:00:00+08:00")
	if err != nil {
		return Money{}, err
	}
	if day > 1 {
		start = start.AddDate(0, 0, -(day - 1))
	}

	q.Set("end_at", end.Format(time.RFC3339))
	q.Set("start_at", start.Format(time.RFC3339))
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
