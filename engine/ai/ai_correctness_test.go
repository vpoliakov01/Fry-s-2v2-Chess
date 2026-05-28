package ai_test

import (
	"fmt"
	"time"

	. "github.com/vpoliakov01/2v2ChessAI/engine/ai"
	"github.com/vpoliakov01/2v2ChessAI/engine/game"
)

func (s *TestSuite) TestGetBestMove() {
	r := s.Require()

	debugCfg := &DebugConfig{
		// Continuation: "f3-g4 c8-d7 f11-e9 l6-n7 h2-c7 b6-c7 e9-d7 n10-m8 g1-k5 d14-f13 h13-f13 n6-j2",
	}
	engine := New(12, DefaultSpread, DefaultSpreadDrop, 0, WithDebugConfig(debugCfg))
	// engine := New(12, DefaultSpread, DefaultSpreadDrop, 0)
	gameFilter := ""
	gameFilter = "Free queen"

	games := s.solvedGames
	if gameFilter != "" {
		games = []*GameTest{}
		for _, g := range s.solvedGames {
			if g.name == gameFilter {
				games = append(games, g)
			}
		}
	}

	for i, g := range games {
		startTime := time.Now()
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

		moveStr := game.HumanReadableMove(g.Board, move, true)
		g.Play(move)
		g.Board.Draw()

		fmt.Println(moveStr)
		fmt.Printf("Name:          %s\n", g.name)
		fmt.Println("Continuation: ", continuation)
		fmt.Printf("Evaluation:    %.2f\n", score)
		fmt.Println("Depth: ", engine.Depth)
		fmt.Println(time.Since(startTime))
		fmt.Println()

		r.Equal(*g.bestMove, move, "Incorrect best move for game %v: %s, expected %s", g.name, move, g.bestMove)
	}
}

func (s *TestSuite) TestBestMoveIndexes() {
	r := s.Require()
	engine := New(12, DefaultSpread, DefaultSpreadDrop, 0, WithDebugConfig(&DebugConfig{}))

	for _, gt := range s.solvedGames {
		g := gt.Copy()

		continuation, _, err := engine.GetBestMove(g.Game)
		r.NoError(err)

		move := continuation[0]
		r.Equal(move, *gt.bestMove, "Incorrect best move for game %v: %s, expected %s", gt.name, move, gt.bestMove)
	}
}

func (s *TestSuite) TestPosition() {
	pieces := [][]int{
		{int(game.NewPiece(0, game.KindKing)), 10, 13},
		{int(game.NewPiece(0, game.KindPawn)), 9, 13},
		{int(game.NewPiece(0, game.KindPawn)), 10, 12},
		{int(game.NewPiece(0, game.KindPawn)), 9, 12},
		{int(game.NewPiece(1, game.KindKing)), 1, 6},
		{int(game.NewPiece(2, game.KindKing)), 6, 12},
		{int(game.NewPiece(3, game.KindKing)), 13, 8},
		{int(game.NewPiece(2, game.KindQueen)), 12, 9},
		{int(game.NewPiece(0, game.KindQueen)), 13, 10},
	}

	g := game.NewGame()
	g.Board.Clear()

	for i := range pieces {
		piece := game.Piece(pieces[i][0])
		file := pieces[i][1]
		rank := pieces[i][2]

		g.Board.PlacePiece(piece, game.Square{File: file, Rank: rank})
	}

	engine := New(2, DefaultSpread, DefaultSpreadDrop, 0)
	g.Board.Draw()

	for i := 0; i < 30; i++ {
		continuation, _, err := engine.GetBestMove(g)
		if err != nil {
			if err == ErrGameEnded {
				fmt.Printf("%v: Team %v won!\n", i, g.Winner)
			} else {
				fmt.Println(err)
			}
			break
		}
		move := continuation[0]

		fmt.Println(move)

		if !g.Board.IsEmpty(move.To) {
			capturedPiece := game.Piece(g.Board.GetPiece(move.To))
			opponent := capturedPiece.Player()
			piece := game.Piece(g.Board.GetPiece(move.From))
			player := piece.Player()
			fmt.Printf("%v: P%v's %v takes P%v's %v after %v\n", i, player, piece, opponent, capturedPiece, move)
		}

		g.Play(move)
		g.Board.Draw()
	}
}
