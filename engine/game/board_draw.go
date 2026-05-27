package game

import (
	"fmt"

	"github.com/vpoliakov01/2v2ChessAI/engine/color"
)

// Draw draws the board. Clunky but does the job.
func (b *Board) Draw() {
	fmt.Print("    ")
	for file := 0; file < BoardSize; file++ {
		fmt.Printf(" %v  ", fmt.Sprintf("%c", int('A')+file))
	}
	fmt.Println()

	for rank := BoardSize - 1; rank >= 0; rank-- {
		fmt.Println(color.Reset, "  +---+---+---+---+---+---+---+---+---+---+---+---+---+---+")

		fmt.Printf("%2v ", rank+1)

		for file := 0; file < BoardSize; file++ {
			fmt.Printf("|%v", b.GetPiece(Square{Rank: rank, File: file}))
		}

		fmt.Printf("| %-2v\n", rank+1)
	}
	fmt.Println(color.Reset, "  +---+---+---+---+---+---+---+---+---+---+---+---+---+---+")

	fmt.Print("    ")
	for file := 0; file < BoardSize; file++ {
		fmt.Printf(" %v  ", fmt.Sprintf("%c", int('A')+file))
	}

	fmt.Println()
}

// HumanReadableMove returns a move in a human-readable format.
func HumanReadableMove(b *Board, move Move, fixedWidth bool) string {
	piece := Piece(b.GetPiece(move.From))

	if fixedWidth {
		if !b.IsEmpty(move.To) {
			capturedPiece := Piece(b.GetPiece(move.To))
			return fmt.Sprintf("%vx%v%-7s", piece, capturedPiece, move)
		} else {
			return fmt.Sprintf("%v    %-7s", piece, move)
		}
	}

	if !b.IsEmpty(move.To) {
		capturedPiece := Piece(b.GetPiece(move.To))
		return fmt.Sprintf("%vx%v %v", piece, capturedPiece, move)
	} else {
		return fmt.Sprintf("%v %v", piece, move)
	}
}
