package testhelpers

import (
	"fmt"
	"net/http"
	"net/http/httptest"
)

func assertError(expected, actual interface{}) string {
	return fmt.Sprintf("Expected %v, got %v", expected, actual)
}

type HTTPServer struct {
	mux    *http.ServeMux
	server *httptest.Server
}

func NewHTTPServer() HTTPServer {
	mux := http.NewServeMux()
	return HTTPServer{server: httptest.NewServer(mux), mux: mux}
}

func (hs *HTTPServer) Register(path string, f http.HandlerFunc) {
	hs.mux.HandleFunc(path, f)
}

func (hs *HTTPServer) ShutDown() {
	hs.server.Close()
}

func (hs *HTTPServer) URL() string {
	return hs.server.URL
}
