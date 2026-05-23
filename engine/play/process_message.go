package play

import (
	"context"
	"fmt"
	"log"
	"math"
	"runtime/debug"
	"slices"
	"time"

	"github.com/vpoliakov01/2v2ChessAI/engine/ai"
	g "github.com/vpoliakov01/2v2ChessAI/engine/game"
)

func (c *Connection) ProcessMessage(msg *Message) {
	log.Printf("Processing message: %v", msg)
	switch msg.Type {
	case MessageTypeSetSettings:
		cfg, err := CastData[Config](msg.Data)
		if err != nil {
			log.Printf("Error casting settings: %v", err)
			return
		}
		c.processSetSettings(cfg)
	case MessageTypeGetAvailableMoves:
		c.processGetAvailableMoves()
	case MessageTypePlayerMove:
		move, err := CastData[PGNMove](msg.Data)
		if err != nil {
			log.Printf("Error casting move: %v", err)
			return
		}
		c.processPlayerMove(move)
	case MessageTypeSaveGame:
		c.processSaveGame()
	case MessageTypeLoadGame:
		c.processLoadGame(msg.Data.(string))
	case MessageTypeNewGame:
		c.processNewGame()
	case MessageTypeSetCurrentMove:
		c.processSetCurrentMove(int(msg.Data.(float64)))
	default:
		log.Printf("Unknown message type: %s", msg.Type)
	}
}

func (c *Connection) processSetSettings(cfg Config) {
	wasEngineActive := !slices.Contains(c.cfg.HumanPlayers, c.gs.ActivePlayer)
	willBeEngineActive := !slices.Contains(cfg.HumanPlayers, c.gs.ActivePlayer)

	if wasEngineActive && !willBeEngineActive {
		c.stopPlayingEngineMovesIfRunning(true)
	}

	c.cfg = &cfg
	c.engine.Depth = cfg.Depth
	c.engine.Spread = cfg.Spread
	c.engine.SpreadDrop = cfg.SpreadDrop

	if cfg.EvalLimit == 0 {
		c.engine.EvalLimit = ai.MaxEvalLimit
	} else {
		c.engine.EvalLimit = cfg.EvalLimit
	}

	if !wasEngineActive && willBeEngineActive {
		c.playUntilPlayerMove()
	} else {
		c.processGetAvailableMoves()
	}
}

func (c *Connection) processGetAvailableMoves() {
	gameMoves := c.gs.GetMoves(nil)
	moves := make([]PGNMove, len(gameMoves))
	for i, gameMove := range gameMoves {
		moves[i] = PGNMoveFromGameMove(gameMove)
	}
	c.SendMessage(MessageTypeAvailableMoves, moves)
}

func (c *Connection) processPlayerMove(move PGNMove) {
	c.stopPlayingEngineMovesIfRunning(true)
	game := c.gs

	gameMove := GameMoveFromPGN(move)
	if err := game.ValidateMove(&gameMove); err != nil {
		c.SendMessage(MessageTypeInvalidMove, err.Error())
		return
	}

	game.Play(gameMove)
	game.Board.Draw()

	c.playUntilPlayerMove()
}

func (c *Connection) processSaveGame() {
	c.SendMessage(MessageTypeSaveGameResponse, SaveGameResponse{
		PGN: c.gs.PGN(),
	})
}

func (c *Connection) processLoadGame(data string) {
	c.stopPlayingEngineMovesIfRunning(true)

	game, err := g.LoadPGN(data)
	if err != nil {
		log.Printf("Error loading game: %v", err)
		return
	}
	c.gs = game
	c.engine.ResetCache()

	c.SendMessage(MessageTypeLoadGameResponse, LoadGameResponse{
		PastMoves:   PGNMovesFromGameMoves(c.gs.PastMoves),
		CurrentMove: c.gs.CurrentMove,
	})
	c.playUntilPlayerMove()
}

func (c *Connection) processNewGame() {
	c.stopPlayingEngineMovesIfRunning(true)

	c.gs = g.NewGameSession()
	c.engine.ResetCache()
	c.SendMessage(MessageTypeLoadGameResponse, LoadGameResponse{
		PastMoves:   PGNMovesFromGameMoves(c.gs.PastMoves),
		CurrentMove: c.gs.CurrentMove,
	})
	c.playUntilPlayerMove()
}

func (c *Connection) processSetCurrentMove(moveIndex int) {
	c.stopPlayingEngineMovesIfRunning(true)

	err := c.gs.SetCurrentMove(moveIndex)
	if err != nil {
		log.Printf("Error setting current move: %v", err)
		return
	}
	c.processGetAvailableMoves()
}

// playUntilPlayerMove proceeds until the active player is a human player.
func (c *Connection) playUntilPlayerMove() {
	snapshot := c.gs.Copy()

	ctx, cancel := context.WithCancel(context.Background())
	c.engineMutex.Lock()
	c.engineCancel = cancel
	c.engineMutex.Unlock()

	go func() {
		defer cancel()
		defer func() {
			if r := recover(); r != nil {
				log.Printf("panic in playUntilPlayerMove goroutine: %v\n%s", r, debug.Stack())
			}
		}()
		c.playEngineMoves(ctx, snapshot)

		if ctx.Err() != nil {
			return
		}

		if c.gs.HasEnded() {
			winningTeam := c.gs.Winner
			losingKing := g.Player(0)
			for player := g.Player(0); player < 4; player++ {
				if !c.gs.HasKing(player) {
					losingKing = player
					break
				}
			}

			c.SendMessage(MessageTypeGameEnded, GameEndedResponse{
				King:   losingKing.String(),
				Winner: winningTeam.String(),
			})
			return
		}

		c.processGetAvailableMoves()
	}()
}

// stopPlayingEngineMovesIfRunning cancels the currently running engine-moves goroutine (if any).
func (c *Connection) stopPlayingEngineMovesIfRunning(condition bool) {
	if !condition {
		return
	}

	c.engineMutex.Lock()
	cancel := c.engineCancel
	c.engineCancel = nil
	c.engineMutex.Unlock()

	if cancel != nil {
		cancel()
	}
	c.engine.Stop()
}

// playEngineMoves plays engine moves until the active player is a human player
// or the context is cancelled.
func (c *Connection) playEngineMoves(ctx context.Context, game *g.GameSession) {
	for !slices.Contains(c.cfg.HumanPlayers, game.ActivePlayer) {
		if ctx.Err() != nil {
			c.SendMessage(MessageTypeStoppedProcessing, nil)
			return
		}

		c.SendMessage(MessageTypeProcessing, nil)
		now := time.Now()
		moveNumber := game.MoveNumber
		continuation, score, err := c.engine.GetBestMove(game.Game)
		if err != nil {
			log.Printf("Error getting best move: %v", err)
			c.SendMessage(MessageTypeStoppedProcessing, nil)
			return
		}

		if ctx.Err() != nil {
			c.SendMessage(MessageTypeStoppedProcessing, nil)
			return
		}

		bestMove := continuation[0]
		elapsed := time.Since(now)

		c.SendMessage(MessageTypeEngineMove, BestMoveResponse{
			Continuation: PGNMovesFromGameMoves(continuation),
			MoveNumber:   moveNumber,
			Score:        math.Round(score*float64(game.ActivePlayer.Team())*100) / 100,
			Time:         math.Round(elapsed.Seconds()*100) / 100,
			Evaluations:  c.engine.EvalsCount,
		})

		game.Play(bestMove)
		c.gs = game.Copy()

		fmt.Println("Move number:", game.MoveNumber)
		fmt.Println("Active player:", game.ActivePlayer)
		fmt.Println("Score:", score)
		fmt.Println("Time:", elapsed)
		fmt.Println("Evaluations:", c.engine.EvalsCount)
		game.Board.Draw()
	}
}
