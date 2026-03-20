// Package main is the entry point for the terminal chess game.
package main

import (
	"fmt"
	"os"

	"chess/profile"
	"chess/ui"

	tea "github.com/charmbracelet/bubbletea"
)

func main() {
	// Initialize profile store
	store, err := profile.NewProfileStore()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error initializing profile store: %v\n", err)
		os.Exit(1)
	}

	// Run setup to get player names and time control
	setupModel := ui.NewSetupModel(store)
	setupProgram := tea.NewProgram(
		setupModel,
		tea.WithAltScreen(),
	)

	finalSetup, err := setupProgram.Run()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error running setup: %v\n", err)
		os.Exit(1)
	}

	// Get configuration from setup
	setup, ok := finalSetup.(ui.SetupModel)
	if !ok || !setup.IsComplete() {
		fmt.Println("\nSetup cancelled.")
		return
	}

	whiteName, blackName, timeControl := setup.GetGameConfig()

	// Create the game model with configuration
	model := ui.NewModelWithConfig(store, whiteName, blackName, timeControl)

	// Create the Bubble Tea program
	p := tea.NewProgram(
		model,
		tea.WithAltScreen(),       // Use alternate screen buffer
		tea.WithMouseCellMotion(), // Enable mouse support for better terminal integration
	)

	// Run the program
	finalModel, err := p.Run()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error running chess game: %v\n", err)
		os.Exit(1)
	}

	// Print final state if needed
	if m, ok := finalModel.(ui.Model); ok {
		if m.Game.IsCheckmate() {
			fmt.Println("\nGame over - Checkmate!")
		} else if m.Game.IsStalemate() {
			fmt.Println("\nGame over - Stalemate!")
		}
	}
}
