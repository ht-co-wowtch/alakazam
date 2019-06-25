package client

import (
	"encoding/json"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
	"io/ioutil"
	"net/http"
	"net/url"
	"testing"
	"time"
)

type clientTestSuite struct {
	suite.Suite
	client *Client
}

func TestClientSuite(t *testing.T) {
	mux := http.NewServeMux()

	mux.Handle("/get", http.HandlerFunc(get))
	mux.Handle("/postJson", http.HandlerFunc(postJson))
	mux.Handle("/putJson", http.HandlerFunc(putJson))
	mux.Handle("/delete", http.HandlerFunc(deletes))

	server := &http.Server{Addr: ":1111", Handler: mux}

	go func() {
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			t.Fatal(err)
		}
	}()

	suite.Run(t, new(clientTestSuite))

	server.Close()
}

func (suite *clientTestSuite) SetupTest() {
	suite.client = New(&Conf{
		Host:            "127.0.0.1:1111",
		Scheme:          "http",
		MaxConns:        10,
		MaxIdleConns:    1,
		IdleConnTimeout: time.Second,
	})

}

func (suite *clientTestSuite) TestGet() {
	t := suite.T()
	resp, err := suite.client.Get("/get", url.Values{"key": {"value"}}, nil)

	if err != nil {
		t.Fatalf("http get error(%v)", err)
	}

	b, err := ioutil.ReadAll(resp.Body)

	if err != nil {
		t.Fatalf("ioutil.ReadAll error(%v)", err)
	}

	assert.Equal(t, string(b), "ok")
	assert.Equal(t, resp.StatusCode, http.StatusOK)
}

func (suite *clientTestSuite) TestPostJson() {
	t := suite.T()

	p := testJson{Status: 1}

	resp, err := suite.client.PostJson("/postJson", nil, p, nil)

	if err != nil {
		t.Fatalf("http postJson error(%v)", err)
	}

	assertRequest(t, resp, p)
}

func (suite *clientTestSuite) TestPutJson() {
	t := suite.T()

	p := testJson{Status: 1}

	resp, err := suite.client.PutJson("/putJson", nil, p, nil)

	if err != nil {
		t.Fatalf("http putJson error(%v)", err)
	}

	assertRequest(t, resp, p)
}

func (suite *clientTestSuite) TestDelete() {
	t := suite.T()
	resp, err := suite.client.Delete("/delete", url.Values{"key": {"value"}}, nil)

	if err != nil {
		t.Fatalf("http get error(%v)", err)
	}

	b, err := ioutil.ReadAll(resp.Body)

	if err != nil {
		t.Fatalf("ioutil.ReadAll error(%v)", err)
	}

	assert.Equal(t, string(b), "ok")
	assert.Equal(t, resp.StatusCode, http.StatusOK)
}

func assertRequest(t *testing.T, resp *http.Response, p testJson) {
	b, err := ioutil.ReadAll(resp.Body)

	if err != nil {
		t.Fatalf("ioutil.ReadAll error(%v)", err)
	}

	actual := new(testJson)
	err = json.Unmarshal(b, &actual)

	if err != nil {
		t.Fatalf("json.Unmarshal error(%v)", err)
	}

	assert.Equal(t, actual, &p)
	assert.Equal(t, resp.StatusCode, http.StatusOK)
}

func get(w http.ResponseWriter, req *http.Request) {
	if req.Method != "GET" {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Write([]byte(`ok`))
}

type testJson struct {
	Status int `json:"status"`
}

func postJson(w http.ResponseWriter, req *http.Request) {
	toResponse("POST", w, req)
}

func putJson(w http.ResponseWriter, req *http.Request) {
	toResponse("PUT", w, req)
}

func deletes(w http.ResponseWriter, req *http.Request) {
	if req.Method != "DELETE" {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Write([]byte(`ok`))
}

func toResponse(method string, w http.ResponseWriter, req *http.Request) {
	if req.Method != method {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	if req.Header.Get(contentType) != jsonHeaderType {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	b, err := ioutil.ReadAll(req.Body)

	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Write(b)
}
