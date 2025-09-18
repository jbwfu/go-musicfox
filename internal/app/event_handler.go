package app

import (
	"log"
)

// TeaModel defines the interface that the UI model must implement
// to receive events from the AppEventHandler. This allows the core
// to communicate with the UI without depending on its concrete implementation.
type TeaModel interface {
	Send(msg interface{})
}

// AppEventHandler listens to the central event channel and dispatches events
// to appropriate services and the UI. It acts as the central event hub,
// decoupling event producers from consumers.
type AppEventHandler struct {
	uiModel TeaModel
	// Other services that need to react to events will be added here, e.g.,
	// reporterSvc *services.Reporter
	// historySvc  *services.HistoryService
}

// NewAppEventHandler creates a new central event handler.
// It requires all event-consuming services to be injected.
func NewAppEventHandler(uiModel TeaModel) *AppEventHandler {
	return &AppEventHandler{
		uiModel: uiModel,
	}
}

// StartListening starts a goroutine to listen for events from the central channel.
// This should be called once at application startup.
func (h *AppEventHandler) StartListening(eventChan <-chan Event) {
	go func() {
		for event := range eventChan {
			if err := h.dispatchEvent(event); err != nil {
				// In a real application, use a structured logger.
				// For now, logging to stderr is sufficient.
				log.Printf("Error dispatching event: %v\n", err)
			}
		}
	}()
}

// dispatchEvent routes a single event to the correct handlers.
// For now, it only forwards events to the UI. In the future, it will
// route specific events to other services as well.
func (h *AppEventHandler) dispatchEvent(event Event) error {
	// Always forward all events to the UI model.
	// The UI's own Update function will decide what to do with them.
	if h.uiModel != nil {
		h.uiModel.Send(event)
	}

	// In the future, we will add routing to other services here.
	// Example:
	// switch e := event.(type) {
	// case player.SongChangedEvent:
	// 	if h.reporterSvc != nil {
	// 		h.reporterSvc.HandleSongChanged(e)
	// 	}
	// }

	return nil
}
