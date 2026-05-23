package ai_test

import (
	"fmt"
	"runtime"
	"time"

	. "github.com/vpoliakov01/2v2ChessAI/engine/ai"
	"github.com/vpoliakov01/2v2ChessAI/engine/game"
)

func (s *TestSuite) TestConsecutiveMoves() {
	engine := New(16, DefaultSpread, DefaultSpreadDrop, 0, WithEnableDebug(true))
	g := s.GetGame("4 queens in the middle, bishops ready").Copy()
	moves := 10

	startTime := time.Now()
	for i := 0; i < moves; i++ {
		continuation, score, err := engine.GetBestMove(g.Game)
		if err != nil {
			if err == ErrGameEnded {
				fmt.Printf("%v: Team %v won!\n", i, g.Winner)
			} else {
				fmt.Println(err)
			}
			break
		}
		move := continuation[0]

		piece := game.Piece(g.Board.GetPiece(move.From))
		if !g.Board.IsEmpty(move.To) {
			capturedPiece := game.Piece(g.Board.GetPiece(move.To))
			fmt.Printf("%v: %v takes %v after %v\n", i, piece, capturedPiece, move)
		} else {
			fmt.Printf("%v: %v moves %v\n", i, piece, move)
		}

		g.Play(move)
		g.Board.Draw()
		fmt.Printf("Evaluation:    %.2f\n", score)
		fmt.Println("Continuation: ", continuation)
	}

	fmt.Println("Depth: ", engine.Depth)
	fmt.Println(time.Since(startTime))
}

func (s *TestSuite) TestEngineDepthsPerformance() {
	r := s.Require()

	moves := 1
	cpus := runtime.NumCPU()
	runtime.GOMAXPROCS(cpus)

	depths := []int{
		// 2,
		// 3,
		// 4,
		// 5,
		// 6,
		// 7,
		// 8,
		// 9,
		// 10,
		// 11,
		12,
		// 13,
		// 14,
		// 15,
		// 16,
	}

	games := append(s.openGames, s.solvedGames...)

	last := time.Duration(0)
	totalStart := time.Now()
	for _, testGame := range games {
		g := testGame.Game.Copy()

		for _, d := range depths {
			start := time.Now()
			engine := New(d, DefaultSpread, DefaultSpreadDrop, 0, WithEnableDebug(true))

			continuations := [][]game.Move{}
			scores := []float64{}

			for i := 0; i < moves; i++ {
				continuation, score, err := engine.GetBestMove(g)
				if err != nil {
					fmt.Println(err)
					break
				}
				continuations = append(continuations, continuation)
				scores = append(scores, score)
			}

			t := time.Since(start)
			if last == time.Duration(0) {
				last = t
			}

			testGame.Print(scores[0], continuations[0])

			totalPossibleEvals := engine.TotalPossibleEvals()
			fmt.Printf(
				"Depth: %v   Spread: %v/%v   t: %.2fs   t/m: %.2fs   r: %.2fx   e: %v   p: %.3f%%   t/e: %.2fµs\n",
				engine.Depth,
				engine.Spread,
				engine.SpreadDrop,
				t.Seconds(),
				t.Seconds()/float64(moves),
				float64(t)/float64(last),
				engine.EvalsCount,
				(1-(float64(engine.EvalsCount)/float64(totalPossibleEvals)))*100,
				float64(t.Microseconds())/float64(engine.EvalsCount),
			)
			last = t

			engine.PrintBestMoveIndexes(false, true)

			if testGame.bestMove != nil {
				r.Equal(testGame.bestMove.String(), continuations[0][0].String())
			}
			if testGame.score != nil && len(scores) > 0 {
				r.Equal(*testGame.score, scores[0])
			}
		}
	}

	fmt.Printf("Total time: %.2fs\n", time.Since(totalStart).Seconds())
}
