package client

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/stretchr/testify/assert"
	"gitlab.com/jetfueltw/cpw/alakazam/errors"
	"io/ioutil"
	"net/http"
	"net/url"
	"testing"
)

func newMockClient(doer func(*http.Request) (*http.Response, error)) *Client {
	return &Client{
		client: &http.Client{
			Transport: transportFunc(doer),
		},
	}
}

func TestUrl(t *testing.T) {
	expectedScheme := "http"
	expectedHost := "127.0.0.1:8080"
	c := givenClient(t, "http://127.0.0.1:8080", func(request *http.Request) (response *http.Response, e error) {
		if request.URL.Scheme != expectedScheme {
			return nil, fmt.Errorf("expected url scheme %s, got %s", expectedScheme, request.URL.Scheme)
		}
		if request.URL.Host != expectedHost {
			return nil, fmt.Errorf("expected url host %s, got %s", expectedHost, request.URL.Host)
		}
		return &http.Response{
			StatusCode: http.StatusOK,
			Body:       ioutil.NopCloser(bytes.NewReader([]byte(``))),
		}, nil
	})

	b, err := c.get("/test", nil, nil)

	assert.Nil(t, err)
	assert.Empty(t, b)
}

func TestGet(t *testing.T) {
	expectedURL := "/test"
	query := url.Values{}
	query.Set("id", "1")

	c := givenClient(t, "http://127.0.0.1:8080", func(request *http.Request) (response *http.Response, e error) {
		if request.URL.Path != expectedURL {
			return nil, fmt.Errorf("Expected URL '%s', got '%s'", expectedURL, request.URL.Path)
		}
		if request.URL.RawQuery != query.Encode() {
			return nil, fmt.Errorf("Expected Raw Query '%s', got '%s'", query.Encode(), request.URL.RawQuery)
		}
		return &http.Response{
			StatusCode: http.StatusOK,
			Body:       ioutil.NopCloser(bytes.NewReader([]byte(``))),
		}, nil
	})

	b, err := c.get(expectedURL, query, nil)

	assert.Nil(t, err)
	assert.Empty(t, b)
}

type testJson struct {
	Message string `json:"test"`
}

func TestPost(t *testing.T) {
	expectedMessage := "test"
	expectedURL := "/test"
	expectedHeader := "application/json"
	expectedAuthorization := "token"

	c := givenClient(t, "http://127.0.0.1:8080", func(request *http.Request) (response *http.Response, e error) {
		if request.URL.Path != expectedURL {
			return nil, fmt.Errorf("Expected URL '%s', got '%s'", expectedURL, request.URL.Path)
		}

		if header := request.Header.Get("Content-Type"); header != expectedHeader {
			return nil, fmt.Errorf("Expected Content-Type '%s', got '%s'", expectedHeader, header)
		}

		if authorization := request.Header.Get("Authorization"); authorization != expectedAuthorization {
			return nil, fmt.Errorf("Expected Authorization '%s', got '%s'", expectedAuthorization, authorization)
		}

		b, err := ioutil.ReadAll(request.Body)
		if err != nil {
			return nil, err
		}

		var actual testJson
		if err := json.Unmarshal(b, &actual); err != nil {
			return nil, err
		}

		if actual.Message != expectedMessage {
			return nil, fmt.Errorf("Expected Json Message '%s', got '%s'", expectedMessage, actual.Message)
		}

		return &http.Response{
			StatusCode: http.StatusOK,
			Body:       ioutil.NopCloser(bytes.NewReader([]byte(``))),
		}, nil
	})

	j := testJson{Message: expectedMessage}

	b, err := c.post(expectedURL, nil, j, map[string][]string{"Authorization": []string{expectedAuthorization}})

	assert.Nil(t, err)
	assert.Empty(t, b)
}

func TestHttpError(t *testing.T) {
	expected := errors.FailureError

	c := givenClient(t, "http://127.0.0.1:8080", func(request *http.Request) (response *http.Response, e error) {
		params := bytes.NewBuffer(nil)
		if err := json.NewEncoder(params).Encode(errors.FailureError); err != nil {
			return nil, err
		}

		return &http.Response{
			StatusCode: expected.Status,
			Body:       ioutil.NopCloser(params),
		}, nil
	})

	b, err := c.get("/", nil, nil)

	assert.Equal(t, expected, err)
	assert.Empty(t, b)
}

func givenClient(t *testing.T, host string, client func(request *http.Request) (response *http.Response, e error)) *Client {
	u, err := url.Parse(host)
	assert.Nil(t, err)
	c := newMockClient(client)
	c.host = u.Host
	c.scheme = u.Scheme
	assert.Nil(t, err)
	return c
}
