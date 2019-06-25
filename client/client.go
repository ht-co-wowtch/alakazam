package client

import (
	"bytes"
	"encoding/json"
	"gitlab.com/jetfueltw/cpw/alakazam/errors"
	"gitlab.com/jetfueltw/cpw/alakazam/logic/conf"
	"io"
	"io/ioutil"
	"net"
	"net/http"
	"net/url"
	"time"
)

var (
	contentType = "Content-Type"

	jsonHeaderType = "application/json"
)

type Client struct {
	host string

	scheme string

	client *http.Client
}

func New(c *conf.Api) *Client {
	return Create(c, hTTPClient())
}

func Create(c *conf.Api, client *http.Client) *Client {
	return &Client{
		scheme: "http",
		host:   "127.0.0.1:3005",
		client: client,
	}
}

func hTTPClient() *http.Client {
	return &http.Client{
		Transport: &http.Transport{
			Proxy: http.ProxyFromEnvironment,
			DialContext: (&net.Dialer{
				// 連線timeout
				Timeout: 5 * time.Second,
				// 心跳週期
				KeepAlive: time.Minute,
			}).DialContext,
			// 全部host連線上限
			MaxConnsPerHost: 400,
			// 單一host連線閒置數
			MaxIdleConns: 200,
			// 全部host連線閒置數
			MaxIdleConnsPerHost: 200,
			// 單一連線最多閒置多久
			IdleConnTimeout: 3 * time.Minute,
			// tls交握timeout
			TLSHandshakeTimeout:   10 * time.Second,
			ExpectContinueTimeout: 1 * time.Second,
		},
		// 整個Request timeout
		Timeout: 10 * time.Second,
	}
}

type Params struct {
	// user uid
	Uid string

	// 三方應用接口jwt
	Token string
}

func (cli *Client) get(path string, query url.Values, headers map[string][]string) ([]byte, error) {
	return cli.sendRequest("GET", path, query, nil, headers)
}

func (cli *Client) post(path string, query url.Values, obj interface{}, headers map[string][]string) ([]byte, error) {
	body, headers, err := encodeBody(obj, headers)
	if err != nil {
		return nil, err
	}
	return cli.sendRequest("POST", path, query, body, headers)
}

func (cli *Client) sendRequest(method, path string, query url.Values, body io.Reader, headers headers) ([]byte, error) {
	req, err := cli.buildRequest(method, cli.getAPIPath(path, query), body, headers)
	if err != nil {
		return nil, err
	}

	resp, err := cli.client.Do(req)
	if err != nil {
		return nil, err
	}

	return cli.checkResponseErr(resp)
}

func (cli *Client) checkResponseErr(response *http.Response) ([]byte, error) {
	if response.StatusCode >= 200 && response.StatusCode < 400 {
		return ioutil.ReadAll(response.Body)
	}

	var e errors.Error
	b, err := ioutil.ReadAll(response.Body)

	if err != nil {
		return nil, err
	}

	if err := json.Unmarshal(b, &e); err != nil {
		return nil, err
	}

	e.Status = response.StatusCode

	return nil, e
}

func (cli *Client) getAPIPath(p string, query url.Values) string {
	return (&url.URL{Path: p, RawQuery: query.Encode()}).String()
}

func bearer(option *Params) map[string][]string {
	return map[string][]string{
		"Authorization": []string{"Bearer " + option.Token},
	}
}

func (cli *Client) buildRequest(method, path string, body io.Reader, headers headers) (*http.Request, error) {
	req, err := http.NewRequest(method, path, body)
	if err != nil {
		return nil, err
	}
	req = cli.addHeaders(req, headers)

	req.URL.Host = cli.host
	req.URL.Scheme = cli.scheme
	return req, nil
}

func (cli *Client) addHeaders(req *http.Request, headers headers) *http.Request {
	if headers != nil {
		for k, v := range headers {
			req.Header[k] = v
		}
	}
	return req
}

type headers map[string][]string

func encodeBody(obj interface{}, headers headers) (io.Reader, headers, error) {
	if obj == nil {
		return nil, headers, nil
	}

	body, err := encodeData(obj)
	if err != nil {
		return nil, headers, err
	}
	if headers == nil {
		headers = make(map[string][]string)
	}
	headers[contentType] = []string{jsonHeaderType}
	return body, headers, nil
}

func encodeData(data interface{}) (*bytes.Buffer, error) {
	params := bytes.NewBuffer(nil)
	if data != nil {
		if err := json.NewEncoder(params).Encode(data); err != nil {
			return nil, err
		}
	}
	return params, nil
}
