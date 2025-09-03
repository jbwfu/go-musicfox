package app

// Event is the top-level interface for all application events.
// It ensures type safety within the event channel.
type Event interface {
	isEvent()
}

// AppBusyEvent is published when a long-running, user-initiated operation starts.
// The UI should react to this by showing a loading indicator.
type AppBusyEvent struct {
	Message string
}

func (e AppBusyEvent) isEvent() {}

// AppIdleEvent is published when a long-running, user-initiated operation finishes.
// The UI should react to this by hiding the loading indicator.
type AppIdleEvent struct{}

func (e AppIdleEvent) isEvent() {}
