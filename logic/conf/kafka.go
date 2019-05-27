package conf

import "github.com/spf13/viper"

// Kafka
type Kafka struct {
	// Kafka 推送與接收Topic
	Topic string

	// 節點ip
	Brokers []string
}

func newKafka() *Kafka {
	return &Kafka{
		Topic:   viper.GetString("kafka.topic"),
		Brokers: viper.GetStringSlice("kafka.brokers"),
	}
}
