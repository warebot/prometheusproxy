package prometheusproxy

import (
	"github.com/warebot/prometheusproxy/config"
	th "github.com/warebot/prometheusproxy/testhelpers"

	"net/http"
	"testing"
)

func TestRequestEndpoint_Valid(t *testing.T) {
	httpServer := th.NewHTTPServer()
	defer httpServer.ShutDown()

	config := &config.Config{
		Port: "9191",
		Services: map[string]config.Service{
			"awesome-service": config.Service{Endpoint: httpServer.URL()},
		},
	}

	req, err := http.NewRequest("GET", httpServer.URL()+"/?service=awesome-service", nil)
	_, err = NewRequestEndpoint(req, config)
	if err != nil {
		t.Fatal(assertError(nil, err))
	}

}

func TestRequestEndpoint_InValid(t *testing.T) {
	httpServer := th.NewHTTPServer()
	defer httpServer.ShutDown()

	config := &config.Config{
		Port: "9191",
		Services: map[string]config.Service{
			"awesome-service": config.Service{Endpoint: httpServer.URL()},
		},
	}

	req, err := http.NewRequest("GET", "/service=awesome-service", nil)
	_, err = NewRequestEndpoint(req, config)
	e := UnknownService{}
	if e != err {
		t.Fatal(assertError(nil, err))
	}

}
