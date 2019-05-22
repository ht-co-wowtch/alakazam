package client

import "net/http"

type Client struct {
	host string

	client *http.Client
}

func New() {

}
