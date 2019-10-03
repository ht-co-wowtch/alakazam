package message

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/rcrowley/go-metrics"
	"gitlab.com/jetfueltw/cpw/micro/log"
	"strings"
)

// 參考 https://godoc.org/github.com/Shopify/sarama
const (
	// broker
	incomingByteRateName   = "incoming-byte-rate"
	outgoingByteRateName   = "outgoing-byte-rate"
	requestRateName        = "request-rate"
	requestSizeName        = "request-size"
	requestLatencyInMsName = "request-latency-in-ms"
	responseRateName       = "response-rate"
	responseSizeName       = "response-size"

	// producer
	batchSizeName         = "batch-size"
	recordSendRateName    = "record-send-rate"
	recordsPerRequestName = "records-per-request"
	compressionRatioName  = "compression-ratio"
)

func registerProducerMetric(registry metrics.Registry) error {
	metric := newMetricCollector(registry)
	metric.producerDesc = newProducerDesc()
	return prometheus.Register(metric)
}

type brokerDesc struct {
	incomingByteRate   *prometheus.Desc
	outgoingByteRate   *prometheus.Desc
	requestRate        *prometheus.Desc
	requestSize        *prometheus.Desc
	requestLatencyInMs *prometheus.Desc
	responseRate       *prometheus.Desc
	responseSize       *prometheus.Desc
}

func newBrokerDesc() brokerDesc {
	return brokerDesc{
		incomingByteRate: prometheus.NewDesc(
			nameNamespace(incomingByteRateName),
			"Bytes/second read off all brokers",
			nil, nil,
		),
		outgoingByteRate: prometheus.NewDesc(
			nameNamespace(outgoingByteRateName),
			"Bytes/second written off all brokers",
			nil, nil,
		),
		requestRate: prometheus.NewDesc(
			nameNamespace(requestRateName),
			"Requests/second sent to all brokers",
			nil, nil,
		),
		requestSize: prometheus.NewDesc(
			nameNamespace(requestSizeName),
			"Distribution of the request size in bytes for all brokers",
			nil, nil,
		),
		requestLatencyInMs: prometheus.NewDesc(
			nameNamespace(requestLatencyInMsName),
			"Distribution of the request latency in ms for all brokers",
			nil, nil,
		),
		responseRate: prometheus.NewDesc(
			nameNamespace(responseRateName),
			"Responses/second received from all brokers",
			nil, nil,
		),
		responseSize: prometheus.NewDesc(
			nameNamespace(responseSizeName),
			"Distribution of the response size in bytes for all brokers",
			nil, nil,
		),
	}
}

type producerDesc struct {
	batchSize         *prometheus.Desc
	recordSendRate    *prometheus.Desc
	recordsPerRequest *prometheus.Desc
	compressionRatio  *prometheus.Desc
}

func newProducerDesc() producerDesc {
	return producerDesc{
		batchSize: prometheus.NewDesc(
			nameNamespace(batchSizeName),
			"Distribution of the number of bytes sent per partition per request for topics",
			nil, nil,
		),
		recordSendRate: prometheus.NewDesc(
			nameNamespace(recordSendRateName),
			"Records/second sent to topic",
			nil, nil,
		),
		recordsPerRequest: prometheus.NewDesc(
			nameNamespace(recordsPerRequestName),
			"Distribution of the number of records sent per request for topics",
			nil, nil,
		),
		compressionRatio: prometheus.NewDesc(
			nameNamespace(compressionRatioName),
			"Distribution of the number of records sent per request for topics",
			nil, nil,
		),
	}
}

type metricCollector struct {
	brokerDesc       brokerDesc
	producerDesc     producerDesc
	registry         metrics.Registry
	histogramBuckets []float64
}

func newMetricCollector(registry metrics.Registry) *metricCollector {
	return &metricCollector{
		brokerDesc:       newBrokerDesc(),
		registry:         registry,
		histogramBuckets: []float64{0.50, 0.75, 0.9, 0.95, 0.99},
	}
}

func (m *metricCollector) Describe(ch chan<- *prometheus.Desc) {
	ch <- m.brokerDesc.incomingByteRate
	ch <- m.brokerDesc.outgoingByteRate
	ch <- m.brokerDesc.requestRate
	ch <- m.brokerDesc.requestSize
	ch <- m.brokerDesc.requestLatencyInMs
	ch <- m.brokerDesc.responseRate
	ch <- m.brokerDesc.responseSize

	if m.producerDesc.batchSize != nil {
		ch <- m.producerDesc.batchSize
		ch <- m.producerDesc.recordSendRate
		ch <- m.producerDesc.recordsPerRequest
		ch <- m.producerDesc.compressionRatio
	}
}

func (m *metricCollector) Collect(ch chan<- prometheus.Metric) {
	m.registry.Each(func(name string, collector interface{}) {
		switch v := collector.(type) {
		case metrics.Meter:
			if err := m.collectMeter(ch, name, v); err != nil {
				log.Errorf("collect metric [%s] meter error: %v", name, err)
			}
		case metrics.Histogram:
			if err := m.collectHistogram(ch, name, v); err != nil {
				log.Errorf("collect histogram [%s] meter error: %v", name, err)
			}
		}
	})
}

func (m *metricCollector) collectMeter(ch chan<- prometheus.Metric, name string, meter metrics.Meter) error {
	var metric prometheus.Metric
	var desc *prometheus.Desc
	var err error

	switch name {
	case incomingByteRateName:
		desc = m.brokerDesc.incomingByteRate
	case outgoingByteRateName:
		desc = m.brokerDesc.outgoingByteRate
	case requestRateName:
		desc = m.brokerDesc.requestRate
	case responseRateName:
		desc = m.brokerDesc.responseRate
	case recordSendRateName:
		desc = m.producerDesc.recordSendRate
	default:
		return nil
	}

	metric, err = prometheus.NewConstMetric(
		desc,
		prometheus.CounterValue,
		float64(meter.Snapshot().Count()),
	)

	if err == nil && metric != nil {
		ch <- metric
	}
	return err
}

func (m *metricCollector) collectHistogram(ch chan<- prometheus.Metric, name string, histogram metrics.Histogram) error {
	var metric prometheus.Metric
	var desc *prometheus.Desc
	var err error

	switch name {
	case requestSizeName:
		desc = m.brokerDesc.requestSize
	case requestLatencyInMsName:
		desc = m.brokerDesc.requestLatencyInMs
	case responseSizeName:
		desc = m.brokerDesc.responseSize
	case batchSizeName:
		desc = m.producerDesc.batchSize
	case recordsPerRequestName:
		desc = m.producerDesc.recordsPerRequest
	case compressionRatioName:
		desc = m.producerDesc.compressionRatio
	default:
		return nil
	}

	buckets := make(map[float64]uint64)
	snapshot := histogram.Snapshot()

	ps := snapshot.Percentiles(m.histogramBuckets)
	for i, b := range m.histogramBuckets {
		buckets[b] = uint64(ps[i])
	}

	metric, err = prometheus.NewConstHistogram(
		desc,
		uint64(snapshot.Count()),
		float64(snapshot.Sum()),
		buckets,
	)

	if err == nil && metric != nil {
		ch <- metric
	}
	return err
}

func nameNamespace(name string) string {
	return strings.ReplaceAll(name, "-", "_")
}
