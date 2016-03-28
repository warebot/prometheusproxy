package main

import (
	"encoding/json"
	dto "github.com/prometheus/client_model/go"
	"net"
	"time"
)

type Message struct {
	Owner   string            `json:"owner,omitempty"`
	Payload *dto.MetricFamily `json:"payload"`
}

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
			if ok {

				data, _ := json.Marshal(m)
				_, err := t.conn.Write(data)
				_, err = t.conn.Write([]byte("\n"))
				if err != nil {
					ok = false
					Error.Println("lost connection")

				} else {
					exported.Inc()
				}
			} else {
				dropped.Inc()
			}
		}

		if !ok && !exhausted {
			select {
			case reconnect <- struct{}{}:
				go t.reconnect(reconnect, established, quit)
			case <-established:
				ok = true
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
