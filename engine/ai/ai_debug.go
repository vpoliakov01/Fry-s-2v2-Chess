package ai

import (
	"fmt"
	"sort"
	"strings"
	"sync"

	"github.com/gofiber/fiber/v2/log"
	"github.com/vpoliakov01/2v2ChessAI/engine/game"
)

type DebugConfig struct {
	// Continuation is a space-separated list of PGN moves (e.g. "j2-j3 a7-a8 ...")
	// Move ordering stats are printed along this path
	Continuation string

	continuation     []continuationEntry
	contMovesPrinted int
	mu               sync.Mutex

	lastResults  []candidateResult // Captured snapshot of the latest root search.
	lastBfsDepth int
}

type continuationEntry struct {
	hash     uint64
	depth    int
	nextMove game.Move
}

// bestmoveDataAvgAcc accumulates analytics on the indexes of best moves at each depth, for debugging and tuning move ordering heuristics.
type bestmoveDataAvgAcc struct {
	IndexSum   int
	MaxIndex   int
	TotalMoves int
	ScoreDelta float64
	Count      int
}

func (ai *AI) InitDebug(g *game.Game) {
	if ai.debugConfig == nil {
		return
	}

	ai.BestMoves = make([][]bestmoveDataAvgAcc, cpus)
	for i := range ai.BestMoves {
		ai.BestMoves[i] = make([]bestmoveDataAvgAcc, ai.Depth+1)
	}

	cfg := ai.debugConfig

	// Parse moves
	tokens := strings.Fields(cfg.Continuation)
	moves := make([]game.Move, len(tokens))
	for i, tok := range tokens {
		moves[i] = game.MoveFromPGN(tok)
	}

	// Populate pendingPrefixes
	gameCopy := g.Copy()
	entries := []continuationEntry{}
	rootNext := game.NullMove
	if len(moves) > 0 {
		rootNext = moves[0]
	}
	entries = append(entries, continuationEntry{hash: g.Hash, depth: 1, nextMove: rootNext})

	fmt.Println()

	for i, move := range moves {
		err := gameCopy.ValidateMove(&move)
		if err != nil {
			log.Errorf("Invalid move: %v", err)
			break
		}

		fmt.Printf("%s|", game.HumanReadableMove(gameCopy.Board, move, true))
		if gameCopy.Board.GetPiece(move.From).Player() == 3 || i == len(moves)-1 {
			fmt.Println()
		}

		gameCopy.Play(move)

		next := game.NullMove
		if i+1 < len(moves) {
			next = moves[i+1]
		}

		entries = append(entries, continuationEntry{hash: gameCopy.Hash, depth: i + 2, nextMove: next})
	}

	cfg.continuation = entries
}

func (ai *AI) TotalPossibleEvals() int {
	total := 1
	for depth := 1; depth <= ai.Depth; depth++ {
		total *= ai.Spread - depth/4*ai.SpreadDrop
	}
	return total
}

// popContinuationEntry atomically pops the next pending prefix if its hash matches.
func (c *DebugConfig) popContinuationEntry(hash uint64) (continuationEntry, bool) {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.contMovesPrinted == len(c.continuation) {
		return continuationEntry{}, false
	}

	next := c.continuation[c.contMovesPrinted]
	if next.hash != hash {
		return continuationEntry{}, false
	}

	c.contMovesPrinted++
	return next, true
}

func (ai *AI) PrintMoveOrderingStats(g *game.Game, moveEvals []moveScore, eval float64, movesLeft int, firstMove game.Move) {
	c := ai.debugConfig

	entry, ok := c.popContinuationEntry(g.Hash)
	if !ok {
		return
	}

	firstMoveStr := ""
	if firstMove != game.NullMove {
		firstMoveStr = fmt.Sprintf(", firstMove: %s", game.HumanReadableMove(g.Board, firstMove, false))
	}

	fmt.Printf("\nMove ordering (Depth %d, eval: %.2f%s):\n", entry.depth, eval, firstMoveStr)
	for i, moveEval := range moveEvals {
		if i == movesLeft {
			fmt.Println("-----------------------------------------")
		}

		marker := "  "
		if entry.nextMove != game.NullMove && moveEval.move == entry.nextMove {
			marker = "->"
		}

		if marker != "->" && i >= movesLeft {
			continue
		}

		fmt.Printf("%2d:%s %s m:%7.2f\tp:%7.2f\n", i, marker, game.HumanReadableMove(g.Board, moveEval.move, true), moveEval.score, moveEval.posEval)
	}

	if c.contMovesPrinted%4 == 0 {
		gameCopy := g.Copy()
		gameCopy.Play(entry.nextMove)
		gameCopy.Board.Draw()
		fmt.Println()
	}
}

func (ai *AI) PrintSearchResults(g *game.Game) {
	results := ai.debugConfig.lastResults
	bfsDepth := ai.debugConfig.lastBfsDepth
	c := ai.debugConfig

	if len(results) == 0 {
		return
	}

	sort.Slice(results, func(i, j int) bool {
		return results[i].score > results[j].score
	})

	fmt.Printf("\nsearchRootMovesParallel results (Depth:%v, eval: %.2f):\n", bfsDepth, c.lastResults[0].score)
	for i, result := range results {
		move := result.continuation[0]
		marker := "  "
		if move == c.continuation[0].nextMove {
			marker = "->"
		}

		fmt.Printf("%2d:%s %s%7.2f\n", i, marker, game.HumanReadableMove(g.Board, move, true), result.score)
	}
}

// captureSearchResults stores max-depth results from searchRootMovesParallel.
func (c *DebugConfig) captureSearchResults(results []candidateResult, bfsDepth int) {
	snapshot := make([]candidateResult, len(results))
	for i, r := range results {
		contCopy := make([]game.Move, len(r.continuation))
		copy(contCopy, r.continuation)
		snapshot[i] = candidateResult{score: r.score, continuation: contCopy}
	}

	c.lastResults = snapshot
	c.lastBfsDepth = bfsDepth
}

// recordBestMove locates bestMove in the sorted move list and accumulates
// per-depth move-ordering analytics. No-op when debug is disabled.
func (ai *AI) recordBestMove(buffer *buffer, depth int, bestMove game.Move, bestScore float64, cpu int) {
	if ai.debugConfig == nil || bestMove == game.NullMove {
		return
	}

	moveEvals := buffer.moveEvals[depth]
	bestMoveIndex := 0
	for i := range moveEvals {
		if moveEvals[i].move == bestMove {
			bestMoveIndex = i
			break
		}
	}

	acc := &ai.BestMoves[cpu][depth]
	acc.Count++
	acc.IndexSum += bestMoveIndex
	acc.MaxIndex = max(acc.MaxIndex, bestMoveIndex)
	acc.TotalMoves += len(moveEvals)
	acc.ScoreDelta += bestScore - moveEvals[0].posEval
}

func (ai *AI) PrintBestMoveIndexes(printIndividualCPUs bool, printAllCPUs bool) {
	fmt.Println("        dep  best max  moves    ratio    Δscore    total")

	cpuAcc := make([][]bestmoveDataAvgAcc, cpus)
	for i := range cpuAcc {
		cpuAcc[i] = make([]bestmoveDataAvgAcc, ai.Depth+1)
	}

	for cpu := range ai.BestMoves {
		hasData := false
		for depth := range ai.BestMoves[cpu] {
			if ai.BestMoves[cpu][depth].Count > 0 {
				hasData = true
				break
			}
		}
		if !hasData {
			continue
		}

		if printIndividualCPUs {
			fmt.Printf("CPU %v:\n", cpu)
		}

		for depth := range ai.BestMoves[cpu] {
			if depth == 0 {
				continue
			}

			acc := ai.BestMoves[cpu][depth]
			if acc.Count == 0 {
				continue
			}

			avgIndex := float64(acc.IndexSum) / float64(acc.Count)
			maxIndex := acc.MaxIndex
			moves := float64(acc.TotalMoves) / float64(acc.Count)
			scoreDelta := acc.ScoreDelta / float64(acc.Count)

			sharedAcc := &cpuAcc[cpu][depth]
			sharedAcc.Count += acc.Count
			sharedAcc.IndexSum += acc.IndexSum
			sharedAcc.MaxIndex = max(sharedAcc.MaxIndex, acc.MaxIndex)
			sharedAcc.TotalMoves += acc.TotalMoves
			sharedAcc.ScoreDelta += acc.ScoreDelta

			if printIndividualCPUs {
				fmt.Printf(
					"\t %2v:%5.2f (%2d) /%4.0f  = %4.1f%%   %7.2f  %7v\n",
					depth,
					avgIndex+1, // Human index
					maxIndex+1,
					moves,
					avgIndex/moves*100,
					scoreDelta,
					acc.Count,
				)

				if depth == ai.Depth {
					fmt.Println("\t -----------------------------------------------")
				}
			}
		}
	}

	if printAllCPUs {
		fmt.Println("All CPUs")
		for depth := 1; depth <= ai.Depth; depth++ {
			acc := &cpuAcc[0][depth]

			avgIndex := float64(acc.IndexSum) / float64(acc.Count)
			maxIndex := acc.MaxIndex
			moves := float64(acc.TotalMoves) / float64(acc.Count)
			scoreDelta := acc.ScoreDelta / float64(acc.Count)

			fmt.Printf(
				"\t %2v:%5.2f (%2d) /%4.0f  = %4.1f%%   %7.2f  %7v\n",
				depth,
				avgIndex+1, // Human index
				maxIndex+1,
				moves,
				avgIndex/moves*100,
				scoreDelta,
				acc.Count,
			)
			if depth == ai.Depth {
				fmt.Println("\t -----------------------------------------------")
			}
		}
	}

	fmt.Println()
}
