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

func Get(url string, data url.Values) Response {
	return response(get(url, data))
}

func PostJson(url string, body []byte) Response {
	return response(postJson(url, body))
}

func PostJsonNotToken(url string, body []byte) Response {
	return response(postJsonNotToken(url, body))
}

func PutJson(url string, body []byte) Response {
	return response(putJson(url, body))
}

func Delete(url string, data url.Values) Response {
	return response(deletes(url, data))
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
	fmt.Printf("post: %s\n", string(body))
	req, err := http.NewRequest("POST", url, bytes.NewBuffer(body))
	if err != nil {
		return nil, err
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer eyJ0eXAiOiJKV1QiLCJhbGciOiJIUzI1NiJ9.eyJpYXQiOjE1NTcyMTE2NTAsIm5iZiI6MTU1NzIxMTY1MCwiaXNzIjoibG9naW4iLCJzZXNzaW9uX3Rva2VuIjoiZjc2OTYyM2Y0YTNlNDE4MWE4NzAwYWNkYTE3NzE1MmIiLCJkYXRhIjp7InVpZCI6IjEyNTdlN2Q5ZTFjOTQ0ZWY5YTZmMTI5Y2I5NDk1ZDAyIiwidXNlcm5hbWUiOiJyb290In19.7VJxH3tQpnJqWTlPbId7f0Rt7eQoaVvaJmbWxtHTqRU")
	resp, err := httpClient.Do(req)
	if err != nil {
		return nil, err
	}

	return resp, nil
}

func postJsonNotToken(url string, body []byte) (*http.Response, error) {
	fmt.Printf("post: %s\n", string(body))
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

func putJson(url string, body []byte) (*http.Response, error) {
	fmt.Printf("post: %s\n", string(body))
	req, err := http.NewRequest("PUT", url, bytes.NewBuffer(body))
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

func deletes(url string, body url.Values) (*http.Response, error) {
	req, err := http.NewRequest("DELETE", fmt.Sprintf("%s?%s", url, body.Encode()), nil)
	if err != nil {
		return nil, err
	}

	resp, err := httpClient.Do(req)
	if err != nil {
		return nil, err
	}

	return resp, nil
}

func get(url string, body url.Values) (*http.Response, error) {
	u := body.Encode()
	if u != "" {
		url = fmt.Sprintf("%s?%s", url, u)
	}
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}

	resp, err := httpClient.Do(req)
	if err != nil {
		return nil, err
	}

	return resp, nil
}
