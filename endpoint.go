package prometheusproxy

import (
	"github.com/warebot/prometheusproxy/config"
	"net/http"
	"net/url"
	"strings"
)

type Endpoint struct {
	Owner  string
	URL    *url.URL
	Labels map[string]string
}

func NewRequestEndpoint(req *http.Request, c *config.Config) (*Endpoint, error) {
	queryParams := req.URL.Query()
	serviceName := queryParams.Get("service")
	labels := queryParams.Get("labels")
	owner := queryParams.Get("owner")

	service, ok := c.Services[serviceName]
	if !ok {
		return nil, UnknownService{}
	}

	endpointURL, err := url.Parse(service.Endpoint)
	if err != nil {
		return nil, err
	}

	// Merge adhoc-labels if any were provided
	if len(labels) > 0 {
		labelPairSet := strings.Split(labels, ",")
		for _, labelPair := range labelPairSet {
			labelKeyValue := strings.Split(labelPair, "|")
			if len(labelKeyValue) > 1 {
				service.Labels[labelKeyValue[0]] = labelKeyValue[1]
			}
		}
	}

	return &Endpoint{URL: endpointURL,
		Labels: service.Labels,
		Owner:  owner,
	}, nil
}
