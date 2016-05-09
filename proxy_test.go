package prometheusproxy

import (
	"fmt"
	"github.com/warebot/prometheusproxy/config"
	th "github.com/warebot/prometheusproxy/testhelpers"

	"net/http"
	"testing"
)

func assertError(expected, actual interface{}) string {
	return fmt.Sprintf("Expected %v, got %v", expected, actual)
}

func TestScrapeClient_GetUnknownService(t *testing.T) {
	//client := NewHTTPScraper()
	httpServer := th.NewHTTPServer()
	defer httpServer.ShutDown()

	config := &config.Config{
		Port: "9191",
		Services: map[string]config.Service{
			"bad-service": config.Service{Endpoint: httpServer.URL()},
		},
	}

	req, err := http.NewRequest("GET", httpServer.URL()+"/?service=unknown", nil)
	_, err = NewRequestEndpoint(req, config)
	expectedError := UnknownService{}
	if err != expectedError {
		t.Fatal(assertError(expectedError, err))
	}

}

// Testing the error handling for malfunctioning endpoints
func TestScrapeClient_BadService(t *testing.T) {
	httpServer := th.NewHTTPServer()
	defer httpServer.ShutDown()

	config := &config.Config{
		Port: "9191",
		Services: map[string]config.Service{
			"bad-service": config.Service{Endpoint: httpServer.URL()},
		},
	}

	// Mocks a scrapable service endpoint that statically returns a 500 status code
	httpServer.Register("/",
		func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusInternalServerError)
		},
	)
	// The ScrapeClient should return (nil, RemoteServiceError) for unreachable or
	// malfunctioning endpoints

	req, err := http.NewRequest("GET", httpServer.URL()+"/?service=bad-service", nil)
	if err != nil {
		t.Fatal(err)
	}

	endpoint, err := NewRequestEndpoint(req, config)
	if err != nil {
		t.Fatal(err)
	}
	client := NewHTTPScraper()
	_, errors, err := client.Scrape(endpoint)

	// Excpect 1 error
	errCnt := 0
	for _ = range errors {
		errCnt++
	}

	if errCnt != 1 {
		t.Error(assertError(1, errCnt))
	}
}
