package player

import (
	"time"

	"github.com/go-musicfox/go-musicfox/internal/app"
	"github.com/go-musicfox/go-musicfox/internal/structs"
	"github.com/go-musicfox/go-musicfox/internal/types"
)

// Service defines the public API for the player core.
// It is responsible for managing playback logic, playlists, and state,
// but is completely decoupled from the UI. The rest of the application
// will interact with the player through this interface.
type Service interface {
	// Playback control
	PlaySong(song structs.Song)
	Pause()
	Resume()
	Toggle()
	Stop()
	Seek(duration time.Duration)
	NextSong()
	PreviousSong()

	// Volume control
	SetVolume(volume int)
	UpVolume()
	DownVolume()

	// State retrieval
	CurrentSong() structs.Song
	Playlist() []structs.Song
	State() types.State
	PassedTime() time.Duration

	// Close shuts down the service and releases resources.
	Close()
}

// serviceImpl is the concrete, unexported implementation of the Service interface.
type serviceImpl struct {
	// eventChan is a write-only channel to send events out to listeners.
	eventChan chan<- app.Event

	// Other dependencies and state will be added here in subsequent commits.
}

// NewService creates a new player service.
// It requires an event channel to be injected for publishing events.
func NewService(eventChan chan<- app.Event) Service {
	s := &serviceImpl{
		eventChan: eventChan,
	}
	// In the future, we would start internal goroutines for the player here.
	return s
}

// Close closes the event channel and cleans up resources.
func (s *serviceImpl) Close() {
	// In a real implementation, we might need to stop background goroutines first.
	close(s.eventChan)
}

// publishEvent is a helper to send an event to the listener in a non-blocking way.
func (s *serviceImpl) publishEvent(e app.Event) {
	select {
	case s.eventChan <- e:
	default:
		// This case prevents blocking if the channel buffer is full
		// or if there's no receiver. It's a safeguard.
	}
}

// Empty implementations to satisfy the interface contract.
// Logic will be migrated into these methods in later commits.

func (s *serviceImpl) PlaySong(song structs.Song)  {}
func (s *serviceImpl) Pause()                      {}
func (s *serviceImpl) Resume()                     {}
func (s *serviceImpl) Toggle()                     {}
func (s *serviceImpl) Stop()                       {}
func (s *serviceImpl) Seek(duration time.Duration) {}
func (s *serviceImpl) NextSong()                   {}
func (s *serviceImpl) PreviousSong()               {}
func (s *serviceImpl) SetVolume(volume int)        {}
func (s *serviceImpl) UpVolume()                   {}
func (s *serviceImpl) DownVolume()                 {}
func (s *serviceImpl) CurrentSong() structs.Song {
	return structs.Song{}
}
func (s *serviceImpl) Playlist() []structs.Song {
	return nil
}
func (s *serviceImpl) State() types.State {
	return types.Stopped
}
func (s *serviceImpl) PassedTime() time.Duration {
	return 0
}
