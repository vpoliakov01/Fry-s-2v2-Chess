package play

import (
	"bufio"
	"fmt"
	"os"
	"slices"
	"strings"
	"time"

	"github.com/vpoliakov01/2v2ChessAI/engine/ai"
	"github.com/vpoliakov01/2v2ChessAI/engine/game"
)

func RunCLI(cfg *Config) {
	fmt.Printf("\nDepth: %v\nMoves limit: %v\nHuman players: %v\nEvaluation: %v\nLoad: %v\n\n", cfg.Depth, cfg.MoveLimit, cfg.HumanPlayers, cfg.Evaluation, cfg.Load)

	engine := ai.New(cfg.Depth, cfg.Spread, cfg.SpreadDrop, cfg.EvalLimit, ai.WithEnableStoredCache(true))
	startTime := time.Now()

	g := game.SetupBoard(cfg.Load)
	g.Board.Draw()

	// Play the game.
	for i := 0; !g.HasEnded(); i++ {
		if cfg.MoveLimit > 0 && i >= cfg.MoveLimit {
			break
		}

		fmt.Println()

		var move *game.Move
		var score float64
		moveStartTime := time.Now()

		if slices.Contains(cfg.HumanPlayers, g.ActivePlayer) { // Human player's turn.
			for {
				in, err := ReadInput()
				if err != nil {
					fmt.Println(err)
					continue
				}

				switch {
				case strings.ToLower(in) == "save":
					file, err := game.SaveToFile(g)
					if err != nil {
						fmt.Println(err)
						continue
					}
					fmt.Printf("Saved to %v\n", file)
					continue
				case strings.ToLower(in) == "exit":
					os.Exit(0)
				default:
					move, err = game.ParseMove(in)
				}

				if err != nil {
					fmt.Println(err)
					continue
				}

				if err := g.ValidateMove(move); err != nil {
					fmt.Println(err)
					continue
				}
				break
			}
		} else { // AI's turn.
			continuation, s, err := engine.GetBestMove(g.Game)
			if err != nil {
				if err == ai.ErrGameEnded {
					fmt.Printf("%v: Team %v won!\n", i, g.Winner)
				} else {
					fmt.Println(err)
				}
				break
			}
			move = &continuation[0]
			score = s

			if cfg.Evaluation {
				fmt.Printf("Evaluation: %.3f\n", score*float64(g.ActivePlayer.Team()))
				fmt.Printf("Continuation: %v\n", continuation)
			}
		}

		fmt.Printf("Time used: %.3fs\n", time.Since(moveStartTime).Seconds())
		piece := game.Piece(g.Board.GetPiece(move.From))

		if !g.Board.IsEmpty(move.To) {
			capturedPiece := game.Piece(g.Board.GetPiece(move.To))
			fmt.Printf("%v: %v takes %v after %v\n", i, piece, capturedPiece, move)
		} else {
			fmt.Printf("%v: %v moves %v\n", i, piece, move)
		}

		g.Play(*move)
		g.Board.Draw()
	}

	if g.Winner != 0 {
		fmt.Printf("Team %v won!\n", g.Winner)
	}
	fmt.Printf("Total time: %v\n", time.Since(startTime))
}

// ReadInput reads user io from STDIN.
func ReadInput() (string, error) {
	reader := bufio.NewReader(os.Stdin)
	fmt.Print("Enter a command (save / exit) or a move in format e2e4: ")

	in, err := reader.ReadString('\n')
	if err != nil {
		return "", err
	}

	return strings.Trim(in, " \n\t"), nil
}
