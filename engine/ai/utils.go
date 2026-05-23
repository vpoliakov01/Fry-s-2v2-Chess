package ai

import (
	"math"
	"sort"
	"sync"

	"github.com/vpoliakov01/2v2ChessAI/engine/game"
)

// candidateResult carries one root-move worker's findings back to GetBestMove.
type candidateResult struct {
	score        float64
	continuation []game.Move // detached from any buffer; safe to keep.
}

// getMoveEvals returns the active player's moves sorted by 1-ply evaluation.
func (ai *AI) getMoveEvals(g *game.Game, buffer *buffer, depth int) []moveScore {
	moves := g.GetMoves(buffer.moves[depth][:0])
	buffer.moves[depth] = moves // In case moves got reallocated by append inside GetMoves.

	moveEvals := buffer.moveEvals[depth][:len(moves)]

	for i := range moves {
		capturedPiece := g.Play(moves[i])
		moveEvals[i] = moveScore{moves[i], -ai.EvaluateCurrent(g, buffer)}
		g.UnplayMove(moves[i], capturedPiece)
	}

	buffer.moveEvals[depth] = moveEvals

	sort.Slice(moveEvals, func(a, b int) bool {
		return moveEvals[a].score > moveEvals[b].score
	})

	return moveEvals
}

// searchRootMove plays the candidate move, runs Negamax on it, and returns the score and continuation.
func (ai *AI) searchRootMove(g *game.Game, buffer *buffer, cpu int, candidate moveScore, alpha, beta float64) (score float64, continuation []game.Move) {
	move := candidate.move

	capturedPiece := g.Play(move)
	opponentScore := ai.Negamax(g, buffer, cpu, 2, -candidate.score, -beta, -alpha)
	g.UnplayMove(move, capturedPiece)

	score = fromOpponentScore(opponentScore)
	ai.raiseSharedAlpha(score)

	childCont := buffer.continuation[2]
	continuation = make([]game.Move, 0, len(childCont)+1)
	continuation = append(continuation, move)
	continuation = append(continuation, childCont...)

	return score, continuation
}

// searchAtDepth runs one iterative-deepening iteration at the given target depth.
func (ai *AI) searchAtDepth(g *game.Game, depth int) (continuation []game.Move, score float64, err error) {
	ai.bfsDepth = depth

	buffer := &ai.buffers[0]
	ai.sharedAlpha.Store(math.Float64bits(-(mateValue + 1)))

	moveEvals := ai.getMoveEvals(g, buffer, 1)
	if len(moveEvals) == 0 {
		return nil, 0, ErrNoMoves
	}

	// Try the previous iteration's best move first.
	cached, ok := ai.cache.Get(g.Hash)
	if ok {
		for i := range moveEvals {
			if moveEvals[i].move == cached.move() {
				// Shift all prev indexes down by one, and put the cached move at index 0.
				cachedMove := moveEvals[i]
				for j := i; j > 0; j-- {
					moveEvals[j] = moveEvals[j-1]
				}
				moveEvals[0] = cachedMove

				break
			}
		}
	}

	bestScore, bestContinuation := ai.searchRootMovesParallel(g, moveEvals, mateValue+1)

	if len(bestContinuation) > 0 && !ai.stopFlag.Load() {
		ai.cache.Set(g.Hash, bestContinuation[0], bestScore, int8(depth-1), BoundExact)
	}

	return bestContinuation, bestScore, nil
}

// searchRootMovesParallel searches the candidates concurrently — one goroutine per
// candidate, each on its own game copy and buffer. Returns the best score and continuation.
func (ai *AI) searchRootMovesParallel(g *game.Game, candidates []moveScore, beta float64) (float64, []game.Move) {
	// Pool of CPUs for the goroutines.
	cpuIDs := make(chan int, cpus)
	for i := range cpus {
		cpuIDs <- i
	}

	results := make([]candidateResult, len(candidates))
	var wg sync.WaitGroup

	for i, candidate := range candidates {
		wg.Add(1)

		go func(slot int, candidate moveScore) {
			defer wg.Done()
			cpuID := <-cpuIDs
			defer func() { cpuIDs <- cpuID }()

			gameCopy := g.Copy()
			alpha := ai.loadSharedAlpha()

			score, continuation := ai.searchRootMove(gameCopy, &ai.buffers[cpuID], cpuID, candidate, alpha, beta)
			results[slot] = candidateResult{score: score, continuation: continuation}
		}(i, candidate)
	}
	wg.Wait()

	bestScore := -math.MaxFloat64
	var bestContinuation []game.Move

	for _, result := range results {
		if result.score > bestScore {
			bestScore = result.score
			bestContinuation = result.continuation
		}
	}

	return bestScore, bestContinuation
}
