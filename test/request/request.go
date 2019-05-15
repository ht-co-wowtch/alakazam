package request

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"strings"
	"time"
)

var httpClient *http.Client

func init() {
	httpClient = &http.Client{
		Timeout: time.Second * 5,
	}
}

type Response struct {
	StatusCode int
	Error      error
	Body       []byte
}

func push(url string, data url.Values) (re Response) {
	r, err := post(url, data)
	if err != nil {
		return
	}

	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		return
	}

	re.Error = r.Body.Close()
	re.StatusCode = r.StatusCode
	re.Body = body
	fmt.Printf("response %s\n", string(body))
	return
}

func post(url string, body url.Values) (*http.Response, error) {
	req, err := http.NewRequest("POST", url, strings.NewReader(body.Encode()))
	if err != nil {
		return nil, err
	}

	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	resp, err := httpClient.Do(req)
	if err != nil {
		return nil, err
	}

	return resp, nil
}

const (
	host      = "http://127.0.0.1:3111"
	adminHost = "http://127.0.0.1:3112"
)

func PushRoom(uid, key, message string) Response {
	data := url.Values{}
	data.Set("uid", uid)
	data.Set("key", key)
	data.Set("message", message)
	return push(host+"/push/room", data)
}

func PushBroadcast(uid, key, message string, roomId []string, ) Response {
	data := url.Values{
		"room_id": roomId,
	}
	data.Set("uid", uid)
	data.Set("key", key)
	data.Set("message", message)
	return push(fmt.Sprintf(adminHost+"/push/all"), data)
}
