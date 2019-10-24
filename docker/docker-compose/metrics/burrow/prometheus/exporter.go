package main

import (
	"context"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"gitlab.com/jetfueltw/cpw/micro/log"
	"go.uber.org/zap"
	"net/http"
	"strconv"
	"time"
)

// https://github.com/linkedin/Burrow/wiki/http-request-consumer-group-status#response
var Status = map[string]int{
	"NOTFOUND": 1,
	"OK":       2,
	"WARN":     3,
	"ERR":      4,
	"STOP":     5,
	"STALL":    6,
	"REWIND":   7,
}

type exporter struct {
	client            *Client
	context           context.Context
	metricsListenAddr string
	interval          time.Duration
	metrics           metrics
}

type metrics struct {
	TopicPartitionOffset           *prometheus.GaugeVec
	ConsumerPartitionLag           *prometheus.GaugeVec
	ConsumerTotalLag               *prometheus.GaugeVec
	ConsumerPartitionCurrentOffset *prometheus.GaugeVec
	ConsumerPartitionStatus        *prometheus.GaugeVec
	ConsumerStatus                 *prometheus.GaugeVec
}

func newExporter(context context.Context, metricsListenAddr, burrowAddr string, interval time.Duration) (*exporter, error) {
	c, err := newClient(burrowAddr)
	if err != nil {
		return nil, err
	}

	m := metrics{
		TopicPartitionOffset: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "kafka_burrow_topic_partition_offset",
				Help: "The latest offset on a topic's partition as reported by burrow.",
			},
			[]string{"cluster", "topic", "partition"},
		),
		ConsumerPartitionLag: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "kafka_burrow_consumer_partition_lag",
				Help: "The lag of the latest offset commit on a partition as reported by burrow.",
			},
			[]string{"cluster", "group", "topic", "partition"},
		),
		ConsumerTotalLag: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "kafka_burrow_consumer_total_lag",
				Help: "The total amount of lag for the consumer group as reported by burrow",
			},
			[]string{"cluster", "group"},
		),
		ConsumerPartitionCurrentOffset: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "kafka_burrow_consumer_partition_current_offset",
				Help: "The latest offset commit on a partition as reported by burrow.",
			},
			[]string{"cluster", "group", "topic", "partition"},
		),
		ConsumerPartitionStatus: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "kafka_burrow_consumer_partition_status",
				Help: "The status of a partition as reported by burrow.",
			},
			[]string{"cluster", "group", "topic", "partition"},
		),
		ConsumerStatus: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "kafka_burrow_consumer_status",
				Help: "The status of a partition as reported by burrow.",
			},
			[]string{"cluster", "group"},
		),
	}

	prometheus.MustRegister(m.ConsumerPartitionLag)
	prometheus.MustRegister(m.ConsumerTotalLag)
	prometheus.MustRegister(m.TopicPartitionOffset)
	prometheus.MustRegister(m.ConsumerPartitionCurrentOffset)
	prometheus.MustRegister(m.ConsumerStatus)
	prometheus.MustRegister(m.ConsumerPartitionStatus)

	return &exporter{
		context:           context,
		metricsListenAddr: metricsListenAddr,
		interval:          interval,
		client:            c,
		metrics:           m,
	}, nil
}

func (e *exporter) Start() {
	http.Handle("/metrics", promhttp.Handler())
	go func() {
		if err := http.ListenAndServe(e.metricsListenAddr, nil); err != nil {
			log.Error("start metrics http server", zap.Error(err))
		}
	}()
	go e.mainLoop()
}

func (e *exporter) mainLoop() {
	interval := time.NewTicker(e.interval)

	e.run()

	for {
		select {
		case <-interval.C:
			log.Info("start collector")
			e.run()
		case <-e.context.Done():
			log.Info("stop for mainLoop")
			interval.Stop()
			return
		}
	}
}

func (e *exporter) run() {
	clustersResp, err := e.client.clusterList()
	if err != nil {
		log.Error("cluster list", zap.Error(err))
		return
	}

	for _, cluster := range clustersResp.Clusters {
		kafkaCluster, err := e.client.topicList(cluster)
		if err != nil {
			log.Error("topic list", zap.Error(err), zap.String("cluster", cluster))
			return
		}

		for _, topic := range kafkaCluster.Topics {
			if topic == "__consumer_offsets" {
				continue
			}

			consumersListResp, err := e.client.consumersList(cluster, topic)
			if err != nil {
				log.Error("consumers list", zap.Error(err), zap.String("cluster", cluster), zap.String("topic", topic))
				return
			}

			for _, consumer := range consumersListResp.Consumers {
				e.consumersLag(cluster, consumer)
			}

			e.topicOffsets(cluster, topic)
		}
	}
}

func (e *exporter) consumersLag(cluster, consumerGroup string) {
	consumerLagResp, err := e.client.consumersLag(cluster, consumerGroup)
	if err != nil {
		log.Error("consumers lag", zap.Error(err), zap.String("cluster", cluster), zap.String("consumer group", consumerGroup))
		return
	}

	for _, lag := range consumerLagResp.Status.Partitions {
		labels := prometheus.Labels{
			"cluster":   cluster,
			"group":     consumerGroup,
			"topic":     lag.Topic,
			"partition": strconv.Itoa(lag.Partition),
		}

		e.metrics.ConsumerPartitionLag.With(labels).Set(float64(lag.CurrentLag))
		e.metrics.ConsumerPartitionCurrentOffset.With(labels).Set(float64(lag.End.Offset))
		e.metrics.ConsumerPartitionStatus.With(labels).Set(float64(Status[lag.Status]))
	}

	labels := prometheus.Labels{
		"cluster": cluster,
		"group":   consumerGroup,
	}

	e.metrics.ConsumerTotalLag.With(labels).Set(float64(consumerLagResp.Status.TotalLag))
	e.metrics.ConsumerStatus.With(labels).Set(float64(Status[consumerLagResp.Status.Status]))
}

func (e *exporter) topicOffsets(cluster, topic string) {
	offsetsResp, err := e.client.offsets(cluster, topic)
	if err != nil {
		log.Error("new offsets", zap.Error(err), zap.String("cluster", cluster), zap.String("topic", topic))
		return
	}

	for partition, offset := range offsetsResp.Offsets {
		e.metrics.TopicPartitionOffset.With(prometheus.Labels{
			"cluster":   cluster,
			"topic":     topic,
			"partition": strconv.Itoa(partition),
		}).Set(float64(offset))
	}
}
