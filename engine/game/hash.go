package game

import "math/rand"

const (
	seed = int64(0xC0FFEE5)
)

var (
	pieceHashes        [pieceVariants][BoardSize][BoardSize]uint64
	activePlayerHashes [4]uint64
)

func init() {
	r := rand.New(rand.NewSource(seed))

	for p := 0; p < pieceVariants; p++ {
		for rank := 0; rank < BoardSize; rank++ {
			for file := 0; file < BoardSize; file++ {
				pieceHashes[p][rank][file] = r.Uint64()
			}
		}
	}

	for i := 0; i < 4; i++ {
		activePlayerHashes[i] = r.Uint64()
	}
}

// ComputeHash recomputes the Zobrist hash from scratch.
// Play and UnplayMove maintain it incrementally.
func (g *Game) ComputeHash() {
	hash := uint64(0)

	for rank := 0; rank < BoardSize; rank++ {
		for file := 0; file < BoardSize; file++ {
			piece := g.Board.Grid[rank][file]

			if piece == EmptySquare || piece == InactiveSquare {
				continue
			}

			hash ^= pieceHashes[piece][rank][file]
		}
	}

	hash ^= activePlayerHashes[g.ActivePlayer]
	g.Hash = hash
}
