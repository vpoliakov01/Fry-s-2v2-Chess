package ai

import "github.com/vpoliakov01/2v2ChessAI/engine/game"

const (
	MovesUpperBound = 256
)

// buffer holds per-worker per-depth reusable storage to avoid repeated allocations.
type buffer struct {
	moves         [][]game.Move
	moveEvals     [][]moveScore
	movesToSearch [][]moveScore
	continuation  [][]game.Move

	evalsCount int
}

// init populates buffers for searches up to maxDepth.
func (buff *buffer) init(maxDepth int) {
	buff.evalsCount = 0

	if len(buff.moves) >= maxDepth {
		return
	}

	buff.moves = make([][]game.Move, maxDepth)
	buff.moveEvals = make([][]moveScore, maxDepth)
	buff.movesToSearch = make([][]moveScore, maxDepth)
	buff.continuation = make([][]game.Move, maxDepth)

	for d := range buff.continuation {
		buff.continuation[d] = make([]game.Move, 0, maxDepth)
	}

	for i := range buff.moves {
		buff.moves[i] = make([]game.Move, 0, MovesUpperBound)
		buff.moveEvals[i] = make([]moveScore, 0, MovesUpperBound)
		buff.movesToSearch[i] = make([]moveScore, 0, MovesUpperBound)
	}
}

// initBuffers lazily allocates one buffer per CPU, sized for the current configuration.
func (ai *AI) initBuffers() {
	if len(ai.buffers) < cpus {
		ai.buffers = make([]buffer, cpus)
	}

	for i := range ai.buffers { // Per each cpu.
		ai.buffers[i].init(ai.Depth + 2)
	}
}
