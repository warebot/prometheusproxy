package main

import (
	"gopkg.in/yaml.v2"
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
	Port     string
	Services map[string]Service
}

// The Service struct encapsulates the metric endpoints and optionals labels to be applied to metrics
type Service struct {
	Endpoint string
	Labels   map[string]string
}

func readConfig(file string) (*Config, error) {

	rawConfig, err := ioutil.ReadFile(*configFile)
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
