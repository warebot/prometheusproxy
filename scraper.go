package prometheusproxy

import (
	"fmt"
	dto "github.com/prometheus/client_model/go"

	"github.com/golang/protobuf/proto"
	"github.com/prometheus/common/expfmt"
	"net/http"
)

type Scraper interface {
	Messages() chan *dto.MetricFamily
	Errors() chan error
	Scrape(Endpoint) (chan *dto.MetricFamily, chan error, error)
}

// HTTP Implementation

const (
	acceptHeader = `application/vnd.google.protobuf;proto=io.prometheus.client.MetricFamily;encoding=delimited;q=0.7,text/plain;version=0.0.4;q=0.3,application/json;schema="prometheus/telemetry";version=0.0.2;q=0.2,*/*;q=0.1`
)

type HTTPScraper struct {
	client   *http.Client
	messages chan *dto.MetricFamily
	errors   chan error
}

func NewHTTPScraper() *HTTPScraper {
	return &HTTPScraper{
		client:   &http.Client{},
		messages: make(chan *dto.MetricFamily, 10),
		errors:   make(chan error),
	}
}

func (hs *HTTPScraper) Scrape(endpoint Endpoint) (chan *dto.MetricFamily, chan error, error) {
	req, err := http.NewRequest("GET", endpoint.URL.String(), nil)
	if err != nil {
		return nil, nil, err
	}

	out := make(chan *dto.MetricFamily, 10)
	errors := make(chan error)

	go func() {
		defer func() {
			close(out)
			close(errors)
		}()

		req.Header.Add("Accept", acceptHeader)
		resp, err := hs.client.Do(req)
		if err != nil {
			errors <- err
			return
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			errors <- fmt.Errorf("server returned HTTP status %s", resp.Status)
			return
		}

		// NewDecoder is part of the prometheus client_go library responsible for returning a data
		// decoder with respect to the Content-Encoding/Type headers.
		dec := expfmt.NewDecoder(resp.Body, expfmt.ResponseFormat(resp.Header))

		for {
			var d *dto.MetricFamily = &dto.MetricFamily{}
			if err = dec.Decode(d); err != nil {
				break
			}

			// Get the pre-configured label pairs from the config for the service name being queried.

			for _, metric := range d.Metric {
				for k, v := range endpoint.Labels {
					metric.Label = append(metric.Label, &dto.LabelPair{
						Name:  proto.String(k),
						Value: proto.String(v)})
				}
			}
			out <- d
		}
	}()
	fmt.Println("exiting")
	return out, errors, nil
}

func (hs *HTTPScraper) Messages() chan *dto.MetricFamily {
	return hs.messages
}

func (hs *HTTPScraper) Errors() chan error {
	return hs.errors
}
