package prometheusproxy

import (
	"net/http"
	"net/url"
	"strings"
)

type InvalidURLErr struct {
	msg string
}

func (i InvalidURLErr) Error() string {
	return i.msg
}

type Endpoint struct {
	URL    *url.URL
	Labels map[string]string
}

type Router struct {
	endpoints map[string]Endpoint
}

func NewRouter() Router {
	return Router{endpoints: make(map[string]Endpoint)}
}
func (r Router) AddEndpoint(serviceName string, eURL string, labels map[string]string) error {
	endpointURL, err := url.Parse(eURL)
	if err != nil {
		return err
	}
	if !endpointURL.IsAbs() {
		return InvalidURLErr{"url must be absolute"}
	}

	r.endpoints[serviceName] = Endpoint{
		URL:    endpointURL,
		Labels: labels,
	}
	return nil
}

func (r Router) Route(req *http.Request) (*Endpoint, error) {
	queryParams := req.URL.Query()
	serviceName := queryParams.Get("service")
	labels := queryParams.Get("labels")

	if _, ok := r.endpoints[serviceName]; !ok {
		return nil, UnknownService{}
	}
	endpoint := r.endpoints[serviceName]
	// Merge adhoc-labels if any were provided
	if len(labels) > 0 {
		labelPairSet := strings.Split(labels, ",")
		for _, labelPair := range labelPairSet {
			labelKeyValue := strings.Split(labelPair, "|")
			if len(labelKeyValue) > 1 {
				endpoint.Labels[labelKeyValue[0]] = labelKeyValue[1]
			}
		}
	}
	return &endpoint, nil
}
