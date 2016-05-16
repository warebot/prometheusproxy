package prometheusproxy

import (
	"encoding/json"
	"github.com/prometheus/client_golang/prometheus"
	dto "github.com/prometheus/client_model/go"
	"net"
	"reflect"
	"strings"
	"time"
)

// Message is the message definition of a metric message on the wire.
type Message struct {
	Owner   string            `json:"owner,omitempty"`
	Payload *dto.MetricFamily `json:"payload"`
}

// TCPMetricsExporter implements Subscriber
type TCPMetricsSubscriber struct {
	destAddr         string
	dataChan         chan Message
	concurrencyLevel int
}

func NewTCPMetricsSubscriber(destAddr string, concurrencyLevel int) *TCPMetricsSubscriber {
	return &TCPMetricsSubscriber{
		concurrencyLevel: concurrencyLevel,
		destAddr:         destAddr,
		dataChan:         make(chan Message, 500),
	}

}

func (t *TCPMetricsSubscriber) Name() string {
	return "tcp_subscriber"
}

func (t *TCPMetricsSubscriber) Equals(s Subscriber) bool {
	if reflect.TypeOf(s) != reflect.TypeOf(&TCPMetricsSubscriber{}) {
		return false
	}

	o := s.(*TCPMetricsSubscriber)
	return t.destAddr == o.destAddr && t.concurrencyLevel == o.concurrencyLevel
}

// Chan() implements Subscriber.
// Returns the subscriber's downstream channel
func (t *TCPMetricsSubscriber) Chan() chan Message {
	return t.dataChan
}

func (t *TCPMetricsSubscriber) Start(exported, dropped *prometheus.CounterVec) {
	for i := 0; i < t.concurrencyLevel; i++ {
		worker := worker{destAddr: t.destAddr, name: t.Name()}
		go worker.work(t.dataChan, exported, dropped)
	}

}

type worker struct {
	destAddr string
	name     string
	conn     net.Conn
}

func (w *worker) work(ch chan Message, exported, dropped *prometheus.CounterVec) {
	connect := make(chan struct{}, 1)
	reconnect := make(chan struct{}, 1)

	w.connect(connect)
	<-connect

	connected := true
	for m := range ch {
		// A message was recieved, and our TCP connection is "ok". Proceed.
		if connected {
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

			data, err := json.Marshal(filtered)
			if err != nil {
				dropped.WithLabelValues(w.name).Inc()
				Logger.Errorln(err.Error())
				continue
			}

			data = append(data, []byte("\n")...)
			if _, err = w.conn.Write(data); err != nil {
				dropped.WithLabelValues(w.name).Inc()
				connected = false
				continue
			}
			exported.WithLabelValues(w.name).Inc()
			continue
		}

		// The exporter is in an errorneous state and needs to re-establish
		// the connection.
		select {
		case reconnect <- struct{}{}:
			Logger.Infoln("re-connecting")
			go w.connect(connect)
		case <-connect:
			Logger.Infoln("connection ok")
			<-reconnect
			connected = true
		default:
			dropped.WithLabelValues(w.name).Inc()
			continue
		}
	}

}

func (w *worker) connect(connected chan struct{}) {
	var err error
	for {
		w.conn, err = net.Dial("tcp", w.destAddr)
		Logger.Infoln("establishing connection")
		if err == nil {
			Logger.Infoln("connected")
			connected <- struct{}{}
			return
		}
		time.Sleep(time.Second * time.Duration(10))
	}
}
