package game

import "fmt"

var (
	KnightDirs = [][2]int{{-2, -1}, {-2, 1}, {-1, -2}, {-1, 2}, {1, -2}, {1, 2}, {2, -1}, {2, 1}}
	BishopDirs = [][2]int{{-1, -1}, {-1, 1}, {1, -1}, {1, 1}}
	RookDirs   = [][2]int{{-1, 0}, {0, -1}, {0, 1}, {1, 0}}
	QueenDirs  = [][2]int{{-1, 0}, {0, -1}, {0, 1}, {1, 0}, {-1, -1}, {-1, 1}, {1, -1}, {1, 1}}
	KingDirs   = [][2]int{{-1, 0}, {0, -1}, {0, 1}, {1, 0}, {-1, -1}, {-1, 1}, {1, -1}, {1, 1}}

	DirsByKind = [8][][2]int{}
)

func init() {
	DirsByKind[KindKnight] = KnightDirs
	DirsByKind[KindBishop] = BishopDirs
	DirsByKind[KindRook] = RookDirs
	DirsByKind[KindQueen] = QueenDirs
	DirsByKind[KindKing] = KingDirs
}

// GetMoves appends the moves this piece can make to dst and returns the extended slice.
// Dispatches on Kind() so the call sites in the search hot path can inline.
func (p Piece) GetMoves(board *Board, from Square, dst []Square) []Square {
	kind := p.Kind()

	switch kind {
	case KindQueen, KindBishop, KindRook:
		return GetDirectionalMoves(board, from, DirsByKind[kind], dst)
	case KindKnight, KindKing:
		return GetEnumeratedMoves(board, from, DirsByKind[kind], dst)
	case KindPawn:
		return GetPawnMoves(board, from, dst)
	default:
		panic(fmt.Sprintf("unsupported piece: %v", p))
	}
}

// GetDirectionalMoves appends valid moves in the given directions (used for queens, rooks, bishops)
// to dst and returns the extended slice. Each direction is followed until the edge of the board
// or a blocking piece is reached.
func GetDirectionalMoves(board *Board, from Square, vectors [][2]int, dst []Square) []Square {
	fromPlayer := Piece(board.GetPiece(from)).Player()

	for _, vector := range vectors {
		for dist := 1; ; dist++ {
			to := from.Add(dist*vector[0], dist*vector[1])

			if !to.IsValid() {
				break
			} else if board.IsEmpty(to) {
				dst = append(dst, to)
				continue
			} else if !Piece(board.GetPiece(to)).Player().IsTeamMate(fromPlayer) {
				dst = append(dst, to)
			}
			break
		}
	}

	return dst
}

// GetEnumeratedMoves appends valid moves produced by adding each vector to from (used for kings
// and knights) to dst and returns the extended slice.
func GetEnumeratedMoves(board *Board, from Square, vectors [][2]int, dst []Square) []Square {
	fromPlayer := Piece(board.GetPiece(from)).Player()

	for _, vector := range vectors {
		to := from.Add(vector[0], vector[1])

		if !to.IsValid() {
			continue
		} else if board.IsEmpty(to) || !Piece(board.GetPiece(to)).Player().IsTeamMate(fromPlayer) {
			dst = append(dst, to)
		}
	}

	return dst
}

// GetOccupiedSquareInDirection returns the first piece found in the given direction from the given square, or EmptySquare if none is found.
func GetOccupiedSquareInDirection(board *Board, from Square, vector [2]int) Square {
	for dist := 1; ; dist++ {
		to := from.Add(dist*vector[0], dist*vector[1])

		if !to.IsValid() {
			break
		} else if !board.IsEmpty(to) {
			return to
		}
	}

	return Square{}
}

// GetAttackers returns the squares of all pieces that can attack the given square.
func GetAttackers(board *Board, square Square, dst []Square) []Square {
	player := board.GetPiece(square).Player()

	// Check knight directions.
	for _, vector := range KnightDirs {
		from := square.Add(vector[0], vector[1])

		if !from.IsValid() || board.IsEmpty(from) {
			continue
		}

		piece := board.GetPiece(from)
		if piece.Kind() == KindKnight && !piece.Player().IsTeamMate(player) {
			dst = append(dst, from)
		}
	}

	// Check bishop directions.
	for _, vector := range BishopDirs {
		from := GetOccupiedSquareInDirection(board, square, vector)

		if !from.IsValid() {
			continue
		}

		piece := board.GetPiece(from)
		if !piece.Player().IsTeamMate(player) && (piece.Kind() == KindQueen || piece.Kind() == KindBishop) {
			dst = append(dst, from)
		}
	}

	// Check rook directions.
	for _, vector := range RookDirs {
		from := GetOccupiedSquareInDirection(board, square, vector)

		if !from.IsValid() {
			continue
		}

		piece := board.GetPiece(from)
		if !piece.Player().IsTeamMate(player) && (piece.Kind() == KindQueen || piece.Kind() == KindRook) {
			dst = append(dst, from)
		}
	}

	// Check pawns.
	opponents := player.Opponents()
	for _, opponent := range opponents {
		for _, vector := range PawnCaptureDirs[opponent] {
			from := square.Add(-vector[0], -vector[1]) // Subtract because the pawn should capture square.

			if !from.IsValid() || board.IsEmpty(from) {
				continue
			}

			piece := board.GetPiece(from)
			if piece.Kind() == KindPawn && piece.Player() == opponent {
				dst = append(dst, from)
			}
		}
	}

	return dst
}
