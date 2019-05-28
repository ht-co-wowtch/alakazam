package client

import (
	"encoding/json"
	"io/ioutil"
	"net/http"
)

type Money struct {
	Dml    int `json:"dml"`
	Amount int `json:"amount"`
}

// TODO 先行實作等待三方接口文件
func (c *Client) GetMoney(uid string, day int) (money Money, err error) {
	req, err := http.NewRequest("GET", "/user/money", nil)
	if err != nil {
		return money, err
	}

	resp, err := c.client.Do(req)

	if err != nil {
		return money, err
	}

	defer resp.Body.Close()

	b, _ := ioutil.ReadAll(resp.Body)
	err = json.Unmarshal(b, &money)
	return money, err
}
