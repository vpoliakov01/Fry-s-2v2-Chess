package ai_test

import (
	"fmt"
	"time"

	. "github.com/vpoliakov01/2v2ChessAI/engine/ai"
	"github.com/vpoliakov01/2v2ChessAI/engine/game"
)

func (s *TestSuite) TestGetBestMove() {
	r := s.Require()

	engine := New(12, 8, DefaultSpreadDrop, 0, WithEnableDebug(true))
	gameFilter := ""
	// gameFilter = "3 queens, mate in 6 (j4-m7)"

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

		moveStr := game.HumanReadableMove(g.Board, move)
		g.Board.Draw()
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
	engine := New(12, DefaultSpread, DefaultSpreadDrop, 0, WithEnableDebug(true))

	for _, gt := range s.solvedGames {
		g := gt.Copy()

		_, _, err := engine.GetBestMove(g.Game)
		s.Require().NoError(err)

		engine.PrintBestMoveIndexes(false, true)
	}
}

func (s *TestSuite) TestPosition() {
	pieces := [][]int{
		{int(game.NewPiece(0, game.KindKing)), 13, 10},
		{int(game.NewPiece(0, game.KindPawn)), 13, 9},
		{int(game.NewPiece(0, game.KindPawn)), 12, 10},
		{int(game.NewPiece(0, game.KindPawn)), 12, 9},
		{int(game.NewPiece(1, game.KindKing)), 6, 1},
		{int(game.NewPiece(2, game.KindKing)), 12, 6},
		{int(game.NewPiece(3, game.KindKing)), 8, 13},
		{int(game.NewPiece(2, game.KindQueen)), 9, 12},
		{int(game.NewPiece(0, game.KindQueen)), 10, 13},
	}

	g := game.NewGame()
	g.Board.Clear()

	for i := range pieces {
		piece := game.Piece(pieces[i][0])
		rank := pieces[i][1]
		file := pieces[i][2]

		g.Board.PlacePiece(piece, game.Square{rank, file})
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
