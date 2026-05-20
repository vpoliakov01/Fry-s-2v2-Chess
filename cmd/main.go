package main

import (
	"flag"
	"log"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/vpoliakov01/2v2ChessAI/engine/ai"
	"github.com/vpoliakov01/2v2ChessAI/engine/game"
	"github.com/vpoliakov01/2v2ChessAI/engine/play"
)

type flags struct {
	Depth        int
	Moves        int
	HumanPlayers string // Comma separated list of players.
	Evaluation   bool
	Load         string
	ReactUI      bool
	Server       bool
}

var flg flags

func main() {
	// Parse command line flags
	flag.IntVar(&flg.Depth, "depth", 12, "depth of the engine")
	flag.IntVar(&flg.Moves, "moves", 0, "the number of moves to play (0 for unlimited)")
	flag.StringVar(&flg.HumanPlayers, "humans", "0 2", "space separated list of players (0 1 2 3)")
	flag.BoolVar(&flg.Evaluation, "eval", true, "print evalution after every move")
	flag.StringVar(&flg.Load, "load", "", "load pgn notation (no sidelines) to setup the board")
	flag.BoolVar(&flg.ReactUI, "ui", false, "start the React UI")
	flag.BoolVar(&flg.Server, "server", false, "start the server for the UI")
	flag.Parse()

	humanPlayersStr := strings.Fields(flg.HumanPlayers)
	humanPlayers := make([]game.Player, len(humanPlayersStr))
	for i, playerStr := range humanPlayersStr {
		player, err := strconv.Atoi(playerStr)
		if err != nil {
			log.Fatalf("Invalid player number: %v", err)
		}
		humanPlayers[i] = game.Player(player)
	}

	cfg := play.Config{
		Depth:        flg.Depth,
		Spread:       ai.DefaultSpread,
		SpreadDrop:   ai.DefaultSpreadDrop,
		MoveLimit:    flg.Moves,
		HumanPlayers: humanPlayers,
		Evaluation:   flg.Evaluation,
		Load:         flg.Load,
	}

	if flg.ReactUI {
		ex, err := os.Executable()
		if err != nil {
			log.Printf("Failed to get executable path: %v", err)
		} else {
			projectRoot := filepath.Dir(filepath.Dir(ex))
			go func() {
				if err := play.StartReactApp(projectRoot); err != nil {
					log.Printf("Failed to start React app: %v", err)
				}
			}()
		}
	}

	if flg.Server {
		play.NewServer(&cfg).Run()
		return
	}

	play.RunCLI(&cfg)
}
