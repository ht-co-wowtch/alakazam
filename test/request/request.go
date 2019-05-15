package request

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"strings"
	"time"
)

const (
	host      = "http://127.0.0.1:3111"
	adminHost = "http://127.0.0.1:3112"
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

func Post(url string, data url.Values) Response {
	return response(post(url, data))
}

func PostJson(url string, body []byte) Response {
	return response(postJson(url, body))
}

func response(r *http.Response, err error) (re Response) {
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

func postJson(url string, body []byte) (*http.Response, error) {
	req, err := http.NewRequest("POST", url, bytes.NewBuffer(body))
	if err != nil {
		return nil, err
	}

	req.Header.Set("Content-Type", "application/json")
	resp, err := httpClient.Do(req)
	if err != nil {
		return nil, err
	}

	return resp, nil
}
