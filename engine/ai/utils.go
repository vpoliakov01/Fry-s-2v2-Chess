package ai

import (
	"math"
	"sync"

	"github.com/vpoliakov01/2v2ChessAI/engine/game"
)

// candidateResult carries one root-move worker's findings back to GetBestMove.
type candidateResult struct {
	score        float64
	continuation []game.Move // detached from any buffer; safe to keep.
}

// searchAtDepth runs one iterative-deepening iteration at the given target depth.
func (ai *AI) searchAtDepth(g *game.Game, depth int) (continuation []game.Move, score float64, err error) {
	ai.bfsDepth = depth

	buffer := &ai.buffers[0]
	ai.sharedAlpha.Store(math.Float64bits(-mateValue))

	eval := 0.0

	cachedMove := game.NullMove
	cached, ok := ai.cache.Get(g.Hash)
	if ok {
		cachedMove = cached.move()
		eval = cached.eval()
	} else {
		eval = ai.EvaluateCurrent(g, buffer)
	}

	movesToSearch := ai.GetMovesToSearch(g, buffer, 1, eval, cachedMove)
	if len(movesToSearch) == 0 {
		return nil, 0, ErrNoMoves
	}

	bestScore, bestContinuation := ai.searchRootMovesParallel(g, movesToSearch, mateValue+1)

	if len(bestContinuation) > 0 && !ai.stopFlag.Load() {
		ai.cache.Set(g.Hash, bestContinuation[0], bestScore, int8(depth-1), BoundExact, g.MoveNumber)
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

	bestScore := -mateValue
	var bestContinuation []game.Move

	for _, result := range results {
		if result.score > bestScore {
			bestScore = result.score
			bestContinuation = result.continuation
		}
	}

	if ai.debugConfig != nil {
		ai.debugConfig.captureSearchResults(results, ai.bfsDepth)
	}

	return bestScore, bestContinuation
}

// searchRootMove plays the candidate move, runs Negamax on it, and returns the score and continuation.
func (ai *AI) searchRootMove(g *game.Game, buffer *buffer, cpu int, candidate moveScore, alpha, beta float64) (score float64, continuation []game.Move) {
	move := candidate.move

	capturedPiece := g.Play(move)
	opponentScore := ai.Negamax(g, buffer, cpu, 2, -candidate.posEval, -beta, -alpha)
	g.UnplayMove(move, capturedPiece)

	score = fromOpponentScore(opponentScore)
	ai.raiseSharedAlpha(score)

	childCont := buffer.continuation[2]
	continuation = make([]game.Move, 0, 1+len(childCont))
	continuation = append(continuation, move)
	continuation = append(continuation, childCont...)

	return score, continuation
}

// Selection sort the top moves for performance.
func sortMoveEvals(moveEvals []moveScore) {
	for i := 0; i < len(moveEvals); i++ {
		maxIndex := i
		maxScore := moveEvals[i].score

		// Find the max to the right
		for j := i + 1; j < len(moveEvals); j++ {
			if moveEvals[j].score > maxScore {
				maxIndex = j
				maxScore = moveEvals[j].score
			}
		}

		// Swap
		if maxIndex != i {
			moveEvals[i], moveEvals[maxIndex] = moveEvals[maxIndex], moveEvals[i]
		}
	}
}
