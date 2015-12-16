package main

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
)

var (
	mux    *http.ServeMux
	server *httptest.Server
	client ScrapeClient
)

func setup() {
	mux = http.NewServeMux()
	server = httptest.NewServer(mux)
	config := &Config{
		Port: "9191",
		Services: map[string]Service{
			"bad-service": Service{Endpoint: server.URL},
		},
	}
	client = ScrapeClient{config: config}
}

func teardown() {
	server.Close()
}

func assertError(expected, actual interface{}) string {
	return fmt.Sprintf("Expected %s, got %s", expected, actual)
}

func TestScrapeClient_GetUnknownService(t *testing.T) {
	setup()
	_, err := client.scrape("unknown", nil)
	expectedError := UnknownService{}

	if err != expectedError {
		t.Fatalf(assertError(expectedError, err))
	}

}

// Testing the error handling for malfunctioning endpoints
func TestScrapeClient_BadService(t *testing.T) {
	setup()
	defer teardown()

	// Mocks a scrapable service endpoint that statically returns a 500 status code
	mux.HandleFunc("/",
		func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusInternalServerError)
		},
	)
	// The ScrapeClient should return (nil, RemoteServiceError) for unreachable or
	// malfunctioning endpoints
	mf, err := client.scrape("bad-service", nil)
	if err == nil {
		t.Fatalf(assertError(RemoteServiceError{}, err))
	}

	if mf != nil {
		t.Fatalf(assertError(nil, mf))
	}
}
