# Terminal Chess Game

A fully offline terminal-based chess game written in Go with a large, clearly visible chess board.

## Features

- **Interactive Terminal UI**: Large, colorful chess board that's easy to read
- **Keyboard Controls**: Navigate the board and make moves using keyboard
- **Full Chess Rules**: 
  - Valid piece movement for all pieces
  - Turn alternation
  - Captures
  - Check and checkmate detection
  - Stalemate detection
  - Castling (kingside and queenside)
  - En passant
  - Pawn promotion (choose from Queen, Rook, Bishop, Knight)
- **Move History**: View all moves made during the game
- **Undo Functionality**: Undo moves to try different strategies
- **Visual Highlights**:
  - Cursor position highlighted
  - Selected piece highlighted
  - Valid move squares highlighted
  - Last move highlighted
  - King in check highlighted

## Installation

```bash
go build -o chess ./cmd/chess/
```

## Running

```bash
./chess
```

## Controls

| Key | Action |
|-----|--------|
| `h` / `←` | Move cursor left |
| `j` / `↓` | Move cursor down |
| `k` / `↑` | Move cursor up |
| `l` / `→` | Move cursor right |
| `Enter` / `Space` | Select piece / Make move |
| `u` | Undo last move |
| `v` | View move history |
| `r` | Start new game |
| `q` / `Ctrl+C` | Quit |

## Pawn Promotion

When a pawn reaches the opposite end of the board, you'll be prompted to choose a piece for promotion:

- `q` - Queen
- `r` - Rook  
- `b` - Bishop
- `n` - Knight

## Project Structure

```
chess/
├── board/          # Board and piece types
│   └── board.go    # Core board representation
├── rules/          # Chess rules and move validation
│   ├── move.go     # Move validation logic
│   └── gamestate.go # Game state tracking
├── game/           # Game management
│   └── game.go     # Game state and history
├── ui/             # Terminal UI
│   └── ui.go       # Bubble Tea UI implementation
├── cmd/chess/      # Entry point
│   └── main.go     # Main function
├── go.mod          # Go module file
├── go.sum          # Dependency checksums
└── README.md       # This file
```

## Dependencies

- [bubbletea](https://github.com/charmbracelet/bubbletea) - Terminal UI framework
- [lipgloss](https://github.com/charmbracelet/lipgloss) - Style definitions

## Requirements

- Go 1.21 or later
- Terminal with color support

## License

MIT License
