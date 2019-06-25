package client

import (
	"github.com/stretchr/testify/assert"
	"testing"
	"time"
)

func TestConfigYml(t *testing.T) {
	c, err := Read("./config.example.yml")

	if err != nil {
		t.Fatalf("read client yml error(%v)", err)
	}

	assert.Equal(t, c, &Conf{
		Host:            "127.0.0.1:8080",
		Scheme:          "http",
		MaxConns:        30,
		MaxIdleConns:    10,
		IdleConnTimeout: time.Minute * 2,
	})
}
