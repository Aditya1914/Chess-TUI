// Package ui handles the terminal user interface for the chess game.
package ui

import (
	"fmt"
	"strings"

	"chess/profile"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// SetupPhase represents the current phase of setup.
type SetupPhase int

const (
	PhaseWhitePlayer SetupPhase = iota
	PhaseBlackPlayer
	PhaseTimeControl
	PhaseConfirm
)

// SetupModel represents the pre-game setup UI.
type SetupModel struct {
	Phase          SetupPhase
	ProfileStore   *profile.ProfileStore
	ExistingNames  []string
	
	// White player
	WhiteInput     string
	WhiteSelected  int // -1 for new player, 0+ for existing
	WhiteName      string
	
	// Black player
	BlackInput     string
	BlackSelected  int
	BlackName      string
	
	// Time control
	TimeControlIdx int
	
	// Navigation
	CursorRow      int
	InputActive    bool
	
	// Error message
	ErrorMessage   string
}

// NewSetupModel creates a new setup model.
func NewSetupModel(store *profile.ProfileStore) SetupModel {
	names := store.GetProfileNames()
	return SetupModel{
		Phase:         PhaseWhitePlayer,
		ProfileStore:  store,
		ExistingNames: names,
		WhiteSelected: -1,
		BlackSelected: -1,
		TimeControlIdx: 4, // Default to 10 min
		CursorRow:     0,
		InputActive:   false,
	}
}

// Init initializes the setup model.
func (m SetupModel) Init() tea.Cmd {
	return nil
}

// Update handles messages for the setup model.
func (m SetupModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		if m.InputActive {
			return m.handleInputKey(msg)
		}
		return m.handleNavKey(msg)
	}
	return m, nil
}

// handleInputKey handles key presses when typing a name.
func (m SetupModel) handleInputKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "enter":
		// Confirm the input
		name := strings.TrimSpace(m.WhiteInput)
		if m.Phase == PhaseBlackPlayer {
			name = strings.TrimSpace(m.BlackInput)
		}
		
		if len(name) < 1 {
			m.ErrorMessage = "Name must be at least 1 character"
			return m, nil
		}
		if len(name) > 20 {
			m.ErrorMessage = "Name must be at most 20 characters"
			return m, nil
		}
		
		if m.Phase == PhaseWhitePlayer {
			m.WhiteName = name
			m.Phase = PhaseBlackPlayer
			m.InputActive = false
			m.WhiteSelected = -1
			m.WhiteInput = ""
			m.ErrorMessage = ""
		} else {
			m.BlackName = name
			m.Phase = PhaseTimeControl
			m.InputActive = false
			m.BlackSelected = -1
			m.BlackInput = ""
			m.ErrorMessage = ""
		}
		return m, nil
		
	case "escape":
		m.InputActive = false
		m.WhiteInput = ""
		m.BlackInput = ""
		m.ErrorMessage = ""
		return m, nil
		
	case "backspace":
		if m.Phase == PhaseWhitePlayer {
			if len(m.WhiteInput) > 0 {
				m.WhiteInput = m.WhiteInput[:len(m.WhiteInput)-1]
			}
		} else {
			if len(m.BlackInput) > 0 {
				m.BlackInput = m.BlackInput[:len(m.BlackInput)-1]
			}
		}
		return m, nil
		
	default:
		// Add character to input
		if len(msg.String()) == 1 {
			char := msg.String()[0]
			if char >= 32 && char <= 126 { // Printable ASCII
				if m.Phase == PhaseWhitePlayer {
					if len(m.WhiteInput) < 20 {
						m.WhiteInput += msg.String()
					}
				} else {
					if len(m.BlackInput) < 20 {
						m.BlackInput += msg.String()
					}
				}
			}
		}
		return m, nil
	}
}

// handleNavKey handles key presses during navigation.
func (m SetupModel) handleNavKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "ctrl+c":
		return m, tea.Quit
		
	case "up", "k":
		if m.CursorRow > 0 {
			m.CursorRow--
		}
		
	case "down", "j":
		maxRow := len(m.ExistingNames) + 1 // existing names + "New Player"
		if m.Phase == PhaseTimeControl {
			maxRow = len(profile.TimeControls)
		} else if m.Phase == PhaseConfirm {
			maxRow = 1
		}
		if m.CursorRow < maxRow-1 {
			m.CursorRow++
		}
		
	case "enter", " ":
		return m.handleSelect()
		
	case "n":
		// Shortcut for new player
		m.InputActive = true
		m.CursorRow = len(m.ExistingNames) // Position on "New Player"
		return m, nil
	}
	
	return m, nil
}

// handleSelect handles selection in setup.
func (m SetupModel) handleSelect() (tea.Model, tea.Cmd) {
	switch m.Phase {
	case PhaseWhitePlayer:
		if m.CursorRow == len(m.ExistingNames) {
			// New player - activate input
			m.InputActive = true
		} else if m.CursorRow < len(m.ExistingNames) {
			// Selected existing player
			m.WhiteName = m.ExistingNames[m.CursorRow]
			m.Phase = PhaseBlackPlayer
			m.CursorRow = 0
		}
		
	case PhaseBlackPlayer:
		if m.CursorRow == len(m.ExistingNames) {
			// New player - activate input
			m.InputActive = true
		} else if m.CursorRow < len(m.ExistingNames) {
			// Selected existing player
			m.BlackName = m.ExistingNames[m.CursorRow]
			m.Phase = PhaseTimeControl
			m.CursorRow = m.TimeControlIdx
		}
		
	case PhaseTimeControl:
		if m.CursorRow < len(profile.TimeControls) {
			m.TimeControlIdx = m.CursorRow
			m.Phase = PhaseConfirm
			m.CursorRow = 0
		}
		
	case PhaseConfirm:
		// Start the game - this will be handled by the main model
		return m, tea.Quit
	}
	
	return m, nil
}

// View renders the setup UI.
func (m SetupModel) View() string {
	var b strings.Builder
	
	// Title
	titleStyle := lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("#E8E8E8")).
		Background(lipgloss.Color("#4A7C4E")).
		Padding(1, 4).
		MarginBottom(1)
	
	b.WriteString(titleStyle.Render("♔ Chess Game Setup ♚"))
	b.WriteString("\n\n")
	
	// Progress indicator
	progressStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#888888"))
	activeStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#4A7C4E")).Bold(true)
	completedStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#7FA650"))
	
	phases := []string{"White Player", "Black Player", "Time Control", "Start Game"}
	for i, phase := range phases {
		if i < int(m.Phase) {
			b.WriteString(completedStyle.Render("✓ " + phase))
		} else if i == int(m.Phase) {
			b.WriteString(activeStyle.Render("► " + phase))
		} else {
			b.WriteString(progressStyle.Render("  " + phase))
		}
		if i < len(phases)-1 {
			b.WriteString(progressStyle.Render(" → "))
		}
	}
	b.WriteString("\n\n")
	
	// Content based on phase
	switch m.Phase {
	case PhaseWhitePlayer:
		b.WriteString(m.renderPlayerSelection("White", &m.WhiteInput, m.WhiteName))
	case PhaseBlackPlayer:
		b.WriteString(m.renderPlayerSelection("Black", &m.BlackInput, m.BlackName))
	case PhaseTimeControl:
		b.WriteString(m.renderTimeControlSelection())
	case PhaseConfirm:
		b.WriteString(m.renderConfirmation())
	}
	
	// Error message
	if m.ErrorMessage != "" {
		errorStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#FF6B6B"))
		b.WriteString("\n")
		b.WriteString(errorStyle.Render("Error: " + m.ErrorMessage))
	}
	
	// Controls
	b.WriteString("\n\n")
	controlsStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#888888"))
	b.WriteString(controlsStyle.Render("↑↓/jk: Navigate | Enter/Space: Select | n: New Player | Ctrl+C: Quit"))
	
	return lipgloss.NewStyle().
		Background(lipgloss.Color("#1E1E1E")).
		Padding(2, 4).
		Render(b.String())
}

// renderPlayerSelection renders the player selection screen.
func (m SetupModel) renderPlayerSelection(color string, input *string, selectedName string) string {
	var b strings.Builder
	
	headerStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("#E8E8E8")).
		Bold(true).
		MarginBottom(1)
	
	b.WriteString(headerStyle.Render(fmt.Sprintf("Select %s Player:", color)))
	b.WriteString("\n\n")
	
	itemStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#CCCCCC"))
	selectedItemStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("#FFFFFF")).
		Background(lipgloss.Color("#4A7C4E")).
		Bold(true)
	
	// Show existing profiles
	for i, name := range m.ExistingNames {
		p := m.ProfileStore.GetProfile(name)
		line := fmt.Sprintf("  %s (W:%d L:%d D:%d)", name, p.Stats.Wins, p.Stats.Losses, p.Stats.Draws)
		if m.CursorRow == i && !m.InputActive {
			b.WriteString(selectedItemStyle.Render("→" + line))
		} else {
			b.WriteString(itemStyle.Render(" " + line))
		}
		b.WriteString("\n")
	}
	
	// New player option
	newPlayerLine := "  New Player"
	if m.CursorRow == len(m.ExistingNames) && !m.InputActive {
		b.WriteString(selectedItemStyle.Render("→" + newPlayerLine))
	} else {
		b.WriteString(itemStyle.Render(" " + newPlayerLine))
	}
	b.WriteString("\n")
	
	// Input field if active
	if m.InputActive {
		b.WriteString("\n")
		inputLabelStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#E8E8E8"))
		b.WriteString(inputLabelStyle.Render(fmt.Sprintf("Enter %s player name: ", color)))
		b.WriteString(*input + "█")
		b.WriteString("\n")
	}
	
	return b.String()
}

// renderTimeControlSelection renders the time control selection screen.
func (m SetupModel) renderTimeControlSelection() string {
	var b strings.Builder
	
	headerStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("#E8E8E8")).
		Bold(true).
		MarginBottom(1)
	
	b.WriteString(headerStyle.Render("Select Time Control:"))
	b.WriteString("\n\n")
	
	itemStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#CCCCCC"))
	selectedItemStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("#FFFFFF")).
		Background(lipgloss.Color("#4A7C4E")).
		Bold(true)
	
	for i, tc := range profile.TimeControls {
		line := fmt.Sprintf("  %s", tc.Name)
		if m.CursorRow == i {
			b.WriteString(selectedItemStyle.Render("→" + line))
		} else {
			b.WriteString(itemStyle.Render(" " + line))
		}
		b.WriteString("\n")
	}
	
	return b.String()
}

// renderConfirmation renders the confirmation screen.
func (m SetupModel) renderConfirmation() string {
	var b strings.Builder
	
	headerStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("#E8E8E8")).
		Bold(true)
	
	whiteStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("#FFFFFF")).
		Background(lipgloss.Color("#333333")).
		Padding(0, 2)
	
	blackStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("#000000")).
		Background(lipgloss.Color("#CCCCCC")).
		Padding(0, 2)
	
	tcStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("#7FA650"))
	
	b.WriteString(headerStyle.Render("Game Configuration:"))
	b.WriteString("\n\n")
	
	// Players
	b.WriteString(fmt.Sprintf("White: %s", whiteStyle.Render(m.WhiteName)))
	b.WriteString("\n")
	b.WriteString(fmt.Sprintf("Black: %s", blackStyle.Render(m.BlackName)))
	b.WriteString("\n\n")
	
	// Time control
	tc := profile.TimeControls[m.TimeControlIdx]
	timeStr := tc.Name
	if tc.Duration > 0 {
		timeStr = fmt.Sprintf("%d minutes each", tc.Duration)
	}
	b.WriteString(fmt.Sprintf("Time Control: %s", tcStyle.Render(timeStr)))
	b.WriteString("\n\n")
	
	// Start button
	startStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("#FFFFFF")).
		Background(lipgloss.Color("#4A7C4E")).
		Padding(1, 4).
		Bold(true)
	
	b.WriteString(startStyle.Render("Press Enter to Start Game"))
	
	return b.String()
}

// GetGameConfig returns the configured game settings.
func (m SetupModel) GetGameConfig() (whiteName, blackName string, timeControl profile.TimeControl) {
	return m.WhiteName, m.BlackName, profile.TimeControls[m.TimeControlIdx]
}

// IsComplete returns true if setup is complete.
func (m SetupModel) IsComplete() bool {
	return m.Phase == PhaseConfirm
}
