package prometheusproxy

import (
	"encoding/json"
	"github.com/prometheus/client_golang/prometheus"
	dto "github.com/prometheus/client_model/go"
	"net"
	"time"
)

// Message is the message definition of a metric message on the wire.
type Message struct {
	Owner   string            `json:"owner,omitempty"`
	Payload *dto.MetricFamily `json:"payload"`
}

// TCPMetricsExporter implements Subscriber
type TCPMetricsExporter struct {
	destAddr          string
	dataChan          chan Message
	exported, dropped prometheus.Counter
	conn              net.Conn
}

func NewTCPMetricsExporter(destAddr string,
	exported, dropped prometheus.Counter) *TCPMetricsExporter {
	return &TCPMetricsExporter{
		destAddr: destAddr,
		dataChan: make(chan Message, 1000),
		exported: exported,
		dropped:  dropped,
	}

}

func (t *TCPMetricsExporter) Name() string {
	return "tcp-metrics-exporter"
}

// Chan() implements Subscriber.
// Returns the subscriber's downstream channel
func (t *TCPMetricsExporter) Chan() chan Message {
	return t.dataChan
}

func (t *TCPMetricsExporter) connect(connected chan struct{}) {
	var err error
	for {
		t.conn, err = net.Dial("tcp", t.destAddr)
		Info.Println("re-establishing connecion")
		if err == nil {
			Info.Println("connected")
			connected <- struct{}{}
			return
		}
		time.Sleep(time.Second * time.Duration(10))
	}
}

func (t *TCPMetricsExporter) Start() {
	connect := make(chan struct{}, 1)
	reconnect := make(chan struct{}, 1)

	t.connect(connect)
	<-connect

	connected := true
	for m := range t.dataChan {
		// A message was recieved, and our TCP connection is "ok". Proceed.
		if connected {
			data, err := json.Marshal(m)
			if err != nil {
				t.dropped.Inc()
				Error.Println(err.Error())
				continue
			}

			data = append(data, []byte("\n")...)
			if _, err = t.conn.Write(data); err != nil {
				t.dropped.Inc()
				connected = false
				continue
			}
			t.exported.Inc()
			continue
		}

		// The exporter is in an errorneous state and needs to re-establish
		// the connection.

		Info.Println(len(connect), cap(connect))
		select {
		case reconnect <- struct{}{}:
			Info.Println("re-connecting")
			go t.connect(connect)
		case <-connect:
			Info.Println("connection ok")
			<-reconnect
			connected = true
		default:
			t.dropped.Inc()
			continue
		}
	}
}
