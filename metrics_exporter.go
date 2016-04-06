package main

import (
	"encoding/json"
	dto "github.com/prometheus/client_model/go"
	"net"
	"time"
)

// Message is the message definition of a metric message on the wire.
type Message struct {
	Owner   string            `json:"owner,omitempty"`
	Payload *dto.MetricFamily `json:"payload"`
}

// MetricsExporter is an interface to allow various concrete implementations of our
// downstream exporting functionality.
type MetricsExporter interface {
	start()
}

type TCPMetricsExporter struct {
	destAddr string
	dataChan chan Message
	conn     net.Conn
}

func (t *TCPMetricsExporter) reconnect(reconnect chan struct{}, established chan struct{}, quit chan struct{}) {
	// Some type of backoff algorithm
	seed := 5
	for i := 1; i < 20000; i++ {
		conn, err := net.Dial("tcp", t.destAddr)
		if err == nil {
			t.conn = conn
			Info.Println("connected")
			established <- struct{}{}
			return
		}
		time.Sleep(time.Second * time.Duration(i*seed))
	}
	quit <- struct{}{}
}

func (t *TCPMetricsExporter) connect() net.Conn {
	conn, err := net.Dial("tcp", t.destAddr)
	if err != nil {
		return nil
	}
	return conn
}

func (t *TCPMetricsExporter) start() {
	conn := t.connect()
	t.conn = conn
	ok := true

	if t.conn == nil {
		ok = false
	}

	established := make(chan struct{})
	reconnect := make(chan struct{}, 1)
	quit := make(chan struct{})
	exhausted := false

	for {

		select {
		case m := <-t.dataChan:
			// A message was recieved, and our TCP connection is "ok". Proceed.
			if ok {

				data, err := json.Marshal(m)
				if err != nil {
					Error.Println(err.Error())
					continue
				}
				_, err = t.conn.Write(data)
				_, err = t.conn.Write([]byte("\n"))

				// If error occurred on write, set flags to re-establish the connection
				if err != nil {
					ok = false
					Error.Println("lost connection")

				} else {
					exported.Inc()
				}
			} else {
				// A message was recieved from upstream, but was not written to the wire
				// due to conncetion errors; increment the dropped messages counter.
				dropped.Inc()
			}
		}

		// The second select block listens for a successful reconnect or a quit event.
		if !ok && !exhausted {
			select {
			// To avoid the attempt of reconnecting on every loop iteration, we are using a channel to
			// drop consecutive attempts to reconnect.
			case reconnect <- struct{}{}:
				go t.reconnect(reconnect, established, quit)
			case <-established:
				ok = true
				// The connection was successfully established, we can now free the reconnect channel to
				// allow future reconnects.
				<-reconnect
				break
			case <-quit:
				quit <- struct{}{}
				<-reconnect
				break
			default:
				continue
			}
		}

	}
}
