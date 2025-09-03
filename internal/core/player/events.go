package player

import (
	"time"

	"github.com/go-musicfox/go-musicfox/internal/structs"
	"github.com/go-musicfox/go-musicfox/internal/types"
)

// StateChangedEvent is published when the player's state changes (e.g., Playing, Paused).
type StateChangedEvent struct{ State types.State }

func (e StateChangedEvent) isEvent() {} // Implements app.Event

// ProgressUpdatedEvent is published on player time updates.
type ProgressUpdatedEvent struct{ Passed, Total time.Duration }

func (e ProgressUpdatedEvent) isEvent() {} // Implements app.Event

// SongChangedEvent is published when the playing song changes.
// It crucially contains information about the previously playing song for reporting purposes.
type SongChangedEvent struct {
	NewSong            structs.Song
	PreviousSong       structs.Song
	PrevSongPlayedTime time.Duration
}

func (e SongChangedEvent) isEvent() {} // Implements app.Event

// PlaylistEndedEvent is published when the playlist reaches its end and there's no next song.
type PlaylistEndedEvent struct{}

func (e PlaylistEndedEvent) isEvent() {} // Implements app.Event

// SongLoadFailedEvent is published when fetching a song's details or URL fails.
type SongLoadFailedEvent struct{ Error error }

func (e SongLoadFailedEvent) isEvent() {} // Implements app.Event
