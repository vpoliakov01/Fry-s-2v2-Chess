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
	cache   *TranspositionTable

	bfsDepth int // Depth of the running BFS.

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
	if ai.cache == nil {
		ai.cache = NewTranspositionTable()
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
		childEval := -ms.score

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
		ai.cache.Set(g.Hash, bestMove, bestScore, remainingDepth, boundOf(bestScore, alphaOrig, beta))
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
			ScoreDelta: bestScore - moveEvals[0].score,
		}, cpu)
	}

	return bestScore
}

// EvaluateCurrent returns the difference between strengths of the team making the move and the opponent team.
// Used to seed the absolute eval at the search root; per-move updates use EvaluateMove for an incremental delta.
func (ai *AI) EvaluateCurrent(g *game.Game, buffer *buffer) float64 {
	buffer.evalsCount++
	playerStrengths := [4]float64{}

	if g.HasEnded() {
		return float64(g.ActivePlayer.Team()*g.Winner) * 1000
	}

	// For each piece, run piece strength evaluation.
	for player := game.Player(0); player < 4; player++ {
		for _, square := range g.Board.PieceSquares[player] {
			piece := g.Board.GetPiece(square)
			playerStrengths[player] += piece.GetStrength(g.Board, square, player)
		}
	}

	redYellowStrength := playerStrengths[0] + playerStrengths[2] - math.Abs(playerStrengths[0]-playerStrengths[2])/3
	blueGreenStrength := playerStrengths[1] + playerStrengths[3] - math.Abs(playerStrengths[1]-playerStrengths[3])/3

	return float64(g.ActivePlayer.Team()) * (redYellowStrength - blueGreenStrength)
}

// EvaluateMove returns the move's score from the moving player's perspective.
func (ai *AI) EvaluateMove(g *game.Game, buffer *buffer, move game.Move) float64 {
	buffer.evalsCount++

	piece := g.Board.GetPiece(move.From)
	player := piece.Player()

	capturedPiece := g.Board.GetPiece(move.To)
	captureValue := 0.0

	if !capturedPiece.IsEmpty() {
		if capturedPiece.Kind() == game.KindKing {
			return mateValue
		}
		captureValue = capturedPiece.GetStrength(g.Board, move.To, capturedPiece.Player())
	}

	newStrength := piece.GetStrength(g.Board, move.To, player)
	oldStrength := piece.GetStrength(g.Board, move.From, player)

	return (newStrength - oldStrength) + captureValue
}

// GetMovesToSearch returns the moves worth searching, each tagged with its post-move absolute eval.
// TODO: This function should ideally pick the most promising moves based on more rules
// (capture value diff, piece position, king safety, etc.)
func (ai *AI) GetMovesToSearch(g *game.Game, buffer *buffer, depth int, eval float64, firstMove game.Move) []moveScore {
	moves := g.GetMoves(buffer.moves[depth][:0])
	buffer.moves[depth] = moves // In case moves got reallocated by append inside GetMoves.

	moveEvals := buffer.moveEvals[depth][:len(moves)]
	for i, move := range moves {
		moveEvals[i] = moveScore{move, eval + ai.EvaluateMove(g, buffer, move)}
	}
	buffer.moveEvals[depth] = moveEvals

	movesLeft := len(moveEvals)
	if depth > 1 {
		movesLeft = max(ai.Spread-depth/4*ai.SpreadDrop, 1)
	}
	capturesLeft := movesLeft/2 + 1

	sortMoveEvals(moveEvals)

	movesToSearch := buffer.movesToSearch[depth][:0]

	if firstMove != game.NullMove {
		for _, moveEval := range moveEvals {
			if moveEval.move == firstMove {
				movesToSearch = append(movesToSearch, moveEval)

				if !g.Board.GetPiece(firstMove.To).IsEmpty() { // Capture
					capturesLeft--
				}

				movesLeft--
				break
			}
		}
	}

	for _, moveEval := range moveEvals {
		if movesLeft == 0 {
			break
		}

		if moveEval.move == firstMove {
			continue
		}

		if !g.Board.GetPiece(moveEval.move.To).IsEmpty() { // Capture
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
