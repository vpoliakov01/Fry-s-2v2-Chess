package ai_test

import (
	"fmt"
	"time"

	. "github.com/vpoliakov01/2v2ChessAI/engine/ai"
	"github.com/vpoliakov01/2v2ChessAI/engine/game"
)

func (s *TestSuite) TestCache() {
	engine := New(12, DefaultSpread, DefaultSpreadDrop, 0, WithEnableDebug(true))
	moves := 5

	conts := make([][]game.Move, moves)
	times := make([]time.Duration, moves)
	scores := make([]float64, moves)
	timeSpent := time.Duration(0)

	for i := 0; i < 3; i++ {
		g := s.GetGame("4 queens in the middle, bishops ready").Copy()
		startTime := time.Now()

		if i == 0 {
			engine.ResetCache()
		}

		for m := 0; m < moves; m++ {
			iterStartTime := time.Now()

			continuation, score, err := engine.GetBestMove(g.Game)
			if err != nil {
				if err == ErrGameEnded {
					fmt.Printf("%v: Team %v won!\n", m, g.Winner)
				} else {
					fmt.Println(err)
				}
				break
			}
			move := continuation[0]

			if i == 0 {
				conts[m] = continuation
				scores[m] = score
				times[m] = time.Since(iterStartTime)
			} else {
				fmt.Printf("%v: %2.2fs e:%2.2f %v\n", m, times[m].Seconds(), scores[m], conts[m])
				fmt.Printf("   %2.2fs e:%2.2f %v\n", time.Since(iterStartTime).Seconds(), score, continuation)
			}

			g.Play(move)
		}

		if i == 0 {
			timeSpent = time.Since(startTime)
		} else {
			fmt.Printf("Time spent 0: %2.2f\n", timeSpent.Seconds())
			fmt.Printf("Time spent 1: %2.2f\n\n", time.Since(startTime).Seconds())
		}
	}
}
