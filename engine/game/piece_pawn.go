package game

type Pawn Piece

var (
	PawnMoveDirs = [4][3][2]int{
		{{0, 1}, {-1, 1}, {1, 1}},
		{{1, 0}, {1, -1}, {1, 1}},
		{{0, -1}, {-1, -1}, {1, -1}},
		{{-1, 0}, {-1, -1}, {-1, 1}},
	}
	PawnCaptureDirs = [4][2][2]int{
		{{-1, 1}, {1, 1}},
		{{1, -1}, {1, 1}},
		{{-1, -1}, {1, -1}},
		{{-1, -1}, {-1, 1}},
	}
)

// GetPawnMoves appends the pawn's moves to dst and returns the extended slice.
func GetPawnMoves(board *Board, from Square, dst []Square) []Square {
	player := Piece(board.GetPiece(from)).Player()
	dirs := PawnMoveDirs[player]

	// Move forward by 1.
	to := from.Add(dirs[0][0], dirs[0][1])
	if to.IsValid() && board.IsEmpty(to) {
		dst = append(dst, to)

		// Move forward by 2.
		to = from.Add(2*dirs[0][0], 2*dirs[0][1])
		if to.IsValid() && board.IsEmpty(to) {
			// Pawns can capture sideways and end up on other players' pawn starting positions,
			// so checking rank/file alone isn't enough.
			if (player == 0 && from.Rank == 1) ||
				(player == 1 && from.File == 1) ||
				(player == 2 && from.Rank == BoardSize-2) ||
				(player == 3 && from.File == BoardSize-2) {
				dst = append(dst, to)
			}
		}
	}

	// Capture.
	for i := 1; i <= 2; i++ {
		dir := dirs[i]
		to := from.Add(dir[0], dir[1])

		if !to.IsValid() {
			continue
		} else if !board.IsEmpty(to) && !Piece(board.GetPiece(to)).Player().IsTeamMate(player) {
			dst = append(dst, to)
		}
	}

	// TODO: add en passant and promotions.

	return dst
}
