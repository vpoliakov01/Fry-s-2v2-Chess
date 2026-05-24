package game

import (
	"fmt"
)

// Game represents a state of the game.
type Game struct {
	ActivePlayer Player
	Board        *Board
	Winner       Team // Red/Yellow win: 1, Blue/Green win: -1.
	MoveNumber   int

	Hash         uint64   `json:"-"` // Zobrist hash, maintained incrementally by Play/UnplayMove.
	SquareBuffer []Square `json:"-"` // Reusable per-piece destination buffer for GetMoves; not shared across copies.
}

// NewGame creates a new Game.
func NewGame() *Game {
	g := Game{
		ActivePlayer: 0,
		Board:        NewBoard(),
		Winner:       0,
		MoveNumber:   1,
	}

	g.Board.SetStartingPosition()
	g.ComputeHash()

	return &g
}

// GetMoves appends all moves available to the active player to dst and returns the extended slice.
func (g *Game) GetMoves(dst []Move) []Move {
	if g.HasEnded() {
		return dst
	}

	for _, from := range g.Board.PieceSquares[g.ActivePlayer] {
		piece := g.Board.GetPiece(from)
		g.SquareBuffer = piece.GetMoves(g.Board, from, g.SquareBuffer[:0])

		for _, to := range g.SquareBuffer {
			dst = append(dst, Move{from, to})
		}
	}

	return dst
}

// Play plays a move in the game.
func (g *Game) Play(move Move) Piece {
	movingPiece := g.Board.GetPiece(move.From)
	capturedPiece := g.Board.GetPiece(move.To)

	g.Hash ^= pieceHashes[movingPiece][move.From.Rank][move.From.File]
	g.Hash ^= pieceHashes[movingPiece][move.To.Rank][move.To.File]

	if !capturedPiece.IsEmpty() {
		if capturedPiece.Kind() == KindKing {
			g.Winner = g.ActivePlayer.Team()
		}

		g.Hash ^= pieceHashes[capturedPiece][move.To.Rank][move.To.File]
	}

	g.Board.Move(move)
	g.MoveNumber++

	g.Hash ^= activePlayerHashes[g.ActivePlayer]
	g.ActivePlayer = (g.ActivePlayer + 1) % 4
	g.Hash ^= activePlayerHashes[g.ActivePlayer]

	return capturedPiece
}

// UnplayMove undoes a move in the game.
func (g *Game) UnplayMove(move Move, capturedPiece Piece) {
	g.Hash ^= activePlayerHashes[g.ActivePlayer]
	g.ActivePlayer = (g.ActivePlayer + 3) % 4
	g.Hash ^= activePlayerHashes[g.ActivePlayer]

	g.MoveNumber--
	g.Board.Unmove(move, capturedPiece)

	movingPiece := g.Board.GetPiece(move.From)
	g.Hash ^= pieceHashes[movingPiece][move.From.Rank][move.From.File]
	g.Hash ^= pieceHashes[movingPiece][move.To.Rank][move.To.File]

	if !capturedPiece.IsEmpty() {
		if capturedPiece.Kind() == KindKing {
			g.Winner = 0
		}

		g.Hash ^= pieceHashes[capturedPiece][move.To.Rank][move.To.File]
	}
}

// HasKing checks if the player still has a king.
func (g *Game) HasKing(player Player) bool {
	for _, square := range g.Board.PieceSquares[player] {
		if g.Board.GetPiece(square).Kind() == KindKing {
			return true
		}
	}
	return false
}

// HasEnded returns whether the game has ended.
func (g *Game) HasEnded() bool {
	return g.Winner != 0
}

// Copy returns a deep copy of the game.
func (g *Game) Copy() *Game {
	newGame := *g
	newGame.Board = g.Board.Copy()
	newGame.SquareBuffer = nil // Don't share scratch buffer with the source.
	return &newGame
}

// ValidateMove validates the move.
func (g *Game) ValidateMove(move *Move) error {
	if move == nil || !move.From.IsValid() || !move.To.IsValid() {
		return fmt.Errorf("move %v is invalid", move)
	}

	for _, m := range g.GetMoves(nil) {
		if m.From == move.From && m.To == move.To {
			return nil
		}
	}

	return fmt.Errorf("move %v is not available to the player", move)
}
