package config

import (
	"bytes"
	proxy "github.com/warebot/prometheusproxy"
	"reflect"
	"testing"
)

var testConfig = `---
port: 9191
services:
   service-a:
     endpoint: http://localhost:9100/metrics
     labels:
       user: mina
       source: proxy.bob
       host: 10.2.3.4
   service-b:
     endpoint: http://localhost:9100/metrics
     labels:
       user: chuck
       source: proxy.bob
       host: 10.1.2.3S
subscribers:
     tcp_subscriber:
        concurrency_level: 1
        destaddr: localhost:9777 
`

func TestBuildSubscribers(t *testing.T) {
	c := bytes.NewBufferString(testConfig)
	cfg, err := ReadConfig(c)
	if err != nil {
		t.Error(err)
	}

	subscribers := cfg.BuildSubscribers()
	if subscribers == nil {
		t.Error("failed to parse subscribers")
	}

	expectedSubscribers := map[string]proxy.Subscriber{
		"tcp_subscriber": proxy.NewTCPMetricsSubscriber("localhost:9777", 1),
	}

	for _, subscriber := range subscribers {
		if reflect.TypeOf(subscriber) != reflect.TypeOf(expectedSubscribers[subscriber.Name()]) {
			t.Fatal("invalid subscriber type")
		}

		if !subscriber.Equals(expectedSubscribers[subscriber.Name()]) {
			t.Fatal("inconsistent subscriber configuration")
		}
		t.Log(subscriber)
	}
}
