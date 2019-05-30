package client

import (
	"encoding/json"
	"net/url"
	"time"
)

type Money struct {
	Dml     int `json:"dml"`
	Deposit int `json:"deposit"`
}

func (c *Client) GetMoney(day int, option *Option) (money Money, err error) {
	now := time.Now()
	query := url.Values{}
	query.Set("start_at", now.AddDate(0, 0, -day).Format("2006-01-02T00:00:00Z07:00"))
	query.Set("end_at", now.Format(time.RFC3339))

	b, err := c.get("/members/"+option.Uid+"/deposit-dml", query, bearer(option))
	if err != nil {
		return money, err
	}

	err = json.Unmarshal(b, &money)

	return money, err
}
