package client

import (
	"fmt"
	"github.com/spf13/viper"
	"gitlab.com/jetfueltw/cpw/alakazam/pkg/config"
	"gitlab.com/jetfueltw/cpw/alakazam/pkg/ip"
	"net/url"
	"time"
)

type Conf struct {
	// client host ex 127.0.0.1
	Host string

	// http or https
	Scheme string

	// 最大連接數量
	MaxConns int

	// 最大空閒連接數
	MaxIdleConns int

	// 單一連線最多閒置多久
	IdleConnTimeout time.Duration
}

func Read(path string) (*Conf, error) {
	return ReadTag(path, "")
}

func ReadTag(path string, tag string) (*Conf, error) {
	v, err := config.ReadTag(path, tag)
	if err != nil {
		return nil, err
	}
	return newConfig(v)
}

func newConfig(v *viper.Viper) (*Conf, error) {
	u, err := url.Parse(v.GetString("host"))

	if err != nil {
		return nil, err
	}

	if u.Scheme != "http" && u.Scheme != "https" {
		return nil, fmt.Errorf("scheme error %s", u.Scheme)
	}

	if err := ip.Check(u.Host); err != nil {
		return nil, err
	}

	return &Conf{
		Host:            u.Host,
		Scheme:          u.Scheme,
		MaxConns:        v.GetInt("maxConns"),
		MaxIdleConns:    v.GetInt("maxIdleConns"),
		IdleConnTimeout: time.Second * time.Duration(v.GetInt("idleConnTimeout")),
	}, nil
}
