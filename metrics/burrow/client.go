package main

import (
	"encoding/json"
	"fmt"
	"gitlab.com/jetfueltw/cpw/alakazam/errors"
	"gitlab.com/jetfueltw/cpw/micro/client"
	"net/http"
	"net/url"
	"time"
)

const (
	ClusterList        = "/v3/kafka"
	TopicList          = "/v3/kafka/%s/topic"
	Offsets            = "/v3/kafka/%s/topic/%s"
	TopicConsumersList = "/v3/kafka/%s/topic/%s/consumers"
	ConsumersLag       = "/v3/kafka/%s/consumer/%s/lag"
)

type causer struct {
	error `json:"-"`

	Message string `json:"message"`
	status  int    `json:"-"`
}

type Client struct {
	client *client.Client
}

func newClient(rawurl string) (*Client, error) {
	u, err := url.Parse(rawurl)
	if err != nil {
		return nil, err
	}

	conf := &client.Conf{
		Host:            u.Host,
		Scheme:          u.Scheme,
		MaxConns:        2,
		MaxIdleConns:    1,
		IdleConnTimeout: time.Minute,
	}
	return &Client{
		client: client.New(conf),
	}, nil
}

type clustersResp struct {
	Error    bool     `json:"error"`
	Message  string   `json:"message"`
	Clusters []string `json:"clusters"`
}

// https://github.com/linkedin/Burrow/wiki/http-request-list-clusters
func (c *Client) clusterList() (clustersResp, error) {
	resp, err := c.client.Get(ClusterList, nil, nil)
	if err != nil {
		return clustersResp{}, err
	}

	defer resp.Body.Close()

	if err := checkResponse(resp); err != nil {
		return clustersResp{}, err
	}

	var data clustersResp
	err = json.NewDecoder(resp.Body).Decode(&data)
	if data.Error {
		return clustersResp{}, errors.New(data.Message)
	}
	return data, err
}

type topicsResp struct {
	Error   bool     `json:"error"`
	Message string   `json:"message"`
	Topics  []string `json:"topics"`
}

// https://github.com/linkedin/Burrow/wiki/http-request-list-cluster-topics
func (c *Client) topicList(cluster string) (topicsResp, error) {
	resp, err := c.client.Get(fmt.Sprintf(TopicList, cluster), nil, nil)
	if err != nil {
		return topicsResp{}, err
	}

	defer resp.Body.Close()

	if err := checkResponse(resp); err != nil {
		return topicsResp{}, err
	}

	var data topicsResp
	err = json.NewDecoder(resp.Body).Decode(&data)
	if data.Error {
		return topicsResp{}, errors.New(data.Message)
	}
	return data, err
}

type offsetsResp struct {
	Error   bool   `json:"error"`
	Message string `json:"message"`
	Offsets []int  `json:"offsets"`
}

// https://github.com/linkedin/Burrow/wiki/http-request-get-topic-detail
func (c *Client) offsets(cluster, topic string) (offsetsResp, error) {
	resp, err := c.client.Get(fmt.Sprintf(Offsets, cluster, topic), nil, nil)
	if err != nil {
		return offsetsResp{}, err
	}

	defer resp.Body.Close()

	if err := checkResponse(resp); err != nil {
		return offsetsResp{}, err
	}

	var data offsetsResp
	err = json.NewDecoder(resp.Body).Decode(&data)
	if data.Error {
		return offsetsResp{}, errors.New(data.Message)
	}
	return data, err
}

type consumersResp struct {
	Error     bool     `json:"error"`
	Message   string   `json:"message"`
	Consumers []string `json:"consumers"`
}

// https://github.com/linkedin/Burrow/wiki/http-request-list-consumers
func (c *Client) consumersList(cluster, consumer string) (consumersResp, error) {
	resp, err := c.client.Get(fmt.Sprintf(TopicConsumersList, cluster, consumer), nil, nil)
	if err != nil {
		return consumersResp{}, err
	}

	defer resp.Body.Close()

	if err := checkResponse(resp); err != nil {
		return consumersResp{}, err
	}

	var data consumersResp
	err = json.NewDecoder(resp.Body).Decode(&data)
	if data.Error {
		return consumersResp{}, errors.New(data.Message)
	}
	return data, err
}

type consumerLagResp struct {
	Error   bool      `json:"error"`
	Message string    `json:"message"`
	Status  logStatus `json:"status"`
}

type logStatus struct {
	Cluster        string             `json:"cluster"`
	Group          string             `json:"group"`
	Status         string             `json:"status"`
	Complete       float64            `json:"complete"`
	PartitionCount int                `json:"partition_count"`
	TotalLag       int                `json:"totallag"`
	Partitions     []partitionsStatus `json:"partitions"`
}

type partitionsStatus struct {
	Topic      string              `json:"topic"`
	Partition  int                 `json:"partition"`
	Status     string              `json:"status"`
	Start      partitionsOffsetLag `json:"start"`
	End        partitionsOffsetLag `json:"end"`
	CurrentLag int                 `json:"current_lag"`
	Complete   float64             `json:"complete"`
}

type partitionsOffsetLag struct {
	Offset    int `json:"offset"`
	Timestamp int `json:"timestamp"`
	Lag       int `json:"lag"`
}

// https://github.com/linkedin/Burrow/wiki/http-request-consumer-group-status
func (c *Client) consumersLag(cluster, consumerGroup string) (consumerLagResp, error) {
	resp, err := c.client.Get(fmt.Sprintf(ConsumersLag, cluster, consumerGroup), nil, nil)
	if err != nil {
		return consumerLagResp{}, err
	}

	defer resp.Body.Close()

	if err := checkResponse(resp); err != nil {
		return consumerLagResp{}, err
	}

	var data consumerLagResp
	err = json.NewDecoder(resp.Body).Decode(&data)
	if data.Error {
		return consumerLagResp{}, errors.New(data.Message)
	}
	return data, err
}
func checkResponse(resp *http.Response) error {
	if resp.StatusCode-http.StatusOK > 100 {
		e := new(causer)
		e.status = resp.StatusCode
		if err := json.NewDecoder(resp.Body).Decode(e); err != nil {
			return err
		}
		return e
	}
	return nil
}
