package ai

import (
	"fmt"
	"math"

	. "github.com/vpoliakov01/2v2ChessAI/engine/game"
)

var (
	kingSafeBoxVectors = [4][4][2]int{
		{{-1, 1}, {0, 1}, {0, 2}, {1, 1}},
		{{1, 1}, {1, 0}, {2, 0}, {1, -1}},
		{{-1, -1}, {0, -1}, {0, -2}, {1, -1}},
		{{-1, -1}, {-1, 0}, {-2, 0}, {-1, 1}},
	}
)

// EvaluateCurrent returns the difference between strengths of the team making the move and the opponent team.
// Used to seed the absolute eval at the search root; per-move updates use EvaluateMove for an incremental delta.
func (ai *AI) EvaluateCurrent(g *Game, buffer *buffer) float64 {
	buffer.evalsCount++

	cached, ok := ai.cache.Get(g.Hash)
	if ok {
		return cached.eval()
	}

	playerStrengths := [4]float64{}

	if g.HasEnded() {
		return float64(g.ActivePlayer.Team()*g.Winner) * 1000
	}

	// For each piece, run piece strength evaluation.
	for player := Player(0); player < 4; player++ {
		for _, square := range g.Board.PieceSquares[player] {
			piece := g.Board.GetPiece(square)
			positionEval, _ := GetPieceStrength(g, piece, square)
			playerStrengths[player] += positionEval
		}
	}

	redYellowStrength := playerStrengths[0] + playerStrengths[2] - math.Abs(playerStrengths[0]-playerStrengths[2])/3
	blueGreenStrength := playerStrengths[1] + playerStrengths[3] - math.Abs(playerStrengths[1]-playerStrengths[3])/3

	return float64(g.ActivePlayer.Team()) * (redYellowStrength - blueGreenStrength)
}

// EvaluateMove returns the move's score from the moving player's perspective.
// The positionEval score is used for end on the search position evaluation.
// THe moveEval score is used for move ordering and pruning.
func (ai *AI) EvaluateMove(g *Game, buffer *buffer, eval float64, move Move) (positionEval, moveEval float64) {
	buffer.evalsCount++
	board := g.Board

	piece := board.GetPiece(move.From)
	capturedPiece := board.GetPiece(move.To)
	captureValueP := 0.0
	captureValueM := 0.0

	g.Play(move)
	cached, ok := ai.cache.Get(g.Hash)
	if ok {
		positionEval = -cached.eval() // Opponent's perspective, since we played a move.
		moveEval = positionEval - eval
		g.UnplayMove(move, capturedPiece)
		return
	}

	newStrengthP, newStrengthM := GetPieceStrength(g, piece, move.To)
	kingSafetyP, kingSafetyM := GetKingSafetyScore(g, buffer, piece, eval)
	g.UnplayMove(move, capturedPiece)

	if !capturedPiece.IsEmpty() {
		if capturedPiece.Kind() == KindKing {
			return mateValue, mateValue
		}
		captureValueP, captureValueM = GetPieceStrength(g, capturedPiece, move.To)
	}

	oldStrengthP, oldStrengthM := GetPieceStrength(g, piece, move.From)

	positionEval = eval + (newStrengthP - oldStrengthP) + captureValueP + kingSafetyP
	moveEval = (newStrengthM - oldStrengthM) + captureValueM + kingSafetyM

	return positionEval, moveEval
}

// GetPieceStrength returns the strength of a piece at a given square.
func GetPieceStrength(g *Game, piece Piece, square Square) (positionEval, moveEval float64) {
	positional := GetPositionalStrength(g, piece, square)
	kingThreatP, kingThreatM := GetKingThreatStrength(g, piece, square)

	return positional + kingThreatP, positional + kingThreatM
}

// GetKingThreatStrength returns a score based on how much threat the piece poses to the opponent king.
func GetKingThreatStrength(g *Game, piece Piece, square Square) (positionEval, moveEval float64) {
	board := g.Board
	squares := piece.GetMoves(board, square, g.SquareBuffer[:0])

	for _, to := range squares {
		if !board.IsEmpty(to) {
			capturedPiece := board.GetPiece(to)
			if capturedPiece.Kind() == KindKing {
				return 4.0, 10.0
			}
		}
	}

	for _, opponent := range piece.Player().Opponents() {
		for _, opponentSquare := range board.PieceSquares[opponent] {
			opponentPiece := board.GetPiece(opponentSquare)

			if opponentPiece.Kind() == KindKing {
				if square.IsWithin(opponentSquare, 2) {
					return Strength[piece.Kind()] / 4, Strength[piece.Kind()] / 2
				}
			}
		}
	}

	return 0.0, 0.0
}

// GetKingSafetyScore returns a score based on how safe the active team's kings are.
func GetKingSafetyScore(g *Game, buffer *buffer, piece Piece, eval float64) (positionEval, moveEval float64) {
	board := g.Board
	player := piece.Player()

	kingSquare := board.Kings[player]

	// King safety box
	for _, vector := range kingSafeBoxVectors[player] {
		square := kingSquare.Add(vector[0], vector[1])
		if !square.IsValid() || board.IsEmpty(square) {
			continue
		}

		piece := board.GetPiece(square)
		if piece.Player().IsTeamMate(player) {
			positionEval += 1
			moveEval += 1
		} else {
			positionEval -= 2
			moveEval -= 2
		}
	}

	// Attackers
	attackers := GetAttackers(board, kingSquare, g.SquareBuffer[:0])
	for range attackers {
		positionEval -= 7
		moveEval -= 10
	}

	if eval < -mateThreshold && piece.Kind() == KindKing {
		moveEval += 10
	}

	return positionEval, moveEval
}

// GetPositionalStrength returns the piece's positional strength at the given square.
// Dispatches on Kind() so the call sites in the search hot path can inline.
func GetPositionalStrength(g *Game, piece Piece, square Square) float64 {
	player := piece.Player()
	kind := piece.Kind()

	switch kind {
	case KindPawn:
		return GetPawnStrength(g.Board, square, player)
	case KindKnight, KindBishop, KindQueen:
		return StrengthPrecomputed[kind][square.Rank][square.File]
	case KindRook, KindKing:
		if player.Team() == 1 {
			return StrengthPrecomputed[kind][square.Rank][square.File]
		}
		return StrengthPrecomputed[kind][square.File][square.Rank]
	default:
		panic(fmt.Sprintf("unsupported piece: %v", piece))
	}
}
