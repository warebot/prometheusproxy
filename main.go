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
	configFile     = flag.String("config.file", "promproxy.yml", "Proxy config file")
	destAddr       = flag.String("dest.addr", "", "Destination host for tcp connection")
	validateConfig = flag.Bool("validate", false, "Validate config only. Do not start service")

	dropped = prometheus.NewCounter(prometheus.CounterOpts{
		Namespace: "prometheus",
		Subsystem: "proxy",
		Name:      "metrics_messages_dropped",
		Help:      "The number of metric family messages dropped.",
	})

	exported = prometheus.NewCounter(prometheus.CounterOpts{
		Namespace: "prometheus",
		Subsystem: "proxy",
		Name:      "metrics_messages_exported",
		Help:      "The number of metric family messages exported.",
	})
)

// acceptHeader is used in content type negotiations - Used by Promethes endpoints that expose metrics.
const acceptHeader = `application/vnd.google.protobuf;proto=io.prometheus.client.MetricFamily;encoding=delimited;q=0.7,text/plain;version=0.0.4;q=0.3,application/json;schema="prometheus/telemetry";version=0.0.2;q=0.2,*/*;q=0.1`

func trapSignal(ch chan os.Signal) {
	signalType := <-ch

	Warning.Println(fmt.Sprintf("Caught [%v]", signalType))
	Warning.Println("Shutting down")

	signal.Stop(ch)
	os.Exit(0)
}

// infoHandler exposes meta-data related to the build of the application via HTTP endpoint.
func infoHandler(w http.ResponseWriter, req *http.Request) {
	encoder := json.NewEncoder(w)
	encoder.Encode(version.Map)
}

// validateAndExit is a feature requested by Nik to validate settings in a config file
// and exit on failure, allowing fail fast abilities.
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
		// Parse endpoints to validate that they are in fact valid HTTP URL endpoints;
		// fail if not.
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

	// dataChan is a channel used by the `Proxy` to pipe metric families to the downstream
	// metrics exporter implementation.
	dataChan := make(chan Message, 1000)
	cfg, err := readConfig(*configFile)

	if err != nil {
		panic(err.Error())
	}

	if *validateConfig {
		validateAndExit(cfg)
	}

	// Logic for graceful exits via interrupts
	// TODO Might need some cleanup.
	ch := make(chan os.Signal, 1)
	go trapSignal(ch)
	signal.Notify(ch, os.Interrupt, os.Kill, syscall.SIGTERM)

	// client is a concrete ScrapeClient instance that performs the actual work of
	// scraping metric endpoints.
	// TODO
	// Probably should be an interface to facilitate testing?
	client := ScrapeClient{config: cfg}
	shouldFlush := false

	// destAddr is a "host:port" address variable representing the TCP endpoint to connect to
	// for metrics export.
	if len(*destAddr) > 0 {
		shouldFlush = true
		fmt.Println("Will flush")
	}

	// HTTP handler that is responsible for handling metrics scrape requests
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

func init() {
	prometheus.MustRegister(dropped)
	prometheus.MustRegister(exported)
}
