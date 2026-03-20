// Package board defines the core chess board and piece types.
package board

// Piece represents a chess piece with its type and color.
type Piece struct {
	Type  PieceType
	Color Color
}

// PieceType represents the type of a chess piece.
type PieceType int

const (
	NoPiece PieceType = iota
	Pawn
	Knight
	Bishop
	Rook
	Queen
	King
)

// Color represents the color of a piece.
type Color int

const (
	NoColor Color = iota
	White
	Black
)

// String returns the Unicode symbol for the piece.
func (p Piece) String() string {
	if p.Type == NoPiece {
		return " "
	}

	symbols := map[PieceType]map[Color]string{
		King:   {White: "♔", Black: "♚"},
		Queen:  {White: "♕", Black: "♛"},
		Rook:   {White: "♖", Black: "♜"},
		Bishop: {White: "♗", Black: "♝"},
		Knight: {White: "♘", Black: "♞"},
		Pawn:   {White: "♙", Black: "♟"},
	}

	return symbols[p.Type][p.Color]
}

// Board represents an 8x8 chess board.
type Board struct {
	Squares [8][8]Piece
}

// NewBoard creates a new board with pieces in starting positions.
func NewBoard() *Board {
	b := &Board{}
	b.setupInitialPosition()
	return b
}

// setupInitialPosition places all pieces in their starting positions.
func (b *Board) setupInitialPosition() {
	// White pieces (rank 1, index 7 in our representation)
	b.Squares[7][0] = Piece{Rook, White}
	b.Squares[7][1] = Piece{Knight, White}
	b.Squares[7][2] = Piece{Bishop, White}
	b.Squares[7][3] = Piece{Queen, White}
	b.Squares[7][4] = Piece{King, White}
	b.Squares[7][5] = Piece{Bishop, White}
	b.Squares[7][6] = Piece{Knight, White}
	b.Squares[7][7] = Piece{Rook, White}

	// White pawns (rank 2, index 6)
	for col := 0; col < 8; col++ {
		b.Squares[6][col] = Piece{Pawn, White}
	}

	// Black pieces (rank 8, index 0)
	b.Squares[0][0] = Piece{Rook, Black}
	b.Squares[0][1] = Piece{Knight, Black}
	b.Squares[0][2] = Piece{Bishop, Black}
	b.Squares[0][3] = Piece{Queen, Black}
	b.Squares[0][4] = Piece{King, Black}
	b.Squares[0][5] = Piece{Bishop, Black}
	b.Squares[0][6] = Piece{Knight, Black}
	b.Squares[0][7] = Piece{Rook, Black}

	// Black pawns (rank 7, index 1)
	for col := 0; col < 8; col++ {
		b.Squares[1][col] = Piece{Pawn, Black}
	}
}

// GetPiece returns the piece at the given position.
func (b *Board) GetPiece(row, col int) Piece {
	if row < 0 || row > 7 || col < 0 || col > 7 {
		return Piece{NoPiece, NoColor}
	}
	return b.Squares[row][col]
}

// SetPiece places a piece at the given position.
func (b *Board) SetPiece(row, col int, p Piece) {
	if row >= 0 && row <= 7 && col >= 0 && col <= 7 {
		b.Squares[row][col] = p
	}
}

// Copy creates a deep copy of the board.
func (b *Board) Copy() *Board {
	newBoard := &Board{}
	for row := 0; row < 8; row++ {
		for col := 0; col < 8; col++ {
			newBoard.Squares[row][col] = b.Squares[row][col]
		}
	}
	return newBoard
}

// Position represents a square on the board.
type Position struct {
	Row, Col int
}

// IsValid checks if the position is within the board bounds.
func (p Position) IsValid() bool {
	return p.Row >= 0 && p.Row <= 7 && p.Col >= 0 && p.Col <= 7
}

// ToAlgebraic returns the algebraic notation for the position (e.g., "e4").
func (p Position) ToAlgebraic() string {
	if !p.IsValid() {
		return ""
	}
	file := string(rune('a' + p.Col))
	rank := string(rune('8' - p.Row))
	return file + rank
}

// FromAlgebraic creates a Position from algebraic notation.
func FromAlgebraic(s string) Position {
	if len(s) != 2 {
		return Position{-1, -1}
	}
	col := int(s[0] - 'a')
	row := 7 - int(s[1]-'1')
	if col < 0 || col > 7 || row < 0 || row > 7 {
		return Position{-1, -1}
	}
	return Position{row, col}
}

// Opponent returns the opposite color.
func (c Color) Opponent() Color {
	if c == White {
		return Black
	}
	return White
}
