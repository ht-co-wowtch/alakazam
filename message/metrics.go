package message

import (
	"fmt"
	kafka "github.com/Shopify/sarama"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/rcrowley/go-metrics"
	"gitlab.com/jetfueltw/cpw/micro/log"
	"strconv"
	"strings"
)

// 參考 https://godoc.org/github.com/Shopify/sarama
// https://docs.confluent.io/current/kafka/monitoring.html#producer-metrics
const (
	forBrokerFormatName = "-for-broker-%d"
	forTopicFormatName  = "-for-topic-%s"

	brokerNameCount   = 7
	producerNameCount = 4
	consumerNameCount = 1

	// broker name
	incomingByteRateName            = "incoming-byte-rate"
	incomingByteRateForBrokerName   = incomingByteRateName + forBrokerFormatName
	outgoingByteRateName            = "outgoing-byte-rate"
	outgoingByteRateForBrokerName   = outgoingByteRateName + forBrokerFormatName
	requestRateName                 = "request-rate"
	requestRateForBrokerName        = requestRateName + forBrokerFormatName
	requestSizeName                 = "request-size"
	requestSizeForBrokerName        = requestSizeName + forBrokerFormatName
	requestLatencyInMsName          = "request-latency-in-ms"
	requestLatencyInMsForBrokerName = requestLatencyInMsName + forBrokerFormatName
	responseRateName                = "response-rate"
	responseRateForBrokerName       = responseRateName + forBrokerFormatName
	responseSizeName                = "response-size"
	responseSizeForBrokerName       = responseSizeName + forBrokerFormatName

	// producer name
	batchSizeName                 = "batch-size"
	batchSizeForTopicName         = batchSizeName + forTopicFormatName
	recordSendRateName            = "record-send-rate"
	recordSendRateForTopicName    = recordSendRateName + forTopicFormatName
	recordsPerRequestName         = "records-per-request"
	recordsPerRequestForTopicName = recordsPerRequestName + forTopicFormatName
	compressionRatioName          = "compression-ratio"
	compressionRatioForTopicName  = compressionRatioName + forTopicFormatName

	// consumer name
	consumerBatchSizeName = "consumer-batch-size"
)

func registerProducerMetric(client kafka.Client, registry metrics.Registry) error {
	metric, err := newMetricCollector(client, registry)
	if err != nil {
		return err
	}

	metric.producerDesc = newProducerDesc(metric.topics)
	metric.isConsumerCollector = false
	return prometheus.Register(metric)
}

func registerConsumerMetric(client kafka.Client, registry metrics.Registry) error {
	metric, err := newMetricCollector(client, registry)
	if err != nil {
		return err
	}

	metric.consumerDesc = newConsumerDesc()
	metric.isProducerCollector = false
	return prometheus.Register(metric)
}

type brokerDesc struct {
	incomingByteRate        *prometheus.Desc
	outgoingByteRate        *prometheus.Desc
	requestRate             *prometheus.Desc
	requestSize             *prometheus.Desc
	requestLatencyInMs      *prometheus.Desc
	responseRate            *prometheus.Desc
	responseSize            *prometheus.Desc
	incomingByteRateNames   []string
	outgoingByteRateNames   []string
	requestRateNames        []string
	requestSizeNames        []string
	requestLatencyInMsNames []string
	responseRateNames       []string
	responseSizeNames       []string
}

func newBrokerDesc(brokerIds []int32) brokerDesc {
	incomingByteRateNames := []string{incomingByteRateName}
	outgoingByteRateNames := []string{outgoingByteRateName}
	requestRateNames := []string{requestRateName}
	requestSizeNames := []string{requestSizeName}
	requestLatencyInMsNames := []string{requestLatencyInMsName}
	responseRateNames := []string{responseRateName}
	responseSizeNames := []string{responseSizeName}

	for _, id := range brokerIds {
		incomingByteRateNames = append(incomingByteRateNames, fmt.Sprintf(incomingByteRateForBrokerName, id))
		outgoingByteRateNames = append(outgoingByteRateNames, fmt.Sprintf(outgoingByteRateForBrokerName, id))
		requestRateNames = append(requestRateNames, fmt.Sprintf(requestRateForBrokerName, id))
		requestSizeNames = append(requestSizeNames, fmt.Sprintf(requestSizeForBrokerName, id))
		requestLatencyInMsNames = append(requestLatencyInMsNames, fmt.Sprintf(requestLatencyInMsForBrokerName, id))
		responseRateNames = append(responseRateNames, fmt.Sprintf(responseRateForBrokerName, id))
		responseSizeNames = append(responseSizeNames, fmt.Sprintf(responseSizeForBrokerName, id))
	}

	return brokerDesc{
		incomingByteRateNames: incomingByteRateNames,
		incomingByteRate: prometheus.NewDesc(
			nameNamespace(incomingByteRateName),
			"Bytes/second read off brokers",
			[]string{"broker", "key"}, nil,
		),
		outgoingByteRateNames: outgoingByteRateNames,
		outgoingByteRate: prometheus.NewDesc(
			nameNamespace(outgoingByteRateName),
			"Bytes/second written off brokers",
			[]string{"broker", "key"}, nil,
		),
		requestRateNames: requestRateNames,
		requestRate: prometheus.NewDesc(
			nameNamespace(requestRateName),
			"Requests/second sent to brokers",
			[]string{"broker", "key"}, nil,
		),
		requestSizeNames: requestSizeNames,
		requestSize: prometheus.NewDesc(
			nameNamespace(requestSizeName),
			"Distribution of the request size in bytes for brokers",
			[]string{"broker"}, nil,
		),
		requestLatencyInMsNames: requestLatencyInMsNames,
		requestLatencyInMs: prometheus.NewDesc(
			nameNamespace(requestLatencyInMsName),
			"Distribution of the request latency in ms for brokers",
			[]string{"broker"}, nil,
		),
		responseRateNames: responseRateNames,
		responseRate: prometheus.NewDesc(
			nameNamespace(responseRateName),
			"Responses/second received from brokers",
			[]string{"broker", "key"}, nil,
		),
		responseSizeNames: responseSizeNames,
		responseSize: prometheus.NewDesc(
			nameNamespace(responseSizeName),
			"Distribution of the response size in bytes for brokers",
			[]string{"broker"}, nil,
		),
	}
}

type producerDesc struct {
	batchSize              *prometheus.Desc
	recordSendRate         *prometheus.Desc
	recordsPerRequest      *prometheus.Desc
	compressionRatio       *prometheus.Desc
	batchSizeNames         []string
	recordSendRateNames    []string
	recordsPerRequestNames []string
	compressionRatioNames  []string
}

func newProducerDesc(topics []string) producerDesc {
	batchSizeNames := []string{batchSizeName}
	recordSendRateNames := []string{recordSendRateName}
	recordsPerRequestNames := []string{recordsPerRequestName}
	compressionRatioNames := []string{compressionRatioName}

	for _, name := range topics {
		batchSizeNames = append(batchSizeNames, fmt.Sprintf(batchSizeForTopicName, name))
		recordSendRateNames = append(recordSendRateNames, fmt.Sprintf(recordSendRateForTopicName, name))
		recordsPerRequestNames = append(recordsPerRequestNames, fmt.Sprintf(recordsPerRequestForTopicName, name))
		compressionRatioNames = append(compressionRatioNames, fmt.Sprintf(compressionRatioForTopicName, name))
	}

	return producerDesc{
		batchSizeNames: batchSizeNames,
		batchSize: prometheus.NewDesc(
			nameNamespace(batchSizeName),
			"Distribution of the number of bytes sent per partition per request for topics",
			[]string{"topic"}, nil,
		),
		recordSendRateNames: recordSendRateNames,
		recordSendRate: prometheus.NewDesc(
			nameNamespace(recordSendRateName),
			"Records/second sent to topic",
			[]string{"topic", "key"}, nil,
		),
		recordsPerRequestNames: recordsPerRequestNames,
		recordsPerRequest: prometheus.NewDesc(
			nameNamespace(recordsPerRequestName),
			"Distribution of the number of records sent per request for topics",
			[]string{"topic"}, nil,
		),
		compressionRatioNames: compressionRatioNames,
		compressionRatio: prometheus.NewDesc(
			nameNamespace(compressionRatioName),
			"Distribution of the number of records sent per request for topics",
			[]string{"topic"}, nil,
		),
	}
}

type consumerDesc struct {
	consumerBatchSize *prometheus.Desc
}

func newConsumerDesc() consumerDesc {
	return consumerDesc{
		consumerBatchSize: prometheus.NewDesc(
			nameNamespace(consumerBatchSizeName),
			"Distribution of the number of messages in a batch",
			nil, nil,
		),
	}
}

type metricCollector struct {
	brokerIds           []int32
	topics              []string
	brokerDesc          brokerDesc
	producerDesc        producerDesc
	consumerDesc        consumerDesc
	registry            metrics.Registry
	histogramBuckets    []float64
	isProducerCollector bool
	isConsumerCollector bool
}

func newMetricCollector(client kafka.Client, registry metrics.Registry) (*metricCollector, error) {
	var bid []int32

	for _, b := range client.Brokers() {
		bid = append(bid, b.ID())
	}

	topics, err := client.Topics()
	if err != nil {
		return nil, err
	}

	return &metricCollector{
		brokerIds:           bid,
		topics:              topics,
		brokerDesc:          newBrokerDesc(bid),
		registry:            registry,
		histogramBuckets:    []float64{0.50, 0.75, 0.9, 0.95, 0.99},
		isProducerCollector: true,
		isConsumerCollector: true,
	}, nil
}

func (m *metricCollector) Describe(ch chan<- *prometheus.Desc) {
	ch <- m.brokerDesc.incomingByteRate
	ch <- m.brokerDesc.outgoingByteRate
	ch <- m.brokerDesc.requestRate
	ch <- m.brokerDesc.requestSize
	ch <- m.brokerDesc.requestLatencyInMs
	ch <- m.brokerDesc.responseRate
	ch <- m.brokerDesc.responseSize

	if m.isProducerCollector {
		ch <- m.producerDesc.batchSize
		ch <- m.producerDesc.recordSendRate
		ch <- m.producerDesc.recordsPerRequest
		// 沒有開訊息壓縮先註解
		//ch <- m.producerDesc.compressionRatio
	}

	if m.isConsumerCollector {
		ch <- m.consumerDesc.consumerBatchSize
	}
}

func (m *metricCollector) Collect(ch chan<- prometheus.Metric) {
	brokerLabel := []string{"all"}
	for _, id := range m.brokerIds {
		brokerLabel = append(brokerLabel, strconv.Itoa(int(id)))
	}

	topicLabel := []string{"all"}
	for _, name := range m.topics {
		if name == "__consumer_offsets" {
			continue
		}
		topicLabel = append(topicLabel, name)
	}

	count := brokerNameCount * len(brokerLabel)

	if m.isProducerCollector {
		count += producerNameCount * len(topicLabel)
	}
	if m.isConsumerCollector {
		count += consumerNameCount
	}

	collect := &metricConst{
		metrics:          make([]prometheus.Metric, 0, count),
		meter:            make(map[string]metrics.Meter),
		histogram:        make(map[string]metrics.Histogram),
		histogramBuckets: m.histogramBuckets,
	}

	m.registry.Each(func(name string, collector interface{}) {
		switch v := collector.(type) {
		case metrics.Meter:
			collect.meter[name] = v
		case metrics.Histogram:
			collect.histogram[name] = v
		}
	})

	collect.putMeter(m.brokerDesc.incomingByteRate, m.brokerDesc.incomingByteRateNames, brokerLabel)
	collect.putMeter(m.brokerDesc.outgoingByteRate, m.brokerDesc.outgoingByteRateNames, brokerLabel)
	collect.putHistogram(m.brokerDesc.requestSize, m.brokerDesc.requestSizeNames, brokerLabel)
	collect.putMeter(m.brokerDesc.requestRate, m.brokerDesc.requestRateNames, brokerLabel)
	collect.putHistogram(m.brokerDesc.requestLatencyInMs, m.brokerDesc.requestLatencyInMsNames, brokerLabel)
	collect.putHistogram(m.brokerDesc.responseSize, m.brokerDesc.responseSizeNames, brokerLabel)
	collect.putMeter(m.brokerDesc.responseRate, m.brokerDesc.responseRateNames, brokerLabel)

	if m.isProducerCollector {
		collect.putHistogram(m.producerDesc.batchSize, m.producerDesc.batchSizeNames, topicLabel)
		collect.putMeter(m.producerDesc.recordSendRate, m.producerDesc.recordSendRateNames, topicLabel)
		collect.putHistogram(m.producerDesc.recordsPerRequest, m.producerDesc.recordsPerRequestNames, topicLabel)
		// 沒有開訊息壓縮先註解
		//collect.putHistogram(m.producerDesc.compressionRatio, m.producerDesc.compressionRatioNames, topicLabel)
	}

	if m.isConsumerCollector {
		if consumer, ok := collect.histogram[consumerBatchSizeName]; ok {
			snapshot := consumer.Snapshot()
			buckets := make(map[float64]uint64)

			ps := snapshot.Percentiles(m.histogramBuckets)
			for i, b := range m.histogramBuckets {
				buckets[b] = uint64(ps[i])
			}

			metric, err := prometheus.NewConstHistogram(
				m.consumerDesc.consumerBatchSize,
				uint64(snapshot.Count()),
				float64(snapshot.Sum()),
				buckets,
			)

			if err == nil {
				ch <- metric
			} else {
				log.Errorf("collect metric [%s] histogram error: %v", consumerBatchSizeName, err)
			}
		}
	}

	for _, v := range collect.metrics {
		ch <- v
	}
}

type metricConst struct {
	metrics          []prometheus.Metric
	meter            map[string]metrics.Meter
	histogram        map[string]metrics.Histogram
	histogramBuckets []float64
}

func (m *metricConst) putMeter(desc *prometheus.Desc, name []string, label []string) {
	for i, n := range name {
		if data, ok := m.meter[n]; ok {
			snapshot := data.Snapshot()

			for _, key := range []string{"1m", "5m", "15m", "mean"} {
				var value float64
				switch key {
				case "1m":
					value = snapshot.Rate1()
				case "5m":
					value = snapshot.Rate5()
				case "15m":
					value = snapshot.Rate15()
				case "mean":
					value = snapshot.RateMean()
				}
				metric, err := prometheus.NewConstMetric(
					desc,
					prometheus.CounterValue,
					value,
					label[i], key,
				)
				if err == nil {
					m.metrics = append(m.metrics, metric)
				} else {
					log.Errorf("collect metric [%s] meter for key: %s error: %v", name, key, err)
				}
			}
		}
	}
}

func (m *metricConst) putHistogram(desc *prometheus.Desc, name []string, label []string) {
	for i, n := range name {
		if data, ok := m.histogram[n]; ok {
			buckets := make(map[float64]uint64)
			snapshot := data.Snapshot()

			ps := snapshot.Percentiles(m.histogramBuckets)
			for of, b := range m.histogramBuckets {
				buckets[b] = uint64(ps[of])
			}

			metric, err := prometheus.NewConstHistogram(
				desc,
				uint64(snapshot.Count()),
				float64(snapshot.Sum()),
				buckets,
				label[i],
			)

			if err == nil {
				m.metrics = append(m.metrics, metric)
			} else {
				log.Errorf("collect metric [%s] histogram error: %v", name, err)
			}
		}
	}
}

func nameNamespace(name string) string {
	return strings.ReplaceAll(name, "-", "_")
}
