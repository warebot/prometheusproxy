package prometheusproxy

// Subscriber is an interface to allow various concrete implementations of our
// downstream exporting functionality.
type Subscriber interface {
	Chan() chan Message
	Start()
	Name() string
}
