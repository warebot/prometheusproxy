package prometheusproxy

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/common/expfmt"
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
	client            Scraper
	router            Router
	subcribers        []Subscriber
	dropped, exported *prometheus.CounterVec
}

// AddSubscriber adds a new Subscriber implementation to PromProxy used for fanning out metrics.
func (p *PromProxy) AddSubscriber(s Subscriber) {
	p.subcribers = append(p.subcribers, s)
}

// NewPromProxy creates a PromProxy instance.
func NewPromProxy(client Scraper, router Router,
	exported, dropped *prometheus.CounterVec) *PromProxy {
	return &PromProxy{
		client:   client,
		router:   router,
		exported: exported,
		dropped:  dropped,
	}
}

func (p *PromProxy) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	endpoint, err := p.router.Route(req)
	if err != nil {

		Logger.Errorf("%v\n", err.Error())
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(err.Error()))
		return
	}

	owner := req.FormValue("owner")

	// Scrape returns a messages channel, error channel and an explicit error for fatalistic errors.
	messages, errors, err := p.client.Scrape(*endpoint)
	if err != nil && err != io.EOF {
		Logger.Errorf("%v\n", err.Error())
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(err.Error()))
		return
	}

	// Negotiates content-type based on accept headers and creates the encoder accordingly
	contentType := expfmt.Negotiate(req.Header)

	// Set the Content type header based on the negotiated accept header negotiation
	w.Header().Set("Content-type", string(contentType))
	encoder := expfmt.NewEncoder(w, contentType)
	go func() {
		for e := range errors {
			Logger.Errorln(e)
		}
	}()

	for m := range messages {
		if err := encoder.Encode(m); err != nil {
			Logger.Errorf("%v\n", err.Error())
			w.Write([]byte(err.Error()))
		}

		for _, s := range p.subcribers {
			select {
			case s.Chan() <- Message{Payload: m, Owner: owner}:
			default:
				p.dropped.WithLabelValues(s.Name()).Inc()
			}
		}
	}
}
