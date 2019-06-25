package client

import (
	"bytes"
	"encoding/json"
	"io"
	"io/ioutil"
	"net/http"
	"net/url"
)

var (
	contentType = "Content-Type"

	jsonHeaderType = "application/json"
)

func NewRequest(client *http.Client, method, path string, query url.Values, body io.Reader, headers headers) ([]byte, error) {
	req, err := buildRequest(method, GetAPIPath(path, query), body, headers)
	if err != nil {
		return nil, err
	}

	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}

	return ioutil.ReadAll(resp.Body)
}

func (cli *Client) Get(path string, query url.Values, headers map[string][]string) (*http.Response, error) {
	return cli.sendRequest("GET", path, query, nil, headers)
}

func (cli *Client) PostJson(path string, query url.Values, obj interface{}, headers map[string][]string) (*http.Response, error) {
	body, headers, err := encodeBody(obj, headers)
	if err != nil {
		return nil, err
	}
	return cli.sendRequest("POST", path, query, body, headers)
}

func (cli *Client) Put(path string, query url.Values, body io.Reader, headers map[string][]string) (*http.Response, error) {
	return cli.sendRequest("PUT", path, query, body, headers)
}

func (cli *Client) PutJson(path string, query url.Values, obj interface{}, headers map[string][]string) (*http.Response, error) {
	body, headers, err := encodeBody(obj, headers)
	if err != nil {
		return nil, err
	}
	return cli.sendRequest("PUT", path, query, body, headers)
}

func (cli *Client) Delete(path string, query url.Values, headers map[string][]string) (*http.Response, error) {
	return cli.sendRequest("DELETE", path, query, nil, headers)
}

func (cli *Client) sendRequest(method, path string, query url.Values, body io.Reader, headers headers) (*http.Response, error) {
	req, err := buildRequest(method, GetAPIPath(path, query), body, headers)
	if err != nil {
		return nil, err
	}

	req.URL.Host = cli.host
	req.URL.Scheme = cli.scheme

	return cli.client.Do(req)
}

func GetAPIPath(p string, query url.Values) string {
	return (&url.URL{Path: p, RawQuery: query.Encode()}).String()
}

func buildRequest(method, path string, body io.Reader, headers headers) (*http.Request, error) {
	req, err := http.NewRequest(method, path, body)
	if err != nil {
		return nil, err
	}

	req = addHeaders(req, headers)

	return req, nil
}

func addHeaders(req *http.Request, headers headers) *http.Request {
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

	body, err := EncodeJson(obj)
	if err != nil {
		return nil, headers, err
	}
	if headers == nil {
		headers = make(map[string][]string)
	}
	headers[contentType] = []string{jsonHeaderType}
	return body, headers, nil
}

func EncodeJson(data interface{}) (*bytes.Buffer, error) {
	params := bytes.NewBuffer(nil)
	if data != nil {
		if err := json.NewEncoder(params).Encode(data); err != nil {
			return nil, err
		}
	}
	return params, nil
}
