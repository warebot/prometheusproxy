package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/warebot/prometheusproxy/version"
	"net/http"
	"net/url"
	"os"
	"os/signal"
	"strconv"
	"syscall"
)

var (
	configFile     = flag.String("config.file", "promproxy.yml", "Proxy config flie")
	destAddr       = flag.String("dest.addr", "", "Destination host for tcp connection")
	validateConfig = flag.Bool("validate", false, "Validate config only. Do not start service")
)

const acceptHeader = `application/vnd.google.protobuf;proto=io.prometheus.client.MetricFamily;encoding=delimited;q=0.7,text/plain;version=0.0.4;q=0.3,application/json;schema="prometheus/telemetry";version=0.0.2;q=0.2,*/*;q=0.1`

func trapSignal(ch chan os.Signal) {
	signalType := <-ch

	Warning.Println(fmt.Sprintf("Caught [%v]", signalType))
	Warning.Println("Shutting down")

	signal.Stop(ch)
	os.Exit(0)
}

func infoHandler(w http.ResponseWriter, req *http.Request) {
	encoder := json.NewEncoder(w)
	encoder.Encode(version.Map)
}

func validateAndExit(cfg *Config) {
	if len(cfg.Port) == 0 {
		os.Exit(1)
	}
	if _, err := strconv.ParseInt(cfg.Port, 10, 64); err != nil {
		os.Exit(1)
	}
	for _, service := range cfg.Services {
		if len(service.Endpoint) == 0 {
			os.Exit(1)
		}
		if _, err := url.Parse(service.Endpoint); err != nil {
			os.Exit(1)
		}
	}
	os.Exit(0)
}

func main() {
	Info.Println("Initializing service")
	Info.Println("Version =>", version.Version)
	Info.Println("Revision =>", version.Revision)
	Info.Println("Build date =>", version.BuildDate)

	flag.Parse()

	dataChan := make(chan Message, 1000)
	cfg, err := readConfig(*configFile)

	if err != nil {
		panic(err.Error())
	}

	if *validateConfig {
		validateAndExit(cfg)
	}

	ch := make(chan os.Signal, 1)
	go trapSignal(ch)
	signal.Notify(ch, os.Interrupt, os.Kill, syscall.SIGTERM)

	client := ScrapeClient{config: cfg}
	shouldFlush := false

	if len(*destAddr) > 0 {
		shouldFlush = true
		fmt.Println("Will flush")
	}

	handler := &PromProxy{client: client, out: dataChan, flush: shouldFlush}
	if shouldFlush {
		tcpServer := TCPMetricsExporter{dataChan: dataChan, destAddr: *destAddr}
		go tcpServer.start()
	}

	http.Handle("/", prometheus.Handler())
	http.Handle("/metrics", handler)
	http.HandleFunc("/info", infoHandler)

	Info.Println("starting proxy service on port", cfg.Port)
	if err = http.ListenAndServe(":"+cfg.Port, nil); err != nil {
		Error.Fatalf("Failed to start the proxy service: %v", err.Error())
	}

}
