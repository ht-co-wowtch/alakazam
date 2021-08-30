package client

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/prometheus/common/log"
	"time"

	// "runtime/pprof"
)

type User struct {
	Uid    string `json:"uid"`
	Name   string `json:"name"`
	Avatar string `json:"avatar"`
	Type   int    `json:"type"`
	Gender int32  `json:"gender"`
}

type UserLiveExpire struct {
	Expire time.Time `json:"live_expire"`
}

var (
	errNoMember = errors.New("member not not found")
)

func (c *Client) Auth(token string) (User, error) {
	resp, err := c.c.Get("/profile", nil, bearer(token))
	if err != nil {
		return User{}, err
	}

	defer resp.Body.Close()
	if err := checkResponse(resp); err != nil {
		return User{}, err
	}
	var u User
	if err := json.NewDecoder(resp.Body).Decode(&u); err != nil {
		return User{}, err
	}
	if u.Uid == "" {
		return u, errNoMember
	}
	return u, nil
}

// TODO
// 取得會員月卡效期
func (c *Client) LiveExpire(uid string) (UserLiveExpire, error) {
	path := fmt.Sprintf("/live/expire/%s", uid)

	// 加上快取
	resp, _ := c.c.Get(path, nil, nil)

	defer resp.Body.Close()

	var ule UserLiveExpire

	if err := json.NewDecoder(resp.Body).Decode(&ule); err != nil {
		log.Error("get member expire log:, %o", err)
		return UserLiveExpire{}, err
	}

	return ule, nil
}
