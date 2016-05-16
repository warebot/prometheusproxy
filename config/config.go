package config

import (
	proxy "github.com/warebot/prometheusproxy"
	"gopkg.in/yaml.v2"
	"io"
	"io/ioutil"
)

/*	Yaml config example:

	--
	port: 9191
	services:
	   service-a:
	     endpoint: http://localhost:9100/metrics
	     labels:
	       user: mina
	       source: proxy
	   service-b:
	     endpoint: http://localhost:9100/metrics
	     labels:
	       user: chuck
	       source: proxy

*/

// The Config struct encapsulates all configuration neccessarry to setup the proxy
type Config struct {
	Port        string
	Services    map[string]Service
	Subscribers map[string]map[string]interface{}
}

// The Service struct encapsulates the metric endpoints and optionals labels to be applied to metrics
type Service struct {
	Endpoint string
	Labels   map[string]string
}

func (c Config) BuildSubscribers() []proxy.Subscriber {
	subscribers := []proxy.Subscriber{}

	for name, config := range c.Subscribers {
		switch name {
		case "tcp_subscriber":
			destAddr := config["destaddr"].(string)
			concurrencyLevel := config["concurrency_level"].(int)
			subscribers = append(subscribers, proxy.NewTCPMetricsSubscriber(destAddr, concurrencyLevel))

		}

	}
	return subscribers
}

func ReadConfig(source io.Reader) (*Config, error) {
	rawConfig, err := ioutil.ReadAll(source)
	if err != nil {
		return nil, err
	}

	cfg := Config{}
	err = yaml.Unmarshal(rawConfig, &cfg)

	if err != nil {
		return nil, err
	}

	return &cfg, err
}
