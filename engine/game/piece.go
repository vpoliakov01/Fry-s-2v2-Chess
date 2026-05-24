package game

import (
	"fmt"

	"github.com/vpoliakov01/2v2ChessAI/engine/color"
)

const (
	// Store the piece as ppkkk (last 3 bits specify the kind, 2 bits before them specify the player).
	pieceBitOffset = 3
	pieceBitMask   = 7 // 00000111.
	pieceVariants  = 32
)

type Piece uint8 // Use uint8 to save some space (the board is a [][]Piece).

type PieceKind uint8

const (
	// Set values from 0 to 7.
	EmptySquare Piece = iota
	InactiveSquare
	KindPawn PieceKind = iota
	KindKnight
	KindBishop
	KindRook
	KindQueen
	KindKing
)

var (
	printMap = map[PieceKind]string{
		KindPawn:   "♟",
		KindKnight: "♞",
		KindBishop: "♝",
		KindRook:   "♜",
		KindQueen:  "♛",
		KindKing:   "♚",
	}
	colorMap = map[Player]color.Color{
		0: color.Red,
		1: color.Blue,
		2: color.Yellow,
		3: color.Green,
	}

	knightDirs = [][2]int{{-2, -1}, {-2, 1}, {-1, -2}, {-1, 2}, {1, -2}, {1, 2}, {2, -1}, {2, 1}}
	bishopDirs = [][2]int{{-1, -1}, {-1, 1}, {1, -1}, {1, 1}}
	rookDirs   = [][2]int{{-1, 0}, {0, -1}, {0, 1}, {1, 0}}
	queenDirs  = [][2]int{{-1, 0}, {0, -1}, {0, 1}, {1, 0}, {-1, -1}, {-1, 1}, {1, -1}, {1, 1}}
	kingDirs   = [][2]int{{-1, 0}, {0, -1}, {0, 1}, {1, 0}, {-1, -1}, {-1, 1}, {1, -1}, {1, 1}}

	dirsByKind = [8][][2]int{}
)

func init() {
	dirsByKind[KindKnight] = knightDirs
	dirsByKind[KindBishop] = bishopDirs
	dirsByKind[KindRook] = rookDirs
	dirsByKind[KindQueen] = queenDirs
	dirsByKind[KindKing] = kingDirs
}

// New creates a new Piece.
func NewPiece(player Player, kind PieceKind) Piece {
	return Piece(int(player)<<pieceBitOffset + int(kind))
}

// Player returns the owner of the piece.
func (p Piece) Player() Player {
	return Player(p >> pieceBitOffset)
}

// Kind returns the kind of the piece.
func (p Piece) Kind() PieceKind {
	return PieceKind(p & pieceBitMask)
}

// IsEmpty returns true if the piece is empty.
func (p Piece) IsEmpty() bool {
	return p == EmptySquare
}

// String implements the Stringer interface.
func (p Piece) String() string {
	switch p {
	case InactiveSquare:
		return "███"
	case EmptySquare:
		return "   "
	default:
		return fmt.Sprintf(" %v%v%v ", colorMap[p.Player()], printMap[p.Kind()], color.Reset)
	}
}

// GetMoves appends the moves this piece can make to dst and returns the extended slice.
// Dispatches on Kind() so the call sites in the search hot path can inline.
func (p Piece) GetMoves(board *Board, from Square, dst []Square) []Square {
	kind := p.Kind()

	switch kind {
	case KindQueen, KindBishop, KindRook:
		return GetDirectionalMoves(board, from, dirsByKind[kind], dst)
	case KindKnight, KindKing:
		return GetEnumeratedMoves(board, from, dirsByKind[kind], dst)
	case KindPawn:
		return GetPawnMoves(board, from, dst)
	default:
		panic(fmt.Sprintf("unsupported piece: %v", p))
	}
}
