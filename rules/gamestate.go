package rules

import (
	"chess/board"
)

// GameState tracks the current state of a chess game.
type GameState struct {
	CurrentTurn             board.Color
	WhiteCanCastleKingside  bool
	WhiteCanCastleQueenside bool
	BlackCanCastleKingside  bool
	BlackCanCastleQueenside bool
	EnPassantTarget         board.Position
	HalfMoveClock           int
	FullMoveNumber          int
}

// NewGameState creates a new game state with default values.
func NewGameState() *GameState {
	return &GameState{
		CurrentTurn:             board.White,
		WhiteCanCastleKingside:  true,
		WhiteCanCastleQueenside: true,
		BlackCanCastleKingside:  true,
		BlackCanCastleQueenside: true,
		EnPassantTarget:         board.Position{-1, -1},
		HalfMoveClock:           0,
		FullMoveNumber:          1,
	}
}

// Copy creates a deep copy of the game state.
func (gs *GameState) Copy() *GameState {
	return &GameState{
		CurrentTurn:             gs.CurrentTurn,
		WhiteCanCastleKingside:  gs.WhiteCanCastleKingside,
		WhiteCanCastleQueenside: gs.WhiteCanCastleQueenside,
		BlackCanCastleKingside:  gs.BlackCanCastleKingside,
		BlackCanCastleQueenside: gs.BlackCanCastleQueenside,
		EnPassantTarget:         gs.EnPassantTarget,
		HalfMoveClock:           gs.HalfMoveClock,
		FullMoveNumber:          gs.FullMoveNumber,
	}
}
