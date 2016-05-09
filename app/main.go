package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"github.com/prometheus/client_golang/prometheus"
	p "github.com/warebot/prometheusproxy"
	"github.com/warebot/prometheusproxy/config"
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
func trapSignal(ch chan os.Signal) {
	signalType := <-ch

	p.Warning.Println(fmt.Sprintf("Caught [%v]", signalType))
	p.Warning.Println("Shutting down")

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
func validateAndExit(cfg *config.Config) {
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
	p.Info.Println("Initializing service")
	p.Info.Println("Version =>", version.Version)
	p.Info.Println("Revision =>", version.Revision)
	p.Info.Println("Build date =>", version.BuildDate)

	flag.Parse()

	// dataChan is a channel used by the `Proxy` to pipe metric families to the downstream
	// metrics exporter implementation.
	cfg, err := config.ReadConfig(*configFile)

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
	scraper := p.NewHTTPScraper()

	// HTTP handler that is responsible for handling metrics scrape requests
	handler := p.NewPromProxy(scraper, cfg, exported, dropped)

	// destAddr is a "host:port" address variable representing the TCP endpoint to connect to
	// for metrics export.
	if len(*destAddr) > 0 {
		subscriber := p.NewTCPMetricsExporter(*destAddr, exported, dropped)
		handler.AddSubscriber(subscriber)
		go subscriber.Start()
	}

	http.Handle("/", prometheus.Handler())
	http.Handle("/metrics", handler)
	http.HandleFunc("/info", infoHandler)

	p.Info.Println("starting proxy service on port", cfg.Port)
	if err = http.ListenAndServe(":"+cfg.Port, nil); err != nil {
		p.Error.Fatalf("Failed to start the proxy service: %v", err.Error())
	}

}

func init() {
	prometheus.MustRegister(dropped)
	prometheus.MustRegister(exported)
}
