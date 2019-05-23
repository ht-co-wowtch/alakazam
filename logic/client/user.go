package client

import (
	"encoding/json"
	"gitlab.com/jetfueltw/cpw/alakazam/errors"
	"io/ioutil"
	"net/http"
)

type User struct {
	Uid    string `json:"uid"`
	Name   string `json:"name"`
	Gender string `json:"gender"`
	Type   string `json:"type"`
}

func (c *Client) GetUser(uid, token string) (auth User, err error) {
	req, err := http.NewRequest("GET", c.host+"/tripartite/user/"+uid+"/token/"+token, nil)
	if err != nil {
		return auth, err
	}

	resp, err := c.client.Do(req)

	if err != nil {
		return auth, err
	}

	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return auth, errors.UserError
	}

	b, _ := ioutil.ReadAll(resp.Body)
	err = json.Unmarshal(b, &auth)

	return auth, err
}
