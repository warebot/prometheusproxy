package main

import (
	"encoding/json"
	"flag"
	"github.com/prometheus/client_golang/prometheus"
	p "github.com/warebot/prometheusproxy"
	"github.com/warebot/prometheusproxy/config"
	"github.com/warebot/prometheusproxy/version"
	"net/http"
	"net/url"
	"os"
	"strconv"
)

var (
	configFile = flag.String("config.file", "promproxy.yml", "Proxy config file")
	//tcpDestAddr        = flag.String("tcp.dest-addr", "", "Destination host for tcp subscriber connection")
	//tcpConurrencyLevel = flag.String("tcp.workers", "", "Number of workers to start for the tcp subscriber")
	validateConfig = flag.Bool("validate", false, "Validate config only. Do not start service")

	dropped = prometheus.NewCounterVec(prometheus.CounterOpts{
		Namespace: "prometheus",
		Subsystem: "proxy",
		Name:      "metrics_messages_dropped",
		Help:      "The number of metric family messages dropped.",
	},

		[]string{"subscriber"})

	exported = prometheus.NewCounterVec(prometheus.CounterOpts{
		Namespace: "prometheus",
		Subsystem: "proxy",
		Name:      "metrics_messages_exported",
		Help:      "The number of metric family messages exported.",
	},
		[]string{"subscriber"})
)

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
	p.Logger.Infoln("Initializing service")
	p.Logger.Infoln("Version =>", version.Version)
	p.Logger.Infoln("Revision =>", version.Revision)
	p.Logger.Infoln("Build date =>", version.BuildDate)

	flag.Parse()

	f, err := os.Open(*configFile)
	if err != nil {
		panic(err.Error())
	}

	cfg, err := config.ReadConfig(f)
	if err != nil {
		panic(err.Error())
	}

	if *validateConfig {
		validateAndExit(cfg)
	}

	router := p.NewRouter()
	for name, service := range cfg.Services {
		router.AddEndpoint(name, service.Endpoint, service.Labels)
	}

	// scraper implements the Scraper interface and is responsible for fetching data.
	scraper := p.NewHTTPScraper()

	// handler is responsible for handling metrics scrape requests
	handler := p.NewPromProxy(scraper, router, exported, dropped)

	// Configure subscribers from config file.
	subscribers := cfg.BuildSubscribers()
	for _, s := range subscribers {
		p.Logger.Infoln("adding subscriber", s.Name())
		handler.AddSubscriber(s)
		go s.Start(exported, dropped)
	}

	http.Handle("/", prometheus.Handler())
	http.Handle("/metrics", handler)
	http.HandleFunc("/info", infoHandler)

	p.Logger.Infoln("starting proxy service on port", cfg.Port)
	if err = http.ListenAndServe(":"+cfg.Port, nil); err != nil {
		p.Logger.Fatalf("Failed to start the proxy service: %v", err.Error())
	}

}

func init() {
	prometheus.MustRegister(dropped)
	prometheus.MustRegister(exported)
}
