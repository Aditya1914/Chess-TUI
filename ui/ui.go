// Package ui handles the terminal user interface for the chess game.
package ui

import (
	"fmt"
	"strings"
	"time"

	"chess/board"
	"chess/game"
	"chess/profile"
	"chess/rules"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// ConfirmationType represents the type of confirmation dialog.
type ConfirmationType int

const (
	NoConfirmation ConfirmationType = iota
	ConfirmQuit
	ConfirmNewGame
)

// TickMsg is sent every second to update the clock.
type TickMsg struct{}

// GamePhase represents the current phase of the application.
type GamePhase int

const (
	PhaseSetup GamePhase = iota
	PhasePlaying
	PhaseGameOver
)

// Model represents the UI model for the chess game.
type Model struct {
	Game             *game.Game
	CursorRow        int
	CursorCol        int
	SelectedPos      *board.Position
	ValidMoves       []rules.Move
	LastMove         *rules.Move
	ViewHistory      bool
	HistoryOffset    int
	PromotionPending bool
	PromotionFrom    board.Position
	PromotionTo      board.Position
	Message          string
	TerminalWidth    int
	TerminalHeight   int
	Confirmation     ConfirmationType
	// Chess clocks
	WhiteTime        time.Duration
	BlackTime        time.Duration
	GameStartTime    time.Time
	LastTickTime     time.Time
	ClockRunning     bool
	// Board orientation
	FlipBoard        bool // If true, view from Black's perspective
	// Player profiles
	ProfileStore     *profile.ProfileStore
	WhitePlayer      *profile.Profile
	BlackPlayer      *profile.Profile
	// Time control
	TimeControl      profile.TimeControl
	// Game phase
	Phase            GamePhase
	// Game result (for statistics)
	GameResult       string // "white", "black", "draw", or ""
}

// NewModel creates a new UI model.
func NewModel() Model {
	return Model{
		Game:          game.NewGame(),
		CursorRow:     4,
		CursorCol:     4,
		SelectedPos:   nil,
		ValidMoves:    nil,
		ViewHistory:   false,
		HistoryOffset: 0,
		Message:       "",
		Confirmation:  NoConfirmation,
		WhiteTime:     10 * time.Minute, // 10 minutes default
		BlackTime:     10 * time.Minute,
		ClockRunning:  false,
		Phase:         PhaseSetup,
		GameResult:    "",
	}
}

// NewModelWithConfig creates a new UI model with player and time control configuration.
func NewModelWithConfig(store *profile.ProfileStore, whiteName, blackName string, tc profile.TimeControl) Model {
	var whiteTime, blackTime time.Duration
	if tc.Duration > 0 {
		whiteTime = time.Duration(tc.Duration) * time.Minute
		blackTime = time.Duration(tc.Duration) * time.Minute
	} else {
		// Unlimited time - set to a very large value
		whiteTime = 1000 * time.Hour
		blackTime = 1000 * time.Hour
	}

	return Model{
		Game:          game.NewGame(),
		CursorRow:     4,
		CursorCol:     4,
		SelectedPos:   nil,
		ValidMoves:    nil,
		ViewHistory:   false,
		HistoryOffset: 0,
		Message:       "",
		Confirmation:  NoConfirmation,
		WhiteTime:     whiteTime,
		BlackTime:     blackTime,
		ClockRunning:  false,
		ProfileStore:  store,
		WhitePlayer:   store.GetProfile(whiteName),
		BlackPlayer:   store.GetProfile(blackName),
		TimeControl:   tc,
		Phase:         PhasePlaying,
		GameResult:    "",
	}
}

// Init initializes the model.
func (m Model) Init() tea.Cmd {
	return nil
}

// tickCmd returns a command that sends a tick message every second.
func tickCmd() tea.Cmd {
	return tea.Tick(time.Second, func(time.Time) tea.Msg {
		return TickMsg{}
	})
}

// Update handles messages and updates the model.
func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case TickMsg:
		// Update clock
		if m.ClockRunning && !m.Game.IsCheckmate() && !m.Game.IsStalemate() {
			now := time.Now()
			elapsed := now.Sub(m.LastTickTime)
			m.LastTickTime = now

			if m.Game.State.CurrentTurn == board.White {
				m.WhiteTime -= elapsed
				if m.WhiteTime <= 0 {
					m.WhiteTime = 0
					m.ClockRunning = false
					m.GameResult = "black"
					m.Message = "Time out! Black wins!"
					m.Phase = PhaseGameOver
					m.recordGameResult("black")
				}
			} else {
				m.BlackTime -= elapsed
				if m.BlackTime <= 0 {
					m.BlackTime = 0
					m.ClockRunning = false
					m.GameResult = "white"
					m.Message = "Time out! White wins!"
					m.Phase = PhaseGameOver
					m.recordGameResult("white")
				}
			}
		}
		return m, tickCmd()

	case tea.KeyMsg:
		return m.handleKeyPress(msg)

	case tea.WindowSizeMsg:
		m.TerminalWidth = msg.Width
		m.TerminalHeight = msg.Height
		return m, nil
	}
	return m, nil
}

// handleKeyPress handles keyboard input.
func (m Model) handleKeyPress(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	// Handle confirmation dialogs first
	if m.Confirmation != NoConfirmation {
		return m.handleConfirmationKey(msg)
	}

	// Handle promotion selection
	if m.PromotionPending {
		return m.handlePromotionKey(msg)
	}

	// Handle history view mode
	if m.ViewHistory {
		return m.handleHistoryKey(msg)
	}

	switch msg.String() {
	case "ctrl+c":
		return m, tea.Quit
	case "q":
		m.Confirmation = ConfirmQuit
		return m, nil
	case "up", "k":
		if m.CursorRow > 0 {
			m.CursorRow--
		}
	case "down", "j":
		if m.CursorRow < 7 {
			m.CursorRow++
		}
	case "left", "h":
		if m.CursorCol > 0 {
			m.CursorCol--
		}
	case "right", "l":
		if m.CursorCol < 7 {
			m.CursorCol++
		}
	case "enter", " ":
		return m.handleSelect()
	case "u":
		// Undo move
		if m.Game.CanUndo() {
			m.Game.Undo()
			m.SelectedPos = nil
			m.ValidMoves = nil
			m.LastMove = nil
			m.Message = "Move undone"
		}
	case "r":
		// Ask for confirmation before new game
		m.Confirmation = ConfirmNewGame
		return m, nil
	case "v":
		// View history
		if m.Game.GetHistoryLength() > 0 {
			m.ViewHistory = true
			m.HistoryOffset = 0
		}
	case "f":
		// Flip board orientation
		m.FlipBoard = !m.FlipBoard
	}

	return m, nil
}

// handleConfirmationKey handles key presses during confirmation dialogs.
func (m Model) handleConfirmationKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "y", "Y":
		if m.Confirmation == ConfirmQuit {
			return m, tea.Quit
		} else if m.Confirmation == ConfirmNewGame {
			m.Game = game.NewGame()
			m.SelectedPos = nil
			m.ValidMoves = nil
			m.LastMove = nil
			m.Message = ""
			// Reset time based on current time control
			if m.TimeControl.Duration > 0 {
				m.WhiteTime = time.Duration(m.TimeControl.Duration) * time.Minute
				m.BlackTime = time.Duration(m.TimeControl.Duration) * time.Minute
			} else {
				m.WhiteTime = 1000 * time.Hour
				m.BlackTime = 1000 * time.Hour
			}
			m.ClockRunning = false
			m.GameResult = ""
			m.Phase = PhasePlaying
		}
		m.Confirmation = NoConfirmation
	case "n", "N", "escape":
		m.Confirmation = NoConfirmation
		m.Message = ""
	}
	return m, nil
}

// handlePromotionKey handles key presses during pawn promotion.
func (m Model) handlePromotionKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	m.Message = "Choose promotion: Q=Queen, R=Rook, B=Bishop, N=Knight"
	
	var promotionPiece board.PieceType
	switch msg.String() {
	case "q":
		promotionPiece = board.Queen
	case "r":
		promotionPiece = board.Rook
	case "b":
		promotionPiece = board.Bishop
	case "n":
		promotionPiece = board.Knight
	default:
		return m, nil
	}

	// Find the move and apply with promotion
	for _, move := range m.ValidMoves {
		if move.To.Row == m.PromotionTo.Row && move.To.Col == m.PromotionTo.Col {
			promotionMove := move
			promotionMove.Promotion = promotionPiece
			m.Game.ExecuteMove(promotionMove)
			m.LastMove = &promotionMove
			break
		}
	}

	m.PromotionPending = false
	m.SelectedPos = nil
	m.ValidMoves = nil
	m.Message = ""
	
	// Check for game end
	if m.Game.IsCheckmate() {
		winner := "Black"
		if m.Game.State.CurrentTurn == board.Black {
			winner = "White"
		}
		m.Message = fmt.Sprintf("Checkmate! %s wins!", winner)
		m.ClockRunning = false
		m.Phase = PhaseGameOver
		m.GameResult = strings.ToLower(winner)
		m.recordGameResult(m.GameResult)
	} else if m.Game.IsStalemate() {
		m.Message = "Stalemate! Game is a draw."
		m.ClockRunning = false
		m.Phase = PhaseGameOver
		m.GameResult = "draw"
		m.recordGameResult("draw")
	} else if m.Game.IsInCheck() {
		m.Message = "Check!"
	}

	return m, nil
}

// handleHistoryKey handles key presses in history view mode.
func (m Model) handleHistoryKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "q", "escape", "v":
		m.ViewHistory = false
	case "up", "k":
		if m.HistoryOffset > 0 {
			m.HistoryOffset--
		}
	case "down", "j":
		if m.HistoryOffset < m.Game.GetHistoryLength()-1 {
			m.HistoryOffset++
		}
	case "left", "h":
		if m.HistoryOffset > 0 {
			m.HistoryOffset--
		}
	case "right", "l":
		if m.HistoryOffset < m.Game.GetHistoryLength()-1 {
			m.HistoryOffset++
		}
	}
	return m, nil
}

// handleSelect handles piece selection and move execution.
func (m Model) handleSelect() (tea.Model, tea.Cmd) {
	currentPos := board.Position{Row: m.CursorRow, Col: m.CursorCol}

	if m.SelectedPos == nil {
		// Select a piece
		piece := m.Game.Board.GetPiece(m.CursorRow, m.CursorCol)
		if piece.Type != board.NoPiece && piece.Color == m.Game.State.CurrentTurn {
			m.SelectedPos = &currentPos
			m.ValidMoves = m.Game.GetValidMoves(currentPos)
			if len(m.ValidMoves) == 0 {
				m.Message = "No valid moves for this piece"
			} else {
				m.Message = ""
			}
		}
	} else {
		// Try to make a move
		moveMade := false
		for _, move := range m.ValidMoves {
			if move.To.Row == currentPos.Row && move.To.Col == currentPos.Col {
				// Check if this is a pawn promotion
				piece := m.Game.Board.GetPiece(m.SelectedPos.Row, m.SelectedPos.Col)
				if piece.Type == board.Pawn {
					promotionRow := 0
					if piece.Color == board.Black {
						promotionRow = 7
					}
					if currentPos.Row == promotionRow {
						m.PromotionPending = true
						m.PromotionFrom = *m.SelectedPos
						m.PromotionTo = currentPos
						m.Message = "Choose promotion: Q=Queen, R=Rook, B=Bishop, N=Knight"
						return m, nil
					}
				}

				m.Game.ExecuteMove(move)
				m.LastMove = &move
				moveMade = true
				
				// Start clock on first move
				if !m.ClockRunning {
					m.ClockRunning = true
					m.LastTickTime = time.Now()
				}
				
				break
			}
		}

		if moveMade {
			m.SelectedPos = nil
			m.ValidMoves = nil
			m.Message = ""
			
			// Check for game end
			if m.Game.IsCheckmate() {
				winner := "Black"
				if m.Game.State.CurrentTurn == board.Black {
					winner = "White"
				}
				m.Message = fmt.Sprintf("Checkmate! %s wins!", winner)
				m.ClockRunning = false
				m.Phase = PhaseGameOver
				m.GameResult = strings.ToLower(winner)
				m.recordGameResult(m.GameResult)
			} else if m.Game.IsStalemate() {
				m.Message = "Stalemate! Game is a draw."
				m.ClockRunning = false
				m.Phase = PhaseGameOver
				m.GameResult = "draw"
				m.recordGameResult("draw")
			} else if m.Game.IsInCheck() {
				m.Message = "Check!"
			}
		} else {
			// Select a different piece or deselect
			piece := m.Game.Board.GetPiece(m.CursorRow, m.CursorCol)
			if piece.Type != board.NoPiece && piece.Color == m.Game.State.CurrentTurn {
				m.SelectedPos = &currentPos
				m.ValidMoves = m.Game.GetValidMoves(currentPos)
				m.Message = ""
			} else {
				m.SelectedPos = nil
				m.ValidMoves = nil
				m.Message = ""
			}
		}
	}

	var cmd tea.Cmd
	if m.ClockRunning {
		cmd = tickCmd()
	}
	return m, cmd
}

// View renders the UI.
func (m Model) View() string {
	if m.ViewHistory {
		return m.renderHistoryView()
	}
	return m.renderGameView()
}

// formatTime formats a duration as MM:SS
func formatTime(d time.Duration) string {
	minutes := int(d.Minutes())
	seconds := int(d.Seconds()) % 60
	return fmt.Sprintf("%02d:%02d", minutes, seconds)
}

// renderGameView renders the main game view.
func (m Model) renderGameView() string {
	// Color scheme - realistic chess board with excellent contrast
	lightSquareBg := "#F0D9B5" // Light cream
	darkSquareBg := "#B58863"  // Wood brown

	// Board border style
	boardStyle := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("#8B7355")).
		Padding(0, 1)

	// Square styles - base background only
	lightSquareStyle := lipgloss.NewStyle().
		Background(lipgloss.Color(lightSquareBg))

	darkSquareStyle := lipgloss.NewStyle().
		Background(lipgloss.Color(darkSquareBg))

	// Highlight styles - clear, distinct colors
	selectedStyle := lipgloss.NewStyle().
		Background(lipgloss.Color("#7FA650")) // Green for selected

	cursorStyle := lipgloss.NewStyle().
		Background(lipgloss.Color("#CDD26A")) // Yellow-green for cursor

	validMoveStyle := lipgloss.NewStyle().
		Background(lipgloss.Color("#646F40")) // Darker green for valid moves

	lastMoveStyle := lipgloss.NewStyle().
		Background(lipgloss.Color("#F0E68C")) // Light yellow for last move

	checkStyle := lipgloss.NewStyle().
		Background(lipgloss.Color("#FF6B6B")) // Red for check

	// Piece colors - WHITE pieces are BLACK text (high contrast), BLACK pieces are WHITE text
	// This makes them clearly distinguishable on any square color
	whitePieceFg := lipgloss.Color("#000000") // Black text for white pieces
	blackPieceFg := lipgloss.Color("#FFFFFF") // White text for black pieces

	// Square dimensions - MUST be consistent for perfect square grid
	squareWidth := 9
	squareHeight := 3

	// Build the board with fixed structure
	var boardBuilder strings.Builder

	// Row label width (for "8 " etc) - must be consistent
	rowLabelWidth := 3

	// Top border with column labels - centered over each square
	boardBuilder.WriteString(strings.Repeat(" ", rowLabelWidth))
	for col := 0; col < 8; col++ {
		// Calculate file letter based on flip state
		var fileChar byte
		if m.FlipBoard {
			fileChar = 'h' - byte(col)
		} else {
			fileChar = 'a' + byte(col)
		}
		// Center the label in squareWidth characters
		label := string(rune(fileChar))
		leftPad := (squareWidth - 1) / 2
		rightPad := squareWidth - 1 - leftPad
		boardBuilder.WriteString(strings.Repeat(" ", leftPad))
		boardBuilder.WriteString(label)
		boardBuilder.WriteString(strings.Repeat(" ", rightPad))
	}
	boardBuilder.WriteString("\n")

	// Find king position for check highlighting
	var kingInCheck *board.Position
	if m.Game.IsInCheck() {
		for row := 0; row < 8; row++ {
			for col := 0; col < 8; col++ {
				p := m.Game.Board.GetPiece(row, col)
				if p.Type == board.King && p.Color == m.Game.State.CurrentTurn {
					pos := board.Position{row, col}
					kingInCheck = &pos
					break
				}
			}
		}
	}

	// Build each row with fixed height
	for rowIdx := 0; rowIdx < 8; rowIdx++ {
		// Calculate actual board row based on flip state
		var boardRow int
		if m.FlipBoard {
			boardRow = 7 - rowIdx // Flip: start from rank 1 going up
		} else {
			boardRow = rowIdx     // Normal: start from rank 8 going down
		}

		for lineIdx := 0; lineIdx < squareHeight; lineIdx++ {
			// Row number on the left (only on middle line)
			if lineIdx == squareHeight/2 {
				// Calculate rank number based on flip state
				var rankNum int
				if m.FlipBoard {
					rankNum = rowIdx + 1 // 1-8 from top to bottom when flipped
				} else {
					rankNum = 8 - rowIdx // 8-1 from top to bottom normally
				}
				boardBuilder.WriteString(fmt.Sprintf("%d  ", rankNum))
			} else {
				boardBuilder.WriteString(strings.Repeat(" ", rowLabelWidth))
			}

			// Build each column
			for colIdx := 0; colIdx < 8; colIdx++ {
				// Calculate actual board column based on flip state
				var boardCol int
				if m.FlipBoard {
					boardCol = 7 - colIdx // Flip: h to a (right to left)
				} else {
					boardCol = colIdx     // Normal: a to h (left to right)
				}

				isLight := (boardRow+boardCol)%2 == 0
				piece := m.Game.Board.GetPiece(boardRow, boardCol)

				// Determine piece foreground color based on piece color
				// White pieces get black text, Black pieces get white text
				var pieceFg lipgloss.Color
				if piece.Color == board.White {
					pieceFg = whitePieceFg
				} else if piece.Color == board.Black {
					pieceFg = blackPieceFg
				}

				// Determine square style
				var baseStyle lipgloss.Style
				if kingInCheck != nil && kingInCheck.Row == boardRow && kingInCheck.Col == boardCol {
					baseStyle = checkStyle
				} else if m.SelectedPos != nil && m.SelectedPos.Row == boardRow && m.SelectedPos.Col == boardCol {
					baseStyle = selectedStyle
				} else if m.CursorRow == boardRow && m.CursorCol == boardCol {
					baseStyle = cursorStyle
				} else if m.LastMove != nil &&
					((m.LastMove.From.Row == boardRow && m.LastMove.From.Col == boardCol) ||
						(m.LastMove.To.Row == boardRow && m.LastMove.To.Col == boardCol)) {
					baseStyle = lastMoveStyle
				} else if m.isValidMoveSquare(boardRow, boardCol) {
					baseStyle = validMoveStyle
				} else if isLight {
					baseStyle = lightSquareStyle
				} else {
					baseStyle = darkSquareStyle
				}

				// Render the square line
				pieceStr := " "
				if piece.Type != board.NoPiece {
					pieceStr = piece.String()
				}

				// Center the piece on the middle line
				if lineIdx == squareHeight/2 && piece.Type != board.NoPiece {
					leftPad := (squareWidth - 1) / 2
					pieceStyle := baseStyle.Copy().Foreground(pieceFg).Bold(true)
					line := strings.Repeat(" ", leftPad) + pieceStr + strings.Repeat(" ", squareWidth-leftPad-1)
					boardBuilder.WriteString(pieceStyle.Render(line))
				} else {
					boardBuilder.WriteString(baseStyle.Render(strings.Repeat(" ", squareWidth)))
				}
			}
			boardBuilder.WriteString("\n")
		}
	}

	boardStr := boardStyle.Render(boardBuilder.String())

	// Build info panel with fixed height
	turnStr := "White"
	turnStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#000000")).Background(lipgloss.Color("#F0D9B5")).Bold(true).Padding(0, 1)
	if m.Game.State.CurrentTurn == board.Black {
		turnStr = "Black"
		turnStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("#FFFFFF")).Background(lipgloss.Color("#333333")).Bold(true).Padding(0, 1)
	}

	infoStyle := lipgloss.NewStyle().
		Padding(1, 2).
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("#8B7355")).
		Background(lipgloss.Color("#2B2B2B"))

	titleStyle := lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("#E8E8E8"))

	// Clock styles
	clockStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("#FFFFFF")).
		Background(lipgloss.Color("#333333")).
		Padding(0, 1).
		Bold(true)

	activeClockStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("#FFFFFF")).
		Background(lipgloss.Color("#4A7C4E")).
		Padding(0, 1).
		Bold(true)

	whiteClockStyle := clockStyle
	blackClockStyle := clockStyle
	if m.ClockRunning && !m.Game.IsCheckmate() && !m.Game.IsStalemate() {
		if m.Game.State.CurrentTurn == board.White {
			whiteClockStyle = activeClockStyle
		} else {
			blackClockStyle = activeClockStyle
		}
	}

	whiteClockStr := formatTime(m.WhiteTime)
	blackClockStr := formatTime(m.BlackTime)

	// Time control display
	tcStr := "Unlimited"
	if m.TimeControl.Duration > 0 {
		tcStr = fmt.Sprintf("%d min", m.TimeControl.Duration)
	}

	// Game status - fixed set of lines
	statusLine := ""
	if m.Game.IsCheckmate() {
		winner := "Black"
		if m.Game.State.CurrentTurn == board.Black {
			winner = "White"
		}
		statusLine = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#FF6B6B")).Bold(true).Render(fmt.Sprintf("CHECKMATE! %s wins!", winner))
	} else if m.Game.IsStalemate() {
		statusLine = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#FFE66D")).Bold(true).Render("STALEMATE!")
	} else if m.Game.IsInCheck() {
		statusLine = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#FF6B6B")).Bold(true).Render("CHECK!")
	}

	// Player names
	whiteName := "White"
	blackName := "Black"
	if m.WhitePlayer != nil {
		whiteName = m.WhitePlayer.Name
	}
	if m.BlackPlayer != nil {
		blackName = m.BlackPlayer.Name
	}

	// Build info panel with consistent structure
	infoLines := []string{
		titleStyle.Render("Chess"),
		"",
		fmt.Sprintf("Time: %s", lipgloss.NewStyle().Foreground(lipgloss.Color("#7FA650")).Render(tcStr)),
		"",
		fmt.Sprintf("Turn: %s", turnStyle.Render(turnStr)),
		fmt.Sprintf("Move: %d", m.Game.State.FullMoveNumber),
		"",
		"Players:",
		fmt.Sprintf("  %s (W): %s", whiteName, whiteClockStyle.Render(whiteClockStr)),
		fmt.Sprintf("  %s (B): %s", blackName, blackClockStyle.Render(blackClockStr)),
		"",
		statusLine,
		"",
		"Controls:",
		"  ←→↑↓/hjkl Move",
		"  Enter/Space Select",
		"  u  Undo",
		"  f  Flip board",
		"  r  New game",
		"  q  Quit",
	}

	// Add confirmation if needed
	if m.Confirmation != NoConfirmation {
		confirmStyle := lipgloss.NewStyle().
			Foreground(lipgloss.Color("#FFE66D")).
			Bold(true)
		confirmMsg := "Quit? (y/n)"
		if m.Confirmation == ConfirmNewGame {
			confirmMsg = "New game? (y/n)"
		}
		infoLines = append(infoLines, "", confirmStyle.Render(confirmMsg))
	}

	infoStr := infoStyle.Render(strings.Join(infoLines, "\n"))

	// Build move history panel with fixed height (show last 8 moves max)
	historyStyle := lipgloss.NewStyle().
		Padding(1, 2).
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("#8B7355")).
		Background(lipgloss.Color("#2B2B2B"))

	moves := m.Game.GetMoveHistory()
	historyTitle := lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("#E8E8E8")).
		Render("History")

	moveStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#CCCCCC"))
	moveNumStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#888888"))

	// Build history lines - always show fixed number of lines
	historyLines := []string{historyTitle, ""}
	
	// Show last 8 full moves (16 half-moves)
	startIdx := 0
	if len(moves) > 16 {
		startIdx = len(moves) - 16
	}
	
	for i := startIdx; i < len(moves); i += 2 {
		moveNum := i/2 + 1
		line := moveNumStyle.Render(fmt.Sprintf("%2d. ", moveNum))
		line += moveStyle.Render(moves[i])
		if i+1 < len(moves) {
			line += " " + moveStyle.Render(moves[i+1])
		}
		historyLines = append(historyLines, line)
	}

	// Pad to ensure consistent height
	for len(historyLines) < 12 {
		historyLines = append(historyLines, "")
	}

	// Promotion dialog
	if m.PromotionPending {
		promoStyle := lipgloss.NewStyle().
			Foreground(lipgloss.Color("#FFE66D")).
			Bold(true)
		historyLines = append(historyLines, "",
			promoStyle.Render("Promote:"),
			"  Q Queen",
			"  R Rook",
			"  B Bishop",
			"  N Knight")
	}

	historyStr := historyStyle.Render(strings.Join(historyLines, "\n"))

	// Layout: Board on left, info and history stacked on right (fixed height)
	// This keeps the board in a fixed position
	rightPanel := lipgloss.JoinVertical(lipgloss.Top, infoStr, historyStr)
	mainContent := lipgloss.JoinHorizontal(lipgloss.Top, boardStr, "  ", rightPanel)

	return lipgloss.NewStyle().
		Background(lipgloss.Color("#1E1E1E")).
		Padding(1, 2).
		Render(mainContent)
}

// renderHistoryView renders the move history view.
func (m Model) renderHistoryView() string {
	titleStyle := lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("86")).
		Padding(1, 2)

	historyStyle := lipgloss.NewStyle().
		Padding(1, 2).
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("62"))

	title := titleStyle.Render("Move History (press v or escape to return)")

	moves := m.Game.GetMoveHistory()
	var lines []string

	for i := 0; i < len(moves); i += 2 {
		moveNum := i/2 + 1
		line := fmt.Sprintf("%2d. %-8s", moveNum, moves[i])
		if i+1 < len(moves) {
			line += fmt.Sprintf("  %-8s", moves[i+1])
		}

		// Highlight current position
		if i/2 == m.HistoryOffset {
			line = lipgloss.NewStyle().
				Background(lipgloss.Color("62")).
				Foreground(lipgloss.Color("15")).
				Render(line)
		}
		lines = append(lines, line)
	}

	if len(moves) == 0 {
		lines = append(lines, "No moves yet")
	}

	controls := "\n\nUse h/l or arrows to navigate, v or escape to return"
	content := strings.Join(lines, "\n") + controls

	return lipgloss.NewStyle().Padding(2, 4).Render(
		lipgloss.JoinVertical(lipgloss.Left, title, historyStyle.Render(content)),
	)
}

// isValidMoveSquare checks if the position is a valid move destination.
func (m Model) isValidMoveSquare(row, col int) bool {
	for _, move := range m.ValidMoves {
		if move.To.Row == row && move.To.Col == col {
			return true
		}
	}
	return false
}

// recordGameResult records the game result in player profiles.
func (m *Model) recordGameResult(winner string) {
	if m.ProfileStore == nil || m.WhitePlayer == nil || m.BlackPlayer == nil {
		return
	}
	m.ProfileStore.RecordGameResult(m.WhitePlayer.Name, m.BlackPlayer.Name, winner)
}

