package app

// MusicfoxApp is the headless application core.
// It will hold and coordinate all core services, acting as the central
// point for business logic orchestration, completely independent of the UI.
type MusicfoxApp struct {
	// Dependencies like player.Service, services.UserService, etc.,
	// will be added here.
}

// New creates a new instance of the application core.
// It will accept all necessary services as dependencies (Dependency Injection).
func New() *MusicfoxApp {
	return &MusicfoxApp{}
}
