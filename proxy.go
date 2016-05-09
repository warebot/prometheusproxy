package prometheusproxy

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/common/expfmt"
	"github.com/warebot/prometheusproxy/config"
	"io"
	"net/http"
)

// Our domain-specific errors
type UnknownService struct {
	msg string
}

func (e UnknownService) Error() string {
	return "unknown service type"
}

type RemoteServiceError struct {
	msg string
}

func (e RemoteServiceError) Error() string {
	return e.msg
}

type PromProxy struct {
	flush             bool
	client            Scraper
	subcribers        []Subscriber
	config            *config.Config
	dropped, exported prometheus.Counter
}

func (p *PromProxy) AddSubscriber(s Subscriber) {
	p.subcribers = append(p.subcribers, s)
}

func NewPromProxy(client Scraper, config *config.Config,
	exported, dropped prometheus.Counter) *PromProxy {
	return &PromProxy{
		client:   client,
		config:   config,
		exported: exported,
		dropped:  dropped,
	}
}

func (p *PromProxy) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	endpoint, err := NewRequestEndpoint(req, p.config)
	if err != nil {
		Error.Printf("%v\n", err.Error())
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(err.Error()))
		return
	}

	messages, errors, err := p.client.Scrape(endpoint)
	if err != nil && err != io.EOF {
		Error.Printf("%v\n", err.Error())
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(err.Error()))
		return
	}

	// Negotiates content-type based on accept headers and creates the encoder accordingly
	contentType := expfmt.Negotiate(req.Header)
	// Set the Content type header based on the negotiated accept header negotiation
	w.Header().Set("Content-type", string(contentType))
	encoder := expfmt.NewEncoder(w, contentType)
	if err != nil {
		if err != io.EOF {
			Error.Printf("%\nv", err.Error())
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte(err.Error()))
			return
		}
	}

	go func() {
		for e := range errors {
			Error.Println(e)
		}
	}()

	for m := range messages {
		if err := encoder.Encode(m); err != nil {
			Error.Printf("%v\n", err.Error())
			w.Write([]byte(err.Error()))
		}

		for _, s := range p.subcribers {
			select {
			case s.Chan() <- Message{Payload: m, Owner: endpoint.Owner}:
			default:
				p.dropped.Inc()
			}
		}
	}
}
