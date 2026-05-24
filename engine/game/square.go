package game

import (
	"fmt"
	"strconv"
)

// Square stores a coordinate on the board.
type Square struct {
	Rank int `json:"rank"`
	File int `json:"file"`
}

// Add adds a vector to the square.
func (s *Square) Add(rank, file int) Square {
	return Square{s.Rank + rank, s.File + file}
}

// String implements the Stringer interface.
func (s Square) String() string {
	return fmt.Sprintf("%v%v", fmt.Sprintf("%c", int('a')+s.File), s.Rank+1)
}

// IsValid checs if the square is on the board.
func (s *Square) IsValid() bool {
	return IsSquareValid(s.Rank, s.File)
}

// IsSquareValid returns whether rank and file are within [0, 13] and outside the excluded corners.
func IsSquareValid(rank, file int) bool {
	return !((file < CornerSize || file >= BoardSize-CornerSize) && (rank < CornerSize || rank >= BoardSize-CornerSize)) &&
		(file >= 0 && file < BoardSize && rank >= 0 && rank < BoardSize)
}

// IsWithin returns whether the square is within the given distance from the other square.
func (s Square) IsWithin(other Square, distance float64) bool {
	return float64((s.Rank-other.Rank)*(s.Rank-other.Rank)+(s.File-other.File)*(s.File-other.File)) <= distance*distance
}

func SquareFromPGN(pgn string) Square {
	rank, err := strconv.Atoi(string(pgn[1:]))
	if err != nil {
		return Square{}
	}
	return Square{Rank: rank - 1, File: int(pgn[0] - 'a')}
}
