package main

import (
	"flag"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
)

var configFile = flag.String("config.file", "promproxy.yml", "proxy config flie")

const acceptHeader = `application/vnd.google.protobuf;proto=io.prometheus.client.MetricFamily;encoding=delimited;q=0.7,text/plain;version=0.0.4;q=0.3,application/json;schema="prometheus/telemetry";version=0.0.2;q=0.2,*/*;q=0.1`

func trapSignal(ch chan os.Signal) {
	signalType := <-ch

	Warning.Println(fmt.Sprintf("Caught [%v]", signalType))
	Warning.Println("Shutting down")

	signal.Stop(ch)
	os.Exit(0)
}

func main() {
	flag.Parse()

	cfg, err := readConfig(*configFile)

	if err != nil {
		panic(err.Error())
	}

	ch := make(chan os.Signal, 1)
	go trapSignal(ch)
	signal.Notify(ch, os.Interrupt, os.Kill, syscall.SIGTERM)

	handler := &PromProxy{Config: cfg}
	http.Handle("/metrics", handler)

	Info.Println("Starting proxy service on port", cfg.Port)
	if err = http.ListenAndServe(":"+cfg.Port, nil); err != nil {
		Error.Fatalf("Failed to start the proxy service: %v", err.Error())
	}

}
