package prometheusproxy

import (
	helper "github.com/warebot/prometheusproxy/testhelpers"
	"net/http"
	"testing"
)

func TestAddEndpoint_ValidEndpoint(t *testing.T) {
	router := NewRouter()
	err := router.AddEndpoint("the-service", "http://localhost/metrics", nil)
	if err != nil {
		t.Error(helper.AssertError(nil, err))
	}
}

func TestAddEndpoint_InvalidEndpoint(t *testing.T) {
	// All endpoint URLs must be absolute
	router := NewRouter()
	err := router.AddEndpoint("the-service", "localhost/metrics", nil)

	expectedErr := InvalidURLErr{"url must be absolute"}
	if err != expectedErr {
		t.Error(helper.AssertError(expectedErr, err))
	}
}

func TestRoute_ValidService(t *testing.T) {
	router := NewRouter()
	router.AddEndpoint("the-service", "http://localhost:8080/metrics", nil)
	req, _ := http.NewRequest("GET", "http://localhost/metrics?service=the-service", nil)
	endpoint, err := router.Route(req)

	if err != nil {
		t.Error(helper.AssertError(nil, err))
	}

	computed := endpoint.URL.String()
	expected := "http://localhost:8080/metrics"
	if computed != expected {
		t.Error(helper.AssertError(expected, computed))
	}
}

func TestRoute_UnknownService(t *testing.T) {
	router := NewRouter()
	router.AddEndpoint("the-service", "http://localhost:8080/metrics", nil)
	req, _ := http.NewRequest("GET", "http://localhost/metrics?service=the-other-service", nil)
	_, err := router.Route(req)

	expectedErr := UnknownService{}
	if err != expectedErr {
		t.Error(helper.AssertError(expectedErr, err))
	}
}
