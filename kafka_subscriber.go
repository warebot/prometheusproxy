package prometheusproxy

import (
	"github.com/Shopify/sarama"
	proto "github.com/golang/protobuf/proto"
	"github.com/prometheus/client_golang/prometheus"
	//dto "github.com/prometheus/client_model/go"
	"strings"
	"time"
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
	producer sarama.AsyncProducer
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

	// start consuming the errors channel in a different go-routine
	go func() {
		for err := range w.producer.Errors() {
			Logger.Errorln(err.Error())
			dropped.WithLabelValues(w.name).Inc()
		}
	}()

	// start consuming the successes channel in a different go-routine
	go func() {
		for _ = range w.producer.Successes() {
			exported.WithLabelValues(w.name).Inc()
		}
	}()

	for m := range ch {
		promMetrics := m.Payload
		message := &MetricsEnvelope{}
		message.MetricFamily = promMetrics
		message.Owner = proto.String(m.Owner)

		message.TimestampMS = proto.Int64(time.Now().Unix() * 1000)
		protoMessage, err := proto.Marshal(message)
		if err != nil {
			dropped.WithLabelValues(w.name).Inc()
			Logger.Errorln(err.Error())
			continue
		}
		w.producer.Input() <- &sarama.ProducerMessage{
			Topic: w.topic,
			Value: sarama.ByteEncoder(protoMessage),
		}
	}

}

func (k *KafkaMetricsSubscriber) connect() sarama.AsyncProducer {
	config := sarama.NewConfig()
	config.Producer.RequiredAcks = sarama.WaitForLocal
	config.Producer.Return.Successes = true
	config.Producer.Retry.Max = 10

	for {
		producer, err := sarama.NewAsyncProducer(k.brokerList, config)
		if err == nil {
			return producer
		}

		Logger.Errorln(err.Error())
		// sleep and try again.
		time.Sleep(5 * time.Second)
	}
}

func (k *KafkaMetricsSubscriber) Start(exported, dropped *prometheus.CounterVec) {

	for i := 0; i < k.concurrencyLevel; i++ {
		producer := k.connect()
		worker := &kafkaWorker{topic: k.topic, producer: producer, name: k.Name()}
		worker.work(k.Chan(), exported, dropped)
	}
}
