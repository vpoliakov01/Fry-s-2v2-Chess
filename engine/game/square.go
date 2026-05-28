package game

import (
	"fmt"
	"strconv"
)

// Square stores a coordinate on the board.
type Square struct {
	File int `json:"file"`
	Rank int `json:"rank"`
}

// Add adds a vector to the square.
func (s *Square) Add(file, rank int) Square {
	return Square{File: s.File + file, Rank: s.Rank + rank}
}

// String implements the Stringer interface.
func (s Square) String() string {
	return fmt.Sprintf("%v%v", fmt.Sprintf("%c", int('a')+s.File), s.Rank+1)
}

// IsValid checs if the square is on the board.
func (s *Square) IsValid() bool {
	return !((s.File < CornerSize || s.File >= BoardSize-CornerSize) && (s.Rank < CornerSize || s.Rank >= BoardSize-CornerSize)) &&
		(s.File >= 0 && s.File < BoardSize && s.Rank >= 0 && s.Rank < BoardSize)
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
	return Square{File: int(pgn[0] - 'a'), Rank: rank - 1}
}
