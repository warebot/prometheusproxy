package prometheusproxy

import (
	"github.com/Shopify/sarama"
	proto "github.com/golang/protobuf/proto"
	"github.com/prometheus/client_golang/prometheus"
	dto "github.com/prometheus/client_model/go"
	"strings"
)

var brokers string = "localhost:9092"

type KafkaMetricsSubscriber struct {
	dataChan         chan Message
	brokerList       []string
	topic            string
	concurrencyLevel int
}

type kafkaWorker struct {
	topic    string
	producer sarama.SyncProducer
	name     string
}

func NewKafkaMetricsSubscriber(brokers string, topic string, concurrencyLevel int) *KafkaMetricsSubscriber {
	return &KafkaMetricsSubscriber{
		dataChan:         make(chan Message, 500),
		topic:            topic,
		brokerList:       strings.Split(brokers, ","),
		concurrencyLevel: concurrencyLevel,
	}
}

func (k *KafkaMetricsSubscriber) Name() string {
	return "kafka_subscriber"
}

func (k *KafkaMetricsSubscriber) Equals(s Subscriber) bool {
	return true
}

func (k *KafkaMetricsSubscriber) Chan() chan Message {
	return k.dataChan
}

func (w *kafkaWorker) work(ch chan Message, exported, dropped *prometheus.CounterVec) {
	for m := range ch {
		promMetrics := m.Payload
		message := &MetricsEnvelope{}
		message.MetricFamily = promMetrics
		message.Owner = proto.String(m.Owner)

		var filtered *dto.MetricFamily = &dto.MetricFamily{}

		filtered.Help = m.Payload.Help
		filtered.Name = m.Payload.Name
		filtered.Type = m.Payload.Type

		for _, metric := range m.Payload.Metric {
			if !strings.Contains(metric.String(), "value:nan") {
				filtered.Metric = append(filtered.Metric, metric)
				continue
			}
			Logger.Warnln("skipping NaN value", metric.String())
		}
		if len(filtered.Metric) == 0 {
			continue
		}

		protoMessage, err := proto.Marshal(message)
		if err != nil {
			dropped.WithLabelValues(w.name).Inc()
			Logger.Errorln(err.Error())
			continue
		}
		_, _, err = w.producer.SendMessage(&sarama.ProducerMessage{
			Topic: w.topic,
			Value: sarama.ByteEncoder(protoMessage),
		})

		if err != nil {
			Logger.Errorln(err.Error())
			dropped.WithLabelValues(w.name).Inc()
			continue
		}

		exported.WithLabelValues(w.name).Inc()
	}

}
func (k *KafkaMetricsSubscriber) Start(exported, dropped *prometheus.CounterVec) {
	config := sarama.NewConfig()
	config.Producer.RequiredAcks = sarama.WaitForAll
	config.Producer.Retry.Max = 10

	for i := 0; i < k.concurrencyLevel; i++ {
		producer, err := sarama.NewSyncProducer(k.brokerList, config)
		if err != nil {
			panic(err)
		}

		worker := &kafkaWorker{topic: k.topic, producer: producer, name: k.Name()}
		worker.work(k.Chan(), exported, dropped)

	}
}
