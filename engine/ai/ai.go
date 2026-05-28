package ai

import (
	"fmt"
	"math"
	"sync/atomic"

	"github.com/vpoliakov01/2v2ChessAI/engine/game"
)

// AI is the ai engine used for evaluating the position and picking the best move.
type AI struct {
	Depth      int
	Spread     int
	SpreadDrop int

	EvalsCount int // Populated after GetBestMove returns.
	EvalLimit  int

	buffers           []buffer // One buffer per CPU.
	cache             *Cache
	enableStoredCache bool

	bfsDepth int // Depth of the running BFS.

	stopFlag    atomic.Bool
	hasStopped  atomic.Bool // True when GetBestMove is not running.
	sharedAlpha atomic.Uint64

	debugConfig *DebugConfig
	BestMoves   [][]bestmoveDataAvgAcc
}

// New creates a new AI.
func New(depth, spread, spreadDrop, evalLimit int, options ...func(*AI)) *AI {
	if evalLimit == 0 {
		evalLimit = MaxEvalLimit
	}

	ai := &AI{
		Depth:      depth,
		Spread:     spread,
		SpreadDrop: spreadDrop,
		EvalLimit:  evalLimit,
	}
	ai.hasStopped.Store(true)
	for _, option := range options {
		option(ai)
	}

	if err := ai.LoadCache(); err != nil {
		fmt.Printf("Failed to load transposition table: %v\n", err)
	}
	ai.installShutdownHook()

	return ai
}

// GetBestMove returns the predicted continuation up to the search depth.
// The first element of the continuation is the best move itself.
func (ai *AI) GetBestMove(g *game.Game) (continuation []game.Move, score float64, err error) {
	ai.InitDebug(g)

	ai.hasStopped.Store(false)
	defer ai.hasStopped.Store(true)

	ai.stopFlag.Store(false)
	ai.EvalsCount = 0

	if g.HasEnded() {
		return nil, float64(g.Winner), ErrGameEnded
	}

	ai.initBuffers()
	if ai.cache == nil {
		ai.cache = NewCache()
	}
	g.ComputeHash()

	bestContinuation := []game.Move{}
	bestScore := 0.0
	haveResult := false

	depthStep := 1
	for depth := 1; depth < ai.Depth+depthStep; depth += depthStep {
		depth = min(depth, ai.Depth)

		cont, score, err := ai.searchAtDepth(g, depth)
		if err != nil {
			if !haveResult {
				ai.sumEvalsCounts()
				return nil, 0, err
			}
			break
		}

		// Only commit a partial result if we have nothing to fall back to.
		if !ai.stopFlag.Load() || !haveResult {
			bestContinuation = cont
			bestScore = score
			haveResult = true
		}

		if ai.stopFlag.Load() || math.Abs(bestScore) >= mateValue-float64(depth) {
			break
		}
	}

	ai.sumEvalsCounts()
	if ai.debugConfig != nil {
		ai.PrintSearchResults(g)
	}
	return bestContinuation, bestScore, nil
}

// Negamax (minimax + negation) recursively finds the position
// reached by each side picking their best move.
// Alpha and beta params are used for alpha-beta pruning (skipping evalution
// of branches that are guaranteed not to be picked by any of players).
func (ai *AI) Negamax(g *game.Game, buffer *buffer, cpu, depth int, eval, alpha, beta float64) (score float64) {

	// Check base cases.
	if depth > ai.bfsDepth {
		return ai.EvaluateCurrent(g, buffer)
		// return eval
	}

	buffer.continuation[depth] = buffer.continuation[depth][:0] // Reset the buffer.
	remainingDepth := int8(ai.bfsDepth - depth)
	alphaOrig := alpha

	cachedMove := game.NullMove
	cached, ok := ai.cache.Get(g.Hash)
	if ok {
		cachedMove = cached.Move()
		eval = cached.Eval()

		if cached.Depth >= remainingDepth && canCutoff(eval, alpha, beta, cached.Bound) {
			buffer.continuation[depth] = append(buffer.continuation[depth][:0], cachedMove)
			return eval
		}
	}

	movesToSearch := ai.GetMovesToSearch(g, buffer, depth, eval, cachedMove)

	bestScore := -mateValue - 1
	bestMove := game.Move{}

	for _, ms := range movesToSearch {
		// If other workers updated alpha, tighten.
		if depth == 2 {
			newBeta := -ai.loadSharedAlpha()
			beta = math.Min(beta, newBeta)
			if alpha >= beta {
				return alpha
			}
		}

		move := ms.move
		childEval := -ms.posEval

		capturedPiece := g.Play(move)
		opponentScore := ai.Negamax(g, buffer, cpu, depth+1, childEval, -beta, -alpha)
		g.UnplayMove(move, capturedPiece)

		score := fromOpponentScore(opponentScore)

		if score > bestScore {
			bestScore = score
			bestMove = move

			continuation := buffer.continuation[depth][:0]
			continuation = append(continuation, move)
			buffer.continuation[depth] = append(continuation, buffer.continuation[depth+1]...)
		}

		if bestScore > alpha {
			alpha = bestScore
		}

		if alpha >= beta || buffer.evalsCount >= ai.EvalLimit || ai.stopFlag.Load() {
			break
		}
	}

	if !ai.stopFlag.Load() {
		ai.cache.Set(g.Hash, bestMove, bestScore, remainingDepth, boundOf(bestScore, alphaOrig, beta), g.MoveNumber)
	}

	ai.recordBestMove(buffer, depth, bestMove, bestScore, cpu)

	return bestScore
}

// GetMovesToSearch returns the moves worth searching, each tagged with its post-move absolute eval.
func (ai *AI) GetMovesToSearch(g *game.Game, buffer *buffer, depth int, eval float64, firstMove game.Move) []moveScore {
	moves := g.GetMoves(buffer.moves[depth][:0])
	buffer.moves[depth] = moves // In case moves got reallocated by append inside GetMoves.
	board := g.Board

	// Evaluate moves.
	moveEvals := buffer.moveEvals[depth][:len(moves)]
	for i, move := range moves {
		posEval, moveEval := ai.EvaluateMove(g, buffer, eval, move)
		moveEvals[i] = moveScore{move, posEval, moveEval}
	}
	buffer.moveEvals[depth] = moveEvals

	// Decide how many moves to pick.
	movesLeft := len(moveEvals)
	if depth > 1 {
		movesLeft = max(ai.Spread-depth/4*ai.SpreadDrop, 1)
	}
	capturesLeft := movesLeft/2 + 1

	sortMoveEvals(moveEvals)

	movesToSearch := buffer.movesToSearch[depth][:0]

	if ai.debugConfig != nil {
		ai.PrintMoveOrderingStats(g, moveEvals, eval, movesLeft, firstMove)
	}

	// Append the most promising external move.
	if firstMove != game.NullMove {
		for _, moveEval := range moveEvals {
			if moveEval.move == firstMove {
				movesToSearch = append(movesToSearch, moveEval)

				if !board.GetPiece(firstMove.To).IsEmpty() { // Capture
					capturesLeft--
				}

				movesLeft--
				break
			}
		}
	}

	// Append the rest of the moves.
	for _, moveEval := range moveEvals {
		if movesLeft == 0 {
			break
		}

		if moveEval.move == firstMove {
			continue
		}

		if !board.GetPiece(moveEval.move.To).IsEmpty() { // Capture
			if capturesLeft == 0 {
				continue
			}
			capturesLeft--
		}

		movesLeft--
		movesToSearch = append(movesToSearch, moveEval)
	}

	buffer.movesToSearch[depth] = movesToSearch
	return movesToSearch
}
