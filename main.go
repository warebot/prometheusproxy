package main

import (
	"bytes"
	"flag"
	"io/ioutil"
	"net/http"
	_ "net/url"
	"strconv"
	"net/http/httputil"
	"net/url"
	"regexp"
	"fmt"
)

var targetUrl = flag.String("target", "", "target url")
var labels = flag.String("labels", "", "default labels")

type transport struct {
	http.RoundTripper
}

var client = &http.Client{Transport:http.DefaultTransport}

func (t *transport) RoundTrip(req *http.Request) (resp *http.Response, err error) {
	hasLabelsPattern := regexp.MustCompile("\\{([^{]+)\\}")
	noLabelsPattern := regexp.MustCompile("(\\w+)\\s\\d")
	resp, err = client.Get(req.URL.String())
	b, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	newMetrics := hasLabelsPattern.ReplaceAllString(string(b), fmt.Sprintf("{${1},%s}", *labels))
	content := []byte(noLabelsPattern.ReplaceAllString(newMetrics, fmt.Sprintf("${1}{%s} ", *labels)))

	body := ioutil.NopCloser(bytes.NewReader(content))
	resp.Body = body
	resp.ContentLength = int64(len(content))
	resp.Header.Set("Content-Length", strconv.Itoa(len(content)))
	return resp, nil
}

//var _ http.RoundTripper = &transport{}


func main() {
	flag.Parse()
	target, err := url.Parse(*targetUrl)

	if err != nil {
		panic(err)
	}
	proxy := httputil.NewSingleHostReverseProxy(target)
	proxy.Transport = &transport{http.DefaultTransport}
	http.Handle("/", proxy)
	http.ListenAndServe(":9191", nil)

}