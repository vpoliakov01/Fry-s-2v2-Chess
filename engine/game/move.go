package game

import (
	"fmt"
	"strings"
)

var (
	NullMove = Move{}
)

// Move stores move coordinates.
type Move struct {
	From Square `json:"from"`
	To   Square `json:"to"`
}

// String implements the Stringer interface.
func (m Move) String() string {
	return fmt.Sprintf("%v-%v", m.From, m.To)
}

func MoveFromPGN(pgn string) Move {
	pos := strings.Split(string(pgn), "-")
	return Move{From: SquareFromPGN(pos[0]), To: SquareFromPGN(pos[1])}
}
