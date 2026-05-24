package game_test

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/vpoliakov01/2v2ChessAI/engine/game"
)

func TestZobristInitialHashNonZero(t *testing.T) {
	g := game.NewGame()
	require.NotZero(t, g.Hash, "initial position should have a non-zero hash")
}

func TestZobristPlayUnplayRoundtrip(t *testing.T) {
	g := game.NewGame()
	originalHash := g.Hash

	moves := g.GetMoves(nil)
	require.NotEmpty(t, moves)

	for _, move := range moves[:10] {
		captured := g.Play(move)
		g.UnplayMove(move, captured)
		require.Equal(t, originalHash, g.Hash,
			"hash should return to the original after Play+UnplayMove for %v", move)
	}
}

func TestZobristIncrementalMatchesRecompute(t *testing.T) {
	g := game.NewGame()

	moves, err := game.ParsePGN(`
1. h2-h3 b8-c8 i13-i12 m8-l8
2. g1-j4 a8-d11 e13-e12 m5-l5
3. e2-e3 d11-a8 h14-k11 n7-l9`)
	require.NoError(t, err)

	for _, move := range moves {
		g.Play(move)

		incremental := g.Hash
		g.ComputeHash()
		require.Equal(t, incremental, g.Hash,
			"incremental hash should match recompute after %v", move)
	}
}

func TestZobristActivePlayerAffectsHash(t *testing.T) {
	g1 := game.NewGame()
	g2 := game.NewGame()

	require.Equal(t, g1.Hash, g2.Hash, "two fresh games should hash equal")

	g2.ActivePlayer = (g2.ActivePlayer + 1) % 4
	g2.ComputeHash()
	require.NotEqual(t, g1.Hash, g2.Hash,
		"changing only the active player should change the hash")
}
