package prometheusproxy

import (
	log "github.com/Sirupsen/logrus"
	helper "github.com/warebot/prometheusproxy/testhelpers"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"testing"
)

// Testing
func TestServeHTTP_GetUnknownService(t *testing.T) {
	// supresses loging.
	log.SetOutput(ioutil.Discard)
	scraper := NewHTTPScraper()
	proxy := NewPromProxy(scraper, NewRouter(), nil, nil)

	req, _ := http.NewRequest("GET", "http://fakeproxy?service=unknown", nil)
	w := httptest.NewRecorder()
	proxy.ServeHTTP(w, req)

	if w.Code != 400 {
		t.Error(helper.AssertError(400, w.Code))
	}
}
