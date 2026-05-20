package ai

import (
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

	buffers []buffer // One buffer per CPU.

	stopFlag    atomic.Bool
	sharedAlpha atomic.Uint64

	enableDebug bool
	BestMoves   [][]BestmoveDataAvgAcc
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
	for _, option := range options {
		option(ai)
	}

	if ai.enableDebug {
		ai.InitDebug()
	}

	return ai
}

// GetBestMove returns the predicted continuation up to the search depth.
// The first element of the continuation is the best move itself.
func (ai *AI) GetBestMove(g *game.Game) (continuation []game.Move, score float64, err error) {
	ai.stopFlag.Store(false)
	ai.EvalsCount = 0

	if g.HasEnded() {
		return nil, float64(g.Winner), ErrGameEnded
	}

	ai.initBuffers()

	buffer := &ai.buffers[0]
	forcedMateScore := 1002 - float64(ai.Depth)
	alpha := -forcedMateScore
	beta := forcedMateScore
	ai.sharedAlpha.Store(math.Float64bits(alpha))

	moveEvals := ai.getMoveEvals(g, buffer, 1)
	if len(moveEvals) == 0 {
		ai.sumEvalsCounts()
		return nil, 0, ErrNoMoves
	}

	// YBW: search the highest-scored move to establish alpha
	bestScore, bestContinuation := ai.searchRootMove(g, buffer, 0, moveEvals[0], alpha, beta)
	alpha = math.Max(alpha, bestScore)

	// Parallel search of the remaining moves with tightened alpha
	if alpha < beta && !ai.stopFlag.Load() && len(moveEvals) > 1 {
		bestScore, bestContinuation = ai.searchRootMovesParallel(g, moveEvals[1:], beta, bestScore, bestContinuation)
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
		return float64(-1002 + depth)
	}
	if depth > ai.Depth {
		return eval
	}

	moveEvals := ai.getMoveEvals(g, buffer, depth)

	// Filter promising moves to actually search.
	moveIndexesToSearch := ai.GetMoveIndexesToSearch(g, moveEvals, depth, buffer.moveIndexesToSearch[depth][:0])
	buffer.moveIndexesToSearch[depth] = moveIndexesToSearch

	bestMoveIndex := moveIndexesToSearch[0]
	bestScore := -math.MaxFloat64

	for _, i := range moveIndexesToSearch {
		// If other workers updated alpha, tighten.
		if depth == 2 {
			newBeta := -ai.loadSharedAlpha()
			beta = math.Min(beta, newBeta)
			if alpha >= beta {
				return alpha
			}
		}

		move := moveEvals[i].move
		childEval := -moveEvals[i].score

		capturedPiece := g.Play(move)
		opponentScore := ai.Negamax(g, buffer, cpu, depth+1, childEval, -beta, -alpha)
		g.UnplayMove(move, capturedPiece)

		score := -opponentScore
		moveEvals[i].score = score

		if score > bestScore {
			bestScore = score
			bestMoveIndex = i

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

	if ai.enableDebug {
		ai.recordBestMove(BestMoveData{
			Depth:      depth,
			MoveIndex:  bestMoveIndex,
			TotalMoves: len(moveEvals),
			ScoreDelta: moveEvals[bestMoveIndex].score - moveEvals[0].score,
		}, cpu)
	}

	return bestScore
}

// EvaluateCurrent returns the difference between strengths of the team making the move and the opponent team.
// Increments the worker's per-buffer eval count to avoid the shared-counter cache-line contention under parallel search.
func (ai *AI) EvaluateCurrent(g *game.Game, buffer *buffer) float64 {
	buffer.evalsCount++
	playerStrengths := [4]float64{}

	if g.HasEnded() {
		return float64(g.ActivePlayer.Team()*g.Winner) * 1000
	}

	// For each piece, run piece strength evaluation.
	for player := range g.Board.PieceSquares {
		for square := range g.Board.PieceSquares[player] {
			piece := g.Board.GetPiece(square)
			playerStrengths[player] += piece.GetStrength(g.Board, square, player)
		}
	}

	redYellowStrength := playerStrengths[0] + playerStrengths[2] - math.Abs(playerStrengths[0]-playerStrengths[2])/3
	blueGreenStrength := playerStrengths[1] + playerStrengths[3] - math.Abs(playerStrengths[1]-playerStrengths[3])/3

	return float64(g.ActivePlayer.Team()) * (redYellowStrength - blueGreenStrength)
}

// GetMoveIndexesToSearch appends the indexes of moves worth searching to dst and returns the extended slice.
func (ai *AI) GetMoveIndexesToSearch(g *game.Game, moveEvals []moveScore, depth int, dst []int) []int {
	movesLeft := max(ai.Spread-depth/4*ai.SpreadDrop, 1)
	capturesLeft := movesLeft/2 + 1

	for i, moveEval := range moveEvals {
		if movesLeft == 0 {
			return dst
		}

		if !g.Board.GetPiece(moveEval.move.To).IsEmpty() { // Capture
			if capturesLeft == 0 {
				continue
			}
			capturesLeft--
		}

		movesLeft--
		dst = append(dst, i)
	}

	return dst
}
