package main

import (
	"encoding/json"
	"fmt"
	dto "github.com/prometheus/client_model/go"
	"net"
	"time"
)

type MetricsExporter interface {
	start()
}

type TCPMetricsExporter struct {
	destAddr string
	dataChan chan *dto.MetricFamily
	conn     net.Conn
}

func (t *TCPMetricsExporter) reconnect(established chan struct{}, quit chan struct{}) {
	// Some type of backoff algorithm
	seed := 5
	for i := 1; i < 20000; i++ {
		conn, err := net.Dial("tcp", t.destAddr)
		if err == nil {
			t.conn = conn
			established <- struct{}{}
		} else {
			fmt.Println(err.Error())
		}
		time.Sleep(time.Second * time.Duration(i*seed))
	}
	quit <- struct{}{}

}

func (t *TCPMetricsExporter) connect() net.Conn {
	conn, err := net.Dial("tcp", t.destAddr)
	if err != nil {
		panic(err.Error())
	}
	return conn
}

func (t *TCPMetricsExporter) start() {
	conn := t.connect()
	t.conn = conn
	ok := true

	established := make(chan struct{})
	quit := make(chan struct{})
	exhausted := false

	for {
		m := <-t.dataChan
		data, _ := json.Marshal(m)
		_, err := t.conn.Write(data)
		_, err = t.conn.Write([]byte("\n"))

		if err != nil && !exhausted {
			Error.Println(err.Error())
			go t.reconnect(established, quit)
			ok = false

			for {
				fmt.Println("Draining")
				select {
				case <-t.dataChan:
				default:
				}

				select {
				case <-established:
					ok = true
					break
				case <-quit:
					quit <- struct{}{}
					exhausted = true
					break
				default:
					continue
				}
				if ok {
					break
				}
			}

		}

	}
}
