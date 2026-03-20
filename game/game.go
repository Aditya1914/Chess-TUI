// Package game manages the chess game state and history.
package game

import (
	"chess/board"
	"chess/rules"
)

// MoveRecord stores information about a move for history.
type MoveRecord struct {
	Move        rules.Move
	BoardBefore *board.Board
	StateBefore *rules.GameState
	MoveNumber  int
	Notation    string
}

// Game represents a complete chess game.
type Game struct {
	Board         *board.Board
	State         *rules.GameState
	Validator     *rules.MoveValidator
	History       []MoveRecord
	HistoryIndex  int
}

// NewGame creates a new chess game.
func NewGame() *Game {
	b := board.NewBoard()
	return &Game{
		Board:        b,
		State:        rules.NewGameState(),
		Validator:    rules.NewMoveValidator(b),
		History:      make([]MoveRecord, 0),
		HistoryIndex: -1,
	}
}

// GetValidMoves returns all valid moves for a piece at the given position.
func (g *Game) GetValidMoves(pos board.Position) []rules.Move {
	return g.Validator.GetValidMoves(pos, g.State)
}

// MakeMove attempts to make a move from one position to another.
func (g *Game) MakeMove(from, to board.Position) bool {
	moves := g.GetValidMoves(from)
	
	for _, move := range moves {
		if move.To.Row == to.Row && move.To.Col == to.Col {
			return g.ExecuteMove(move)
		}
	}
	return false
}

// ExecuteMove executes a validated move.
func (g *Game) ExecuteMove(move rules.Move) bool {
	// Save current state to history
	boardCopy := g.Board.Copy()
	stateCopy := g.State.Copy()
	notation := g.generateNotation(move)
	
	// Clear any redo history
	if g.HistoryIndex < len(g.History)-1 {
		g.History = g.History[:g.HistoryIndex+1]
	}
	
	g.History = append(g.History, MoveRecord{
		Move:        move,
		BoardBefore: boardCopy,
		StateBefore: stateCopy,
		MoveNumber:  g.State.FullMoveNumber,
		Notation:    notation,
	})
	g.HistoryIndex = len(g.History) - 1
	
	// Execute the move
	piece := g.Board.GetPiece(move.From.Row, move.From.Col)
	
	// Handle castling
	if move.IsCastling {
		row := move.To.Row
		if move.To.Col == 6 { // Kingside
			rook := g.Board.GetPiece(row, 7)
			g.Board.SetPiece(row, 7, board.Piece{Type: board.NoPiece, Color: board.NoColor})
			g.Board.SetPiece(row, 5, rook)
		} else { // Queenside
			rook := g.Board.GetPiece(row, 0)
			g.Board.SetPiece(row, 0, board.Piece{Type: board.NoPiece, Color: board.NoColor})
			g.Board.SetPiece(row, 3, rook)
		}
	}
	
	// Handle en passant capture
	if move.IsEnPassant {
		captureRow := move.From.Row
		g.Board.SetPiece(captureRow, move.To.Col, board.Piece{Type: board.NoPiece, Color: board.NoColor})
	}
	
	// Move the piece
	g.Board.SetPiece(move.From.Row, move.From.Col, board.Piece{Type: board.NoPiece, Color: board.NoColor})
	
	// Handle pawn promotion
	if move.Promotion != board.NoPiece {
		g.Board.SetPiece(move.To.Row, move.To.Col, board.Piece{Type: move.Promotion, Color: piece.Color})
	} else {
		g.Board.SetPiece(move.To.Row, move.To.Col, piece)
	}
	
	// Update game state
	g.updateGameState(move, piece)
	
	return true
}

// updateGameState updates the game state after a move.
func (g *Game) updateGameState(move rules.Move, piece board.Piece) {
	// Update castling rights
	if piece.Type == board.King {
		if piece.Color == board.White {
			g.State.WhiteCanCastleKingside = false
			g.State.WhiteCanCastleQueenside = false
		} else {
			g.State.BlackCanCastleKingside = false
			g.State.BlackCanCastleQueenside = false
		}
	}
	
	if piece.Type == board.Rook {
		if move.From.Row == 7 && move.From.Col == 0 {
			g.State.WhiteCanCastleQueenside = false
		} else if move.From.Row == 7 && move.From.Col == 7 {
			g.State.WhiteCanCastleKingside = false
		} else if move.From.Row == 0 && move.From.Col == 0 {
			g.State.BlackCanCastleQueenside = false
		} else if move.From.Row == 0 && move.From.Col == 7 {
			g.State.BlackCanCastleKingside = false
		}
	}
	
	// Update en passant target
	if piece.Type == board.Pawn && abs(move.To.Row-move.From.Row) == 2 {
		enPassantRow := (move.From.Row + move.To.Row) / 2
		g.State.EnPassantTarget = board.Position{enPassantRow, move.From.Col}
	} else {
		g.State.EnPassantTarget = board.Position{-1, -1}
	}
	
	// Update half move clock
	if piece.Type == board.Pawn || move.CapturedPiece.Type != board.NoPiece {
		g.State.HalfMoveClock = 0
	} else {
		g.State.HalfMoveClock++
	}
	
	// Update full move number
	if g.State.CurrentTurn == board.Black {
		g.State.FullMoveNumber++
	}
	
	// Switch turn
	g.State.CurrentTurn = g.State.CurrentTurn.Opponent()
}

// generateNotation generates algebraic notation for a move.
func (g *Game) generateNotation(move rules.Move) string {
	piece := g.Board.GetPiece(move.From.Row, move.From.Col)
	
	// Castling
	if move.IsCastling {
		if move.To.Col == 6 {
			return "O-O"
		}
		return "O-O-O"
	}
	
	notation := ""
	
	// Piece letter (not for pawns)
	pieceLetters := map[board.PieceType]string{
		board.King:   "K",
		board.Queen:  "Q",
		board.Rook:   "R",
		board.Bishop: "B",
		board.Knight: "N",
	}
	
	if piece.Type != board.Pawn {
		notation += pieceLetters[piece.Type]
	}
	
	// Capture indicator
	if move.CapturedPiece.Type != board.NoPiece || move.IsEnPassant {
		if piece.Type == board.Pawn {
			notation += string(rune('a'+move.From.Col))
		}
		notation += "x"
	}
	
	// Destination square
	notation += move.To.ToAlgebraic()
	
	// Promotion
	if move.Promotion != board.NoPiece {
		notation += "=" + pieceLetters[move.Promotion]
	}
	
	// Check/checkmate indicator
	testBoard := g.Board.Copy()
	testValidator := rules.NewMoveValidator(testBoard)
	if testValidator.IsInCheck(g.State.CurrentTurn.Opponent()) {
		// This would need more complex logic to determine checkmate
		notation += "+"
	}
	
	return notation
}

// CanUndo returns true if there are moves to undo.
func (g *Game) CanUndo() bool {
	return g.HistoryIndex >= 0
}

// CanRedo returns true if there are moves to redo.
func (g *Game) CanRedo() bool {
	return g.HistoryIndex < len(g.History)-1
}

// Undo undoes the last move.
func (g *Game) Undo() bool {
	if !g.CanUndo() {
		return false
	}
	
	// Get the record for the move we're undoing
	record := g.History[g.HistoryIndex]
	
	// Restore the board and state from before the move
	g.Board = record.BoardBefore.Copy()
	g.State = record.StateBefore.Copy()
	g.Validator = rules.NewMoveValidator(g.Board)
	
	// Move history index back
	g.HistoryIndex--
	
	return true
}

// GetHistoryPosition returns the board at a specific point in history.
func (g *Game) GetHistoryPosition(index int) *board.Board {
	if index < 0 || index >= len(g.History) {
		return nil
	}
	return g.History[index].BoardBefore
}

// GetCurrentHistoryIndex returns the current position in history.
func (g *Game) GetCurrentHistoryIndex() int {
	return g.HistoryIndex
}

// GetHistoryLength returns the total number of moves in history.
func (g *Game) GetHistoryLength() int {
	return len(g.History)
}

// IsInCheck returns true if the current player's king is in check.
func (g *Game) IsInCheck() bool {
	return g.Validator.IsInCheck(g.State.CurrentTurn)
}

// IsCheckmate returns true if the current player is in checkmate.
func (g *Game) IsCheckmate() bool {
	return g.Validator.IsCheckmate(g.State.CurrentTurn, g.State)
}

// IsStalemate returns true if the current player is in stalemate.
func (g *Game) IsStalemate() bool {
	return g.Validator.IsStalemate(g.State.CurrentTurn, g.State)
}

// IsGameOver returns true if the game is over.
func (g *Game) IsGameOver() bool {
	return g.IsCheckmate() || g.IsStalemate()
}

// GetMoveHistory returns the list of move notations.
func (g *Game) GetMoveHistory() []string {
	var notations []string
	for _, record := range g.History {
		notations = append(notations, record.Notation)
	}
	return notations
}

func abs(x int) int {
	if x < 0 {
		return -x
	}
	return x
}
