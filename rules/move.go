// Package rules implements chess move validation and game rules.
package rules

import (
	"chess/board"
)

// Move represents a chess move.
type Move struct {
	From     board.Position
	To       board.Position
	Promotion board.PieceType // For pawn promotion
	IsCastling bool
	IsEnPassant bool
	CapturedPiece board.Piece
}

// MoveValidator validates chess moves according to the rules.
type MoveValidator struct {
	Board *board.Board
}

// NewMoveValidator creates a new move validator for the given board.
func NewMoveValidator(b *board.Board) *MoveValidator {
	return &MoveValidator{Board: b}
}

// GetValidMoves returns all valid moves for a piece at the given position.
func (mv *MoveValidator) GetValidMoves(pos board.Position, gameState *GameState) []Move {
	piece := mv.Board.GetPiece(pos.Row, pos.Col)
	if piece.Type == board.NoPiece {
		return nil
	}

	var moves []Move

	switch piece.Type {
	case board.Pawn:
		moves = mv.getPawnMoves(pos, piece.Color, gameState)
	case board.Knight:
		moves = mv.getKnightMoves(pos, piece.Color)
	case board.Bishop:
		moves = mv.getBishopMoves(pos, piece.Color)
	case board.Rook:
		moves = mv.getRookMoves(pos, piece.Color)
	case board.Queen:
		moves = mv.getQueenMoves(pos, piece.Color)
	case board.King:
		moves = mv.getKingMoves(pos, piece.Color, gameState)
	}

	// Filter out moves that would leave the king in check
	var validMoves []Move
	for _, move := range moves {
		if !mv.wouldLeaveKingInCheck(move, piece.Color) {
			validMoves = append(validMoves, move)
		}
	}

	return validMoves
}

// getPawnMoves returns valid pawn moves.
func (mv *MoveValidator) getPawnMoves(pos board.Position, color board.Color, gameState *GameState) []Move {
	var moves []Move
	direction := -1
	startRow := 6
	promotionRow := 0
	enPassantRow := 3
	
	if color == board.Black {
		direction = 1
		startRow = 1
		promotionRow = 7
		enPassantRow = 4
	}

	// Single push
	newRow := pos.Row + direction
	if newRow >= 0 && newRow <= 7 {
		target := mv.Board.GetPiece(newRow, pos.Col)
		if target.Type == board.NoPiece {
			move := Move{From: pos, To: board.Position{newRow, pos.Col}}
			if newRow == promotionRow {
				move.Promotion = board.Queen // Default promotion
			}
			moves = append(moves, move)

			// Double push from starting position
			if pos.Row == startRow {
				newRow2 := pos.Row + 2*direction
				target2 := mv.Board.GetPiece(newRow2, pos.Col)
				if target2.Type == board.NoPiece {
					moves = append(moves, Move{From: pos, To: board.Position{newRow2, pos.Col}})
				}
			}
		}
	}

	// Captures (including en passant)
	for _, dc := range []int{-1, 1} {
		newCol := pos.Col + dc
		if newCol >= 0 && newCol <= 7 && newRow >= 0 && newRow <= 7 {
			target := mv.Board.GetPiece(newRow, newCol)
			if target.Type != board.NoPiece && target.Color != color {
				move := Move{From: pos, To: board.Position{newRow, newCol}, CapturedPiece: target}
				if newRow == promotionRow {
					move.Promotion = board.Queen
				}
				moves = append(moves, move)
			}

			// En passant
			if pos.Row == enPassantRow && gameState != nil {
				enPassantTarget := gameState.EnPassantTarget
				if enPassantTarget.Row == newRow && enPassantTarget.Col == newCol {
					moves = append(moves, Move{
						From: pos,
						To: board.Position{newRow, newCol},
						IsEnPassant: true,
						CapturedPiece: board.Piece{Type: board.Pawn, Color: color.Opponent()},
					})
				}
			}
		}
	}

	return moves
}

// getKnightMoves returns valid knight moves.
func (mv *MoveValidator) getKnightMoves(pos board.Position, color board.Color) []Move {
	var moves []Move
	offsets := []struct{ dr, dc int }{
		{-2, -1}, {-2, 1}, {-1, -2}, {-1, 2},
		{1, -2}, {1, 2}, {2, -1}, {2, 1},
	}

	for _, offset := range offsets {
		newRow := pos.Row + offset.dr
		newCol := pos.Col + offset.dc
		if newRow >= 0 && newRow <= 7 && newCol >= 0 && newCol <= 7 {
			target := mv.Board.GetPiece(newRow, newCol)
			if target.Type == board.NoPiece || target.Color != color {
				move := Move{From: pos, To: board.Position{newRow, newCol}}
				if target.Type != board.NoPiece {
					move.CapturedPiece = target
				}
				moves = append(moves, move)
			}
		}
	}

	return moves
}

// getSlidingMoves returns moves for sliding pieces (bishop, rook, queen).
func (mv *MoveValidator) getSlidingMoves(pos board.Position, color board.Color, directions []struct{ dr, dc int }) []Move {
	var moves []Move

	for _, dir := range directions {
		for i := 1; i < 8; i++ {
			newRow := pos.Row + dir.dr*i
			newCol := pos.Col + dir.dc*i
			if newRow < 0 || newRow > 7 || newCol < 0 || newCol > 7 {
				break
			}
			target := mv.Board.GetPiece(newRow, newCol)
			if target.Type == board.NoPiece {
				moves = append(moves, Move{From: pos, To: board.Position{newRow, newCol}})
			} else {
				if target.Color != color {
					moves = append(moves, Move{From: pos, To: board.Position{newRow, newCol}, CapturedPiece: target})
				}
				break
			}
		}
	}

	return moves
}

// getBishopMoves returns valid bishop moves.
func (mv *MoveValidator) getBishopMoves(pos board.Position, color board.Color) []Move {
	directions := []struct{ dr, dc int }{
		{-1, -1}, {-1, 1}, {1, -1}, {1, 1},
	}
	return mv.getSlidingMoves(pos, color, directions)
}

// getRookMoves returns valid rook moves.
func (mv *MoveValidator) getRookMoves(pos board.Position, color board.Color) []Move {
	directions := []struct{ dr, dc int }{
		{-1, 0}, {1, 0}, {0, -1}, {0, 1},
	}
	return mv.getSlidingMoves(pos, color, directions)
}

// getQueenMoves returns valid queen moves.
func (mv *MoveValidator) getQueenMoves(pos board.Position, color board.Color) []Move {
	directions := []struct{ dr, dc int }{
		{-1, -1}, {-1, 0}, {-1, 1},
		{0, -1}, {0, 1},
		{1, -1}, {1, 0}, {1, 1},
	}
	return mv.getSlidingMoves(pos, color, directions)
}

// getKingMoves returns valid king moves including castling.
func (mv *MoveValidator) getKingMoves(pos board.Position, color board.Color, gameState *GameState) []Move {
	var moves []Move

	// Normal king moves
	directions := []struct{ dr, dc int }{
		{-1, -1}, {-1, 0}, {-1, 1},
		{0, -1}, {0, 1},
		{1, -1}, {1, 0}, {1, 1},
	}

	for _, dir := range directions {
		newRow := pos.Row + dir.dr
		newCol := pos.Col + dir.dc
		if newRow >= 0 && newRow <= 7 && newCol >= 0 && newCol <= 7 {
			target := mv.Board.GetPiece(newRow, newCol)
			if target.Type == board.NoPiece || target.Color != color {
				move := Move{From: pos, To: board.Position{newRow, newCol}}
				if target.Type != board.NoPiece {
					move.CapturedPiece = target
				}
				moves = append(moves, move)
			}
		}
	}

	// Castling
	if gameState != nil && !mv.IsInCheck(color) {
		row := 7
		if color == board.Black {
			row = 0
		}

		// Kingside castling
		if (color == board.White && gameState.WhiteCanCastleKingside) ||
			(color == board.Black && gameState.BlackCanCastleKingside) {
			if mv.Board.GetPiece(row, 5).Type == board.NoPiece &&
				mv.Board.GetPiece(row, 6).Type == board.NoPiece {
				// Check if squares are not attacked
				if !mv.isSquareAttacked(board.Position{row, 5}, color.Opponent()) &&
					!mv.isSquareAttacked(board.Position{row, 6}, color.Opponent()) {
					moves = append(moves, Move{
						From: pos,
						To: board.Position{row, 6},
						IsCastling: true,
					})
				}
			}
		}

		// Queenside castling
		if (color == board.White && gameState.WhiteCanCastleQueenside) ||
			(color == board.Black && gameState.BlackCanCastleQueenside) {
			if mv.Board.GetPiece(row, 1).Type == board.NoPiece &&
				mv.Board.GetPiece(row, 2).Type == board.NoPiece &&
				mv.Board.GetPiece(row, 3).Type == board.NoPiece {
				// Check if squares are not attacked
				if !mv.isSquareAttacked(board.Position{row, 2}, color.Opponent()) &&
					!mv.isSquareAttacked(board.Position{row, 3}, color.Opponent()) {
					moves = append(moves, Move{
						From: pos,
						To: board.Position{row, 2},
						IsCastling: true,
					})
				}
			}
		}
	}

	return moves
}

// isSquareAttacked checks if a square is attacked by the given color.
func (mv *MoveValidator) isSquareAttacked(pos board.Position, byColor board.Color) bool {
	// Check for pawn attacks
	pawnDir := 1
	if byColor == board.Black {
		pawnDir = -1
	}
	for _, dc := range []int{-1, 1} {
		pawnRow := pos.Row + pawnDir
		pawnCol := pos.Col + dc
		if pawnRow >= 0 && pawnRow <= 7 && pawnCol >= 0 && pawnCol <= 7 {
			p := mv.Board.GetPiece(pawnRow, pawnCol)
			if p.Type == board.Pawn && p.Color == byColor {
				return true
			}
		}
	}

	// Check for knight attacks
	knightOffsets := []struct{ dr, dc int }{
		{-2, -1}, {-2, 1}, {-1, -2}, {-1, 2},
		{1, -2}, {1, 2}, {2, -1}, {2, 1},
	}
	for _, offset := range knightOffsets {
		newRow := pos.Row + offset.dr
		newCol := pos.Col + offset.dc
		if newRow >= 0 && newRow <= 7 && newCol >= 0 && newCol <= 7 {
			p := mv.Board.GetPiece(newRow, newCol)
			if p.Type == board.Knight && p.Color == byColor {
				return true
			}
		}
	}

	// Check for king attacks
	kingOffsets := []struct{ dr, dc int }{
		{-1, -1}, {-1, 0}, {-1, 1},
		{0, -1}, {0, 1},
		{1, -1}, {1, 0}, {1, 1},
	}
	for _, offset := range kingOffsets {
		newRow := pos.Row + offset.dr
		newCol := pos.Col + offset.dc
		if newRow >= 0 && newRow <= 7 && newCol >= 0 && newCol <= 7 {
			p := mv.Board.GetPiece(newRow, newCol)
			if p.Type == board.King && p.Color == byColor {
				return true
			}
		}
	}

	// Check for sliding piece attacks (bishop, rook, queen)
	diagonals := []struct{ dr, dc int }{{-1, -1}, {-1, 1}, {1, -1}, {1, 1}}
	for _, dir := range diagonals {
		for i := 1; i < 8; i++ {
			newRow := pos.Row + dir.dr*i
			newCol := pos.Col + dir.dc*i
			if newRow < 0 || newRow > 7 || newCol < 0 || newCol > 7 {
				break
			}
			p := mv.Board.GetPiece(newRow, newCol)
			if p.Type != board.NoPiece {
				if p.Color == byColor && (p.Type == board.Bishop || p.Type == board.Queen) {
					return true
				}
				break
			}
		}
	}

	straights := []struct{ dr, dc int }{{-1, 0}, {1, 0}, {0, -1}, {0, 1}}
	for _, dir := range straights {
		for i := 1; i < 8; i++ {
			newRow := pos.Row + dir.dr*i
			newCol := pos.Col + dir.dc*i
			if newRow < 0 || newRow > 7 || newCol < 0 || newCol > 7 {
				break
			}
			p := mv.Board.GetPiece(newRow, newCol)
			if p.Type != board.NoPiece {
				if p.Color == byColor && (p.Type == board.Rook || p.Type == board.Queen) {
					return true
				}
				break
			}
		}
	}

	return false
}

// FindKing finds the position of the king of the given color.
func (mv *MoveValidator) FindKing(color board.Color) board.Position {
	for row := 0; row < 8; row++ {
		for col := 0; col < 8; col++ {
			p := mv.Board.GetPiece(row, col)
			if p.Type == board.King && p.Color == color {
				return board.Position{row, col}
			}
		}
	}
	return board.Position{-1, -1}
}

// IsInCheck checks if the given color's king is in check.
func (mv *MoveValidator) IsInCheck(color board.Color) bool {
	kingPos := mv.FindKing(color)
	if !kingPos.IsValid() {
		return false
	}
	return mv.isSquareAttacked(kingPos, color.Opponent())
}

// wouldLeaveKingInCheck checks if a move would leave the king in check.
func (mv *MoveValidator) wouldLeaveKingInCheck(move Move, color board.Color) bool {
	// Make a copy of the board
	testBoard := mv.Board.Copy()

	// Apply the move on the copy
	piece := testBoard.GetPiece(move.From.Row, move.From.Col)
	testBoard.SetPiece(move.From.Row, move.From.Col, board.Piece{Type: board.NoPiece, Color: board.NoColor})
	testBoard.SetPiece(move.To.Row, move.To.Col, piece)

	// Handle en passant capture
	if move.IsEnPassant {
		captureRow := move.From.Row
		testBoard.SetPiece(captureRow, move.To.Col, board.Piece{Type: board.NoPiece, Color: board.NoColor})
	}

	// Handle castling
	if move.IsCastling {
		row := move.To.Row
		if move.To.Col == 6 { // Kingside
			rook := testBoard.GetPiece(row, 7)
			testBoard.SetPiece(row, 7, board.Piece{Type: board.NoPiece, Color: board.NoColor})
			testBoard.SetPiece(row, 5, rook)
		} else { // Queenside
			rook := testBoard.GetPiece(row, 0)
			testBoard.SetPiece(row, 0, board.Piece{Type: board.NoPiece, Color: board.NoColor})
			testBoard.SetPiece(row, 3, rook)
		}
	}

	// Check if king is in check
	testValidator := NewMoveValidator(testBoard)
	return testValidator.IsInCheck(color)
}

// HasAnyValidMoves checks if the given color has any valid moves.
func (mv *MoveValidator) HasAnyValidMoves(color board.Color, gameState *GameState) bool {
	for row := 0; row < 8; row++ {
		for col := 0; col < 8; col++ {
			p := mv.Board.GetPiece(row, col)
			if p.Type != board.NoPiece && p.Color == color {
				moves := mv.GetValidMoves(board.Position{row, col}, gameState)
				if len(moves) > 0 {
					return true
				}
			}
		}
	}
	return false
}

// IsCheckmate checks if the given color is in checkmate.
func (mv *MoveValidator) IsCheckmate(color board.Color, gameState *GameState) bool {
	return mv.IsInCheck(color) && !mv.HasAnyValidMoves(color, gameState)
}

// IsStalemate checks if the given color is in stalemate.
func (mv *MoveValidator) IsStalemate(color board.Color, gameState *GameState) bool {
	return !mv.IsInCheck(color) && !mv.HasAnyValidMoves(color, gameState)
}
