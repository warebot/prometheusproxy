package prometheusproxy

import (
	"github.com/prometheus/client_golang/prometheus"
)

// Subscriber is an interface to allow various concrete implementations of our
// downstream exporting functionality.
type Subscriber interface {
	// Chan returns the channel that the subscriber is listening on.
	Chan() chan Message
	// Start starts the activity loop.
	Start(exported, dropped *prometheus.CounterVec)
	// Name returns a human readable string describing the subscriber.
	Name() string
	Equals(o Subscriber) bool
}
