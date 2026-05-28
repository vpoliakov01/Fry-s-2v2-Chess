package ai

import (
	"math"

	. "github.com/vpoliakov01/2v2ChessAI/engine/game"
)

const (
	PiecesAtTheStart = 64
)

// Strength stores relative base strengths for pieces.
var Strength = [8]float64{
	0.0,  // Empty square
	0.0,  // Inactive square
	1.0,  // Pawn
	2.2,  // Knight
	5.0,  // Bishop
	4.5,  // Rook
	14.0, // Queen
	7.0,  // King
}

// GetCenterBonus returns a value between 0 and 1.
// Squares closest to the center of the board produce 1, closest to the edge - 0.
// The coefficients are for scaling the result to (0, 1) range.
func GetCenterBonus(s Square) float64 {
	return 1.1112 - (math.Abs(float64(s.Rank)-6.5)+math.Abs(float64(s.File)-6.5))/9
}

// GetEdgeBonus returns a value between 0 and 1.
// Squares closest to the edge of the board produce 1, closest to the center - 0.
func GetEdgeBonus(s Square) float64 {
	return ((math.Abs(float64(s.Rank)-6.5) + math.Abs(float64(s.File)-6.5)) - 1) / 9
}

// GetBalanceBonus peaks for squares in the ring equidistant from the center and the edges,
// and is lower both at the dead center and on the edges.
func GetBalanceBonus(s Square) float64 {
	return 1.5 - (GetCenterBonus(s)*GetCenterBonus(s) + GetEdgeBonus(s)*GetEdgeBonus(s))
}

// GetDefenseBonus returns a value between 0 and 1, growing toward the team's own back ranks
// (rank for Red/Yellow, file for Blue/Green) — the defensive end of the board.
func GetDefenseBonus(s Square, team Team) float64 {
	switch team {
	case 1:
		return (math.Abs(float64(s.Rank)-6.5) - 0.5) / 6
	case -1:
		return (math.Abs(float64(s.File)-6.5) - 0.5) / 6
	}
	return 0
}

// GetAttackBonus returns a value between 0 and 1, growing toward the opponents' side of the
// board (the complement of GetDefenseBonus).
func GetAttackBonus(s Square, team Team) float64 {
	return 1 - GetDefenseBonus(s, team)
}

// CalculateBonusCoef calculates the overall bonus coef. (0.5, 1.5)
func CalculateBonusCoef(moves, movesMin, movesMax int, positionCoef float64) float64 {
	return (float64(moves-movesMin)/float64(movesMax-movesMin) + positionCoef) / 2
}

// GetPawnStrength returns the pawn's positional value.
func GetPawnStrength(board *Board, square Square, player Player) float64 {
	coef := 0.9
	for _, dir := range PawnCaptureDirs[player] {
		inFront := square.Add(dir[0], dir[1])
		if !inFront.IsValid() || board.IsEmpty(inFront) {
			continue
		}

		coef += 0.2
	}

	return Strength[KindPawn] * coef
}
