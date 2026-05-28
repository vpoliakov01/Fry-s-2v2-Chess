package ai

import (
	"fmt"
	"math"

	. "github.com/vpoliakov01/2v2ChessAI/engine/game"
)

var (
	kingSafeBoxVectors = [4][3][2]int{
		{{-1, 1}, {0, 2}, {1, 1}},
		{{1, 1}, {2, 0}, {1, -1}},
		{{-1, -1}, {0, -2}, {1, -1}},
		{{-1, -1}, {-2, 0}, {-1, 1}},
	}
)

// EvaluateCurrent returns the difference between strengths of the team making the move and the opponent team.
// Used to seed the absolute eval at the search root; per-move updates use EvaluateMove for an incremental delta.
func (ai *AI) EvaluateCurrent(g *Game, buffer *buffer) float64 {
	buffer.evalsCount++

	if g.HasEnded() {
		return float64(g.ActivePlayer.Team()*g.Winner) * mateValue
	}

	cached, ok := ai.cache.Get(g.Hash)
	if ok {
		return cached.Eval()
	}

	playerStrengths := buffer.playerStrengths

	// For each piece, run piece strength evaluation.
	for player := Player(0); player < 4; player++ {
		playerStrengths[player] = 0.0

		for _, square := range g.Board.PieceSquares[player] {
			piece := g.Board.GetPiece(square)
			positionEval, _ := GetPieceStrength(g, piece, square)
			playerStrengths[player] += positionEval
		}

		kingSafety, _ := GetKingSafetyScore(g, buffer, player, g.Board.Kings[player], false)
		playerStrengths[player] += kingSafety
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
	player := piece.Player()
	capturedPiece := board.GetPiece(move.To)
	captureValueP := 0.0
	captureValueM := 0.0

	if ai.debugConfig != nil {
		movePGN := move.String()
		_ = movePGN
	}

	if !capturedPiece.IsEmpty() {
		if capturedPiece.Kind() == KindKing {
			return mateValue, mateValue
		}
		captureValueP, captureValueM = GetPieceStrength(g, capturedPiece, move.To)
	}

	oldStrengthP, oldStrengthM := GetPieceStrength(g, piece, move.From)
	oldPawnStructureP, oldPawnStructureM := GetPawnStructureStrength(g, piece, move.From)
	oldKingSafetyP, oldKingSafetyM := GetKingSafetyScore(g, buffer, player, move.From, true)
	oldTMKingSafetyP, oldTMKingSafetyM := GetKingSafetyScore(g, buffer, player.Teammate(), move.From, false)

	g.Play(move)

	newStrengthP, newStrengthM := GetPieceStrength(g, piece, move.To)
	newPawnStructureP, newPawnStructureM := GetPawnStructureStrength(g, piece, move.To)
	newKingSafetyP, newKingSafetyM := GetKingSafetyScore(g, buffer, player, move.To, false)
	newTMKingSafetyP, newTMKingSafetyM := GetKingSafetyScore(g, buffer, player.Teammate(), move.To, false)

	if piece.Kind() != KindKing {
		pieceThreatRatio := GetPieceThreatRatio(g, piece, move.To)
		newStrengthP *= 1 - pieceThreatRatio
		if newTMKingSafetyP > -5 {
			newStrengthM *= 1 - pieceThreatRatio
		}
	}

	g.UnplayMove(move, capturedPiece)

	pieceStrengthDeltaP := newStrengthP - oldStrengthP
	pieceStrengthDeltaM := newStrengthM - oldStrengthM

	pawnStructureDeltaP := newPawnStructureP - oldPawnStructureP
	pawnStructureDeltaM := newPawnStructureM - oldPawnStructureM

	kingSafetyDeltaP := newKingSafetyP - oldKingSafetyP + newTMKingSafetyP - oldTMKingSafetyP
	kingSafetyDeltaM := newKingSafetyM - oldKingSafetyM + newTMKingSafetyM - oldTMKingSafetyM

	positionEval = pieceStrengthDeltaP + captureValueP + kingSafetyDeltaP + pawnStructureDeltaP + eval
	moveEval = pieceStrengthDeltaM + captureValueM + kingSafetyDeltaM + pawnStructureDeltaM

	return positionEval, moveEval
}

// GetPieceStrength returns the strength of a piece at a given square.
func GetPieceStrength(g *Game, piece Piece, square Square) (positionEval, moveEval float64) {
	positional := GetPositionalStrength(g, piece, square)
	kingThreatP, kingThreatM := GetKingThreatStrength(g, piece, square)

	return positional + kingThreatP, positional + kingThreatM
}

// GetPieceThreatRatio returns the threat ratio [0, 1] signifying how threatened the piece is at the given square.
func GetPieceThreatRatio(g *Game, piece Piece, square Square) (ratio float64) {
	board := g.Board
	attackers := GetAttackers(board, square, g.SquareBuffer[:0])
	pieceValue := Strength[piece.Kind()]

	for _, opponent := range piece.Player().Opponents() {
		maxRatio := 0.0

		for _, attacker := range attackers {
			attackingPiece := board.GetPiece(attacker)
			if attackingPiece.Player() != opponent {
				continue
			}

			if Strength[attackingPiece.Kind()] < pieceValue {
				maxRatio = 0.4
				break
			} else {
				maxRatio = 0.2
			}
		}

		ratio += maxRatio
	}

	return ratio
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
		if square.IsWithin(board.Kings[opponent], 2) {
			return Strength[piece.Kind()] / 4, Strength[piece.Kind()] / 2
		}
	}

	return 0.0, 0.0
}

// GetKingSafetyScore returns a score based on how safe the given player's king is.
func GetKingSafetyScore(g *Game, buffer *buffer, player Player, square Square, boostKingMoves bool) (positionEval, moveEval float64) {
	board := g.Board
	piece := board.GetPiece(square)
	kingSquare := board.Kings[player]

	// King safety box: friendly pieces shield the king, enemy pieces threaten it.
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

	// Attackers on the king.
	attackers := GetAttackers(board, kingSquare, g.SquareBuffer[:0])
	for range attackers {
		positionEval -= 7
		moveEval -= 10
	}

	if boostKingMoves && positionEval <= 1 && piece.Kind() == KindKing && piece.Player() == player {
		moveEval += 3

		if positionEval <= -5 {
			moveEval += 5
		}
	}

	return positionEval, moveEval
}

// GetPawnStructureStrength returns a bonus if the moved piece is a pawn and it's supported by other pawns.
func GetPawnStructureStrength(g *Game, piece Piece, square Square) (positionEval, moveEval float64) {
	if piece.Kind() != KindPawn {
		return 0, 0
	}

	player := piece.Player()

	for _, dir := range PawnCaptureDirs[player] {
		supportingSquare := square.Add(-dir[0], -dir[1])
		if !supportingSquare.IsValid() {
			continue
		}

		supportPiece := g.Board.GetPiece(supportingSquare)
		if supportPiece == piece { // Same player and kind
			positionEval += 0.2
			moveEval += 0.3
		}

		if supportPiece.Player() == player && (supportPiece.Kind() == KindQueen || supportPiece.Kind() == KindBishop) {
			positionEval -= 0.3 // Blocking queen or bishop penalty
			moveEval -= 0.5
		}
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
