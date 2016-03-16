package main

import (
	"encoding/json"
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
}

func (t TCPMetricsExporter) connect() net.Conn {
	// Some type of backoff algorithm
	seed := 5
	for i := 1; i < 20000; i++ {
		conn, err := net.Dial("tcp", t.destAddr)
		if err == nil {
			return conn
		}
		time.Sleep(time.Second * time.Duration(i*seed))
	}
	return nil
}

func (t TCPMetricsExporter) start() {
	conn := t.connect()

	if conn == nil {
		Error.Println("failed to establish tcp connection: exhausted retries")
		return
	}

	for {
		m := <-t.dataChan
		data, _ := json.Marshal(m)
		_, err := conn.Write(data)
		_, err = conn.Write([]byte("\n"))
		if err != nil {
			Error.Println(err.Error())
			conn = t.connect()
		}
	}
}
