package cmd

import (
	"github.com/jorranMLGN/go-command/data"
	"github.com/jorranMLGN/go-command/ui"
)

// Run starts the application
func Run() error {
	// Initialize data store
	store, err := data.NewStore()
	if err != nil {
		return err
	}

	// Start UI
	app := ui.NewApp(store)
	return app.Run()
}