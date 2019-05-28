package client

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"time"
)

type Money struct {
	Dml     int `json:"dml"`
	Deposit int `json:"deposit"`
}

func (c *Client) GetMoney(uid string, day int) (money Money, err error) {
	now := time.Now()
	query := url.Values{}
	query.Set("start_at", now.AddDate(0, 0, -day).Format("2006-01-02T00:00:00Z07:00"))
	query.Set("end_at", now.Format(time.RFC3339))

	p := (&url.URL{Path: fmt.Sprintf(c.host+"/members/%s/deposit-dml", uid), RawQuery: query.Encode()}).String()

	req, err := http.NewRequest("GET", p, nil)
	if err != nil {
		return money, err
	}

	resp, err := c.client.Do(req)
	if err != nil {
		return money, err
	}

	defer resp.Body.Close()

	b, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return money, err
	}

	return money, json.Unmarshal(b, &money)
}
