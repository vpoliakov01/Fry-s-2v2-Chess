package ai_test

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/suite"

	. "github.com/vpoliakov01/2v2ChessAI/engine/ai"
	"github.com/vpoliakov01/2v2ChessAI/engine/game"
)

type TestSuite struct {
	suite.Suite
	engine      *AI
	solvedGames []*GameTest
	openGames   []*GameTest
}

type GameTest struct {
	*game.Game
	name     string
	bestMove *game.Move
	score    *float64
}

type test struct {
	name     string
	bestMove string
	score    float64
	pgn      string
}

func Test(t *testing.T) {
	suite.Run(t, new(TestSuite))
}

func (s *TestSuite) SetupTest() {
	s.engine = New(DefaultDepth, DefaultSpread, DefaultSpreadDrop, 0, WithEnableDebug(true))

	games := []test{
		{
			name:     "Mate in 1 (g1-a7)",
			bestMove: "g1-a7",
			score:    999,
			pgn: `
1. f2-f3 b6-c6 g13-g12 m8-l8`,
		},
		{
			name:     "Mate in 1 (m9-n8)",
			bestMove: "m9-n8",
			score:    999,
			pgn: `
1. h2-h3 b8-c8 i13-i12 m8-l8
2. g1-m7 a9-i1 h14-m9 n7-m7
3. e1-f3 i1-j2`,
		},
		{
			name:     "Free queen (a7-b6)",
			bestMove: "a7-b6",
			pgn: `
1. f2-f3 b7-c7 d13-d12 m7-l7
2. g1-b6`,
		},
		{
			name:     "Mate in 7 (g1-m7)",
			bestMove: "g1-m7",
			pgn: `
1. h2-h3 b8-c8 i13-i12 m8-l8`,
		},
		{
			name:     "3 queens, mate in 6 (j4-m7)",
			bestMove: "j4-m7",
			pgn: `
1. h2-h3 b9-c9 i13-i12 m8-l8
2. g1-j4 a8-d11 e13-e12 m5-l5
3. e2-e3 d11-a8 h14-k11 n7-l9`,
		},
		{
			name:     "Queen trap",
			bestMove: "l6-k5",
			pgn: `
1. h2-h3 b7-d7 g13-g12 m8-k8
2. g1-k5 a8-g2 h14-g13 n9-h3
3. f1-g2 b6-c6 j13-j12 h3-i2
4. h1-i2 a6-g12 f13-g12 m6-l6
5. j2-j3 b10-c10 g13-k9`,
		},
		{
			name: "4 queens in the middle, bishops ready",
			pgn: `
1. h2-h3 b9-c9 i13-i12 m8-l8
2. g1-j4 a8-d11 e13-e12 m5-l5
3. e2-e3 d11-g8 h14-k11 n7-l9`,
		},
		{
			name: "6/10 engine game",
			pgn: `
1. j2-j3 b5-c5 j14-i12 n5-l6
2. e2-e3 a6-f1 e13-e12 m7-k7
3. g1-f1 a5-c4 j13-j12 n10-l9
4. f1-c4 b7-c7 h13-h12 m5-l5`,
		},
		{
			name: "Complex real",
			pgn: `
1. k2-k4 b7-d7 i13-i12 m6-k6
2. f2-f4 a8-b7 g13-g12 m8-l8
3. e1-f3 a10-c9 e14-f12 m10-l10
4. g2-g4 b11-d11 k13-k12 m7-l7`,
		},
		{
			name: "Wild midgame",
			pgn: `
1. h2-h3 b7-c7 i13-i12 m10-l10
2. g1-j4 a8-d5 h14-i13 n9-m10
3. j2-j3 b10-c10 i13-h12 m8-l8
4. j4-j8 a9-b10 j14-k12 n5-l6
5. k2-k4 b4-d4 e14-d12 n7-k10
6. f2-f3 a6-b7`,
		},
	}

	for _, tg := range games {
		g, err := game.LoadPGN(tg.pgn)
		s.Require().NoError(err)

		gt := &GameTest{Game: g.Game, name: tg.name}
		if tg.score != 0 {
			gt.score = &tg.score
		}

		if tg.bestMove != "" {
			bestMove := game.MoveFromPGN(tg.bestMove)
			gt.bestMove = &bestMove
			s.solvedGames = append(s.solvedGames, gt)
		} else {
			s.openGames = append(s.openGames, gt)
		}
	}
}

func (gt *GameTest) Copy() *GameTest {
	return &GameTest{
		Game:     gt.Game.Copy(),
		name:     gt.name,
		bestMove: gt.bestMove,
		score:    gt.score,
	}
}

func (s *TestSuite) GetGame(name string) *GameTest {
	for _, gt := range append(s.solvedGames, s.openGames...) {
		if gt.name == name {
			return gt.Copy()
		}
	}
	return nil
}

func (gt *GameTest) Print(score float64, continuation []game.Move) {
	fmt.Println(gt.name)
	fmt.Printf("Continuation: %v %.2f\n", continuation, score)
}
