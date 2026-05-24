package game

import (
	"fmt"
	"slices"
)

// GameSession represents a game with additional metadata.
type GameSession struct {
	*Game
	CurrentMove int
	PastMoves   []Move
}

// NewGameSession creates a new GameSession.
// The abstraction is useful for keeping track of game data without
// convoluting the engine logic with game metadata.
func NewGameSession() *GameSession {
	return &GameSession{
		Game:        NewGame(),
		CurrentMove: -1,
		PastMoves:   []Move{},
	}
}

// Play plays a move in the game session.
func (g *GameSession) Play(move Move) Piece {
	g.PastMoves = g.PastMoves[:g.CurrentMove+1]
	capturedPiece := g.Game.Play(move)
	g.PastMoves = append(g.PastMoves, move)
	g.CurrentMove++

	return capturedPiece
}

// SetCurrentMove sets the current move index.
func (g *GameSession) SetCurrentMove(moveIndex int) error {
	if moveIndex < 0 || moveIndex > len(g.PastMoves) {
		return fmt.Errorf("move index out of range")
	}
	g.CurrentMove = moveIndex

	g.Game = NewGame()
	for i := 0; i <= moveIndex; i++ {
		g.Game.Play(g.PastMoves[i])
	}

	return nil
}

// Copy returns a deep copy of the game session.
func (g *GameSession) Copy() *GameSession {
	return &GameSession{
		Game:        g.Game.Copy(),
		CurrentMove: g.CurrentMove,
		PastMoves:   slices.Clone(g.PastMoves),
	}
}
