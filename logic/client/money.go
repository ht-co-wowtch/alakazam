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

// TODO 未完成
func (c *Client) GetMoney(uid string, day int) (money Money, err error) {
	if uid == "009422e667c146379b3aa69f336ad4e5" {
		return c.getMoney(uid, day)
	}
	return Money{0, 0}, nil
}

func (c *Client) getMoney(uid string, day int) (money Money, err error) {
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
