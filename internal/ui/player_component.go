package ui

import (
	"fmt"

	"github.com/anhoder/foxful-cli/model"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/go-musicfox/go-musicfox/internal/core/player"
)

// PlayerComponent is the UI component responsible for rendering player information.
// It is completely decoupled from the player's core logic, depending only on the
// player.Service interface to query state for rendering.
type PlayerComponent struct {
	player player.Service // Depends on the interface, not the implementation.
	// Internal state for rendering will be added here, e.g.,
	// currentSongName string
	// progress        float64
}

// NewPlayerComponent creates a new PlayerComponent.
// It requires a player.Service to be injected as a dependency.
func NewPlayerComponent(playerSvc player.Service) *PlayerComponent {
	return &PlayerComponent{
		player: playerSvc,
	}
}

// Update handles messages for the PlayerComponent.
// It will listen for events (forwarded as tea.Msg) and update its internal state.
func (c *PlayerComponent) Update(msg tea.Msg) tea.Cmd {
	// Logic to handle events and update component's internal state for rendering
	// will be added here in a later commit.
	return nil
}

// View renders the player UI based on its current internal state.
// The logic from the old ui.Player's View methods will be moved here.
func (c *PlayerComponent) View(a *model.App, main *model.Main) (view string, lines int) {
	// Placeholder view for now.
	// In the future, this will use c.player.State(), c.player.CurrentSong(), etc.
	// to get data for rendering.
	song := c.player.CurrentSong()
	state := c.player.State()
	passed := c.player.PassedTime()

	info := "Nothing playing."
	if song.Id != 0 {
		info = "Now Playing: " + song.Name
	}

	stateStr := fmt.Sprintf("State: %d", state)

	return info + "\n" + stateStr + " " + passed.String() + "\n\n", 3
}
