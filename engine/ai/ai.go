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

	enableDebug        bool
	enableDebugLogging bool
	BestMoves          [][]BestmoveDataAvgAcc
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

	if ai.enableDebug {
		ai.InitDebug()
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
	ai.hasStopped.Store(false)
	defer ai.StoreCache()
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
	return bestContinuation, bestScore, nil
}

// Negamax (minimax + negation) recursively finds the position
// reached by each side picking their best move.
// Alpha and beta params are used for alpha-beta pruning (skipping evalution
// of branches that are guaranteed not to be picked by any of players).
func (ai *AI) Negamax(g *game.Game, buffer *buffer, cpu, depth int, eval, alpha, beta float64) (score float64) {
	buffer.continuation[depth] = buffer.continuation[depth][:0] // Reset the buffer.

	// Check base cases.
	if g.HasEnded() {
		return -mateValue
	}
	if depth > ai.bfsDepth {
		return eval
	}

	remainingDepth := int8(ai.bfsDepth - depth)
	alphaOrig := alpha

	// Check the transposition table for a previously computed result.
	var cachedMove game.Move

	cached, ok := ai.cache.Get(g.Hash)
	if ok {
		cachedMove = cached.move()

		if cached.depth >= remainingDepth && canCutoff(cached.score, cached.bound, alpha, beta) {
			buffer.continuation[depth] = append(buffer.continuation[depth][:0], cachedMove)
			return float64(cached.score)
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

	if ai.enableDebug {
		// Locate bestMove in the full sorted move list to report its ordering index.
		moveEvals := buffer.moveEvals[depth]
		bestMoveIndex := 0
		for i := range moveEvals {
			if moveEvals[i].move == bestMove {
				bestMoveIndex = i
				break
			}
		}
		ai.recordBestMove(BestMoveData{
			Depth:      depth,
			MoveIndex:  bestMoveIndex,
			TotalMoves: len(moveEvals),
			ScoreDelta: bestScore - moveEvals[0].posEval,
		}, cpu)
	}

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

	if ai.enableDebugLogging && depth == 1 {
		for i, moveEval := range moveEvals {
			move := moveEval.move
			fmt.Println(i, " ", game.HumanReadableMove(board, move), "eval:", moveEval.posEval)
		}
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
