package play_test

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/vpoliakov01/2v2ChessAI/engine/game"
	"github.com/vpoliakov01/2v2ChessAI/engine/play"
)

// validFirstMove is Red's pawn h2-h3, the same opening used in
// engine/play/io_test.go. Player 0 (Red) is the first to move on a fresh game.
const validFirstMove play.PGNMove = "h2-h3"

func TestProcessGetAvailableMoves(t *testing.T) {
	conn := NewConnection(t, nil)

	conn.ProcessMessage(play.MessageTypeGetAvailableMoves, nil)

	moves := availableMovesFromMessage(t, requireSingleMessage(t, conn, play.MessageTypeAvailableMoves))
	require.NotEmpty(t, moves, "expected at least one available move at game start")
}

func TestProcessPlayerMoveValid(t *testing.T) {
	conn := NewConnection(t, nil)

	conn.ProcessMessage(play.MessageTypePlayerMove, validFirstMove)

	require.Empty(t, conn.MessagesOfType(play.MessageTypeInvalidMove),
		"valid move should not produce an invalidMove response")

	moves := availableMovesFromMessage(t, requireSingleMessage(t, conn, play.MessageTypeAvailableMoves))
	require.NotEmpty(t, moves, "expected available moves for the next player after a valid move")
}

func TestProcessPlayerMoveInvalid(t *testing.T) {
	conn := NewConnection(t, nil)

	conn.ProcessMessage(play.MessageTypePlayerMove, play.PGNMove("a1-a2"))

	msg := requireSingleMessage(t, conn, play.MessageTypeInvalidMove)
	reason := dataFromMessage[string](t, msg)
	require.NotEmpty(t, reason, "invalidMove response should include a reason string")

	require.Empty(t, conn.MessagesOfType(play.MessageTypeAvailableMoves),
		"invalid move should not advance the game (no new availableMoves)")
}

func TestProcessSaveGame(t *testing.T) {
	conn := NewConnection(t, nil)

	conn.ProcessMessage(play.MessageTypeSaveGame, nil)

	resp := dataFromMessage[play.SaveGameResponse](t, requireSingleMessage(t, conn, play.MessageTypeSaveGameResponse))
	require.Empty(t, resp.PGN, "fresh game's PGN should be empty")
}

func TestProcessSaveGameAfterPlayerMove(t *testing.T) {
	conn := NewConnection(t, nil)

	conn.ProcessMessage(play.MessageTypePlayerMove, validFirstMove)
	conn.ProcessMessage(play.MessageTypeSaveGame, nil)

	resp := dataFromMessage[play.SaveGameResponse](t, requireSingleMessage(t, conn, play.MessageTypeSaveGameResponse))
	require.Contains(t, resp.PGN, string(validFirstMove),
		"saved PGN should contain the move that was just played")
}

func TestProcessNewGame(t *testing.T) {
	conn := NewConnection(t, nil)

	conn.ProcessMessage(play.MessageTypeNewGame, nil)

	resp := dataFromMessage[play.LoadGameResponse](t, requireSingleMessage(t, conn, play.MessageTypeLoadGameResponse))
	require.Empty(t, resp.PastMoves, "new game should have no past moves")
	require.Equal(t, -1, resp.CurrentMove, "new game's CurrentMove should be -1")

	moves := availableMovesFromMessage(t, requireSingleMessage(t, conn, play.MessageTypeAvailableMoves))
	require.NotEmpty(t, moves, "new game should have legal moves available")
}

func TestProcessLoadGame(t *testing.T) {
	conn := NewConnection(t, nil)

	pgn := "1. h2-h3"
	conn.ProcessMessage(play.MessageTypeLoadGame, pgn)

	resp := dataFromMessage[play.LoadGameResponse](t, requireSingleMessage(t, conn, play.MessageTypeLoadGameResponse))
	require.Equal(t, []play.PGNMove{validFirstMove}, resp.PastMoves)
	require.Equal(t, 0, resp.CurrentMove, "after loading 1 move, CurrentMove should be 0")

	require.NotEmpty(t,
		availableMovesFromMessage(t, requireSingleMessage(t, conn, play.MessageTypeAvailableMoves)),
		"loaded game should have legal moves available for the next player")
}

func TestProcessSetCurrentMove(t *testing.T) {
	conn := NewConnection(t, nil)

	conn.ProcessMessage(play.MessageTypePlayerMove, validFirstMove)
	conn.ProcessMessage(play.MessageTypeSetCurrentMove, float64(0))

	loadResponses := conn.MessagesOfType(play.MessageTypeLoadGameResponse)
	require.NotEmpty(t, loadResponses, "setCurrentMove should produce a loadGameResponse")
	resp := dataFromMessage[play.LoadGameResponse](t, loadResponses[len(loadResponses)-1])
	require.Equal(t, 0, resp.CurrentMove)
	require.Equal(t, []play.PGNMove{validFirstMove}, resp.PastMoves,
		"PastMoves should be preserved across setCurrentMove")
}

func TestProcessSetCurrentMoveOutOfRange(t *testing.T) {
	conn := NewConnection(t, nil)

	conn.ProcessMessage(play.MessageTypeSetCurrentMove, float64(99))

	require.Empty(t, conn.MessagesOfType(play.MessageTypeLoadGameResponse),
		"out-of-range setCurrentMove should not produce a loadGameResponse")
}

func TestProcessConcurrencyBotsSwitchToHumans(t *testing.T) {
	allBots := &play.Config{
		Depth:        10,
		HumanPlayers: []game.Player{},
		EvalLimit:    0,
	}
	conn := NewConnection(t, allBots)

	conn.ProcessMessage(play.MessageTypeSetSettings, allBots)

	allHumans := *allBots
	allHumans.HumanPlayers = []game.Player{playerRed, playerBlue, playerYellow, playerGreen}
	conn.ProcessMessage(play.MessageTypeSetSettings, allHumans)

	conn.ProcessMessage(play.MessageTypePlayerMove, "k2-k4")

	msgs := conn.Messages()
	fmt.Printf("captured %d messages:\n", len(msgs))
	for i, m := range msgs {
		fmt.Printf("  [%2d] %s %v\n", i, m.Parsed.Type, m.Parsed.Data)
	}
}

func TestProcessConcurrencyBotsSwitchToPartialHumans(t *testing.T) {
	allBots := &play.Config{
		Depth:        4,
		HumanPlayers: []game.Player{},
		EvalLimit:    0,
	}
	conn := NewConnection(t, allBots)

	conn.ProcessMessage(play.MessageTypeSetSettings, allBots)

	allHumans := *allBots
	allHumans.HumanPlayers = []game.Player{playerRed}
	conn.ProcessMessage(play.MessageTypeSetSettings, allHumans)

	conn.ProcessMessage(play.MessageTypePlayerMove, "k2-k4")

	msgs := conn.WaitForMessagesOfType(play.MessageTypeEngineMove, 3)
	fmt.Printf("captured %d messages:\n", len(msgs))
	for i, m := range msgs {
		fmt.Printf("  [%2d] %s %v\n", i, m.Parsed.Type, m.Parsed.Data)
	}
}

func TestProcessSetSettingsActivePlayerStaysEngine(t *testing.T) {
	cfg := &play.Config{
		Depth:        4,
		HumanPlayers: []game.Player{playerRed, playerBlue, playerYellow, playerGreen},
	}
	conn := NewConnection(t, cfg)

	// Remove Red — engine must now play for Red.
	removeRed := *cfg
	removeRed.HumanPlayers = []game.Player{playerBlue, playerYellow, playerGreen}
	conn.ProcessMessage(play.MessageTypeSetSettings, removeRed)

	// Immediately remove Yellow too. Active is still Red, so the existing
	// search should continue and no new one should be launched.
	removeYellow := removeRed
	removeYellow.HumanPlayers = []game.Player{playerBlue, playerGreen}
	conn.ProcessMessage(play.MessageTypeSetSettings, removeYellow)

	conn.WaitForMessagesOfType(play.MessageTypeAvailableMoves, 2)

	require.Len(t, conn.MessagesOfType(play.MessageTypeEngineMove), 1,
		"expected exactly one engineMove; multiple indicate concurrent engine searches")
}

func TestProcessSetSettingsSameHumanPlayers(t *testing.T) {
	conn := NewConnection(t, nil)

	updated := *defaultTestConfig()
	updated.Depth = 2
	updated.EvalLimit = 5

	conn.ProcessMessage(play.MessageTypeSetSettings, updated)

	require.NotEmpty(t,
		availableMovesFromMessage(t, requireSingleMessage(t, conn, play.MessageTypeAvailableMoves)),
		"setSettings should trigger a fresh availableMoves response")
}
