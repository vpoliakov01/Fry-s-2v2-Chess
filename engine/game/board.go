package game

const (
	BoardSize  = 14
	CornerSize = 3 // 2v2 chess board has corners (3 x 3) cut out.

	// MaxPiecesPerPlayer is the upper bound on per-player piece count (8 pawns + 8 pieces).
	MaxPiecesPerPlayer = 16

	noPieceIndex = int8(-1)
)

// Board represents the chess board.
type Board struct {
	Grid         [BoardSize][BoardSize]Piece   `json:"grid"`
	PieceSquares [4][]Square                   `json:"-"` // Per-player slice of occupied squares
	pieceIndex   [4][BoardSize][BoardSize]int8 `json:"-"` // Maps each occupied square to its index in PieceSquares for O(1) removal via swap-with-last
}

// NewBoard creates a new board.
func NewBoard() *Board {
	b := Board{}

	for player := 0; player < 4; player++ {
		b.PieceSquares[player] = make([]Square, 0, MaxPiecesPerPlayer)

		for rank := 0; rank < BoardSize; rank++ {
			for file := 0; file < BoardSize; file++ {
				b.pieceIndex[player][rank][file] = noPieceIndex
			}
		}
	}

	for rank := 0; rank < BoardSize; rank++ {
		for file := 0; file < BoardSize; file++ {
			if IsSquareValid(rank, file) {
				b.Grid[rank][file] = Piece(EmptySquare)
			} else {
				b.Grid[rank][file] = Piece(InactiveSquare)
			}
		}
	}

	return &b
}

// GetPiece returns a piece from the square.
// NOTE: it does not check the square's validity.
func (b *Board) GetPiece(s Square) Piece {
	return b.Grid[s.Rank][s.File]
}

// IsEmpty checks if the square is empty (no piece).
// NOTE: it does not check the square's validity.
func (b *Board) IsEmpty(s Square) bool {
	return b.Grid[s.Rank][s.File] == Piece(EmptySquare)
}

// Clear clears all the pieces of the board.
func (b *Board) Clear() {
	*b = *NewBoard()
}

// PlacePiece places a piece onto the board.
func (b *Board) PlacePiece(piece Piece, square Square) {
	b.Grid[square.Rank][square.File] = piece
	b.addToPieceSquares(piece.Player(), square)
}

// SetPieceSquares rebuilds PieceSquares from Grid; use after a manual board edit
// or deserialization that didn't go through Play / PlacePiece.
func (b *Board) SetPieceSquares() {
	for player := 0; player < 4; player++ {
		b.PieceSquares[player] = b.PieceSquares[player][:0]

		for rank := 0; rank < BoardSize; rank++ {
			for file := 0; file < BoardSize; file++ {
				b.pieceIndex[player][rank][file] = noPieceIndex
			}
		}
	}

	for rank := 0; rank < BoardSize; rank++ {
		for file := 0; file < BoardSize; file++ {
			square := Square{rank, file}
			if !square.IsValid() || b.IsEmpty(square) {
				continue
			}

			piece := b.GetPiece(square)
			b.addToPieceSquares(piece.Player(), square)
		}
	}
}

// addToPieceSquares records that player owns the piece at square.
func (b *Board) addToPieceSquares(player Player, square Square) {
	b.pieceIndex[player][square.Rank][square.File] = int8(len(b.PieceSquares[player]))
	b.PieceSquares[player] = append(b.PieceSquares[player], square)
}

// removeFromPieceSquares drops square from player's tracking using swap-with-last.
func (b *Board) removeFromPieceSquares(player Player, square Square) {
	idx := b.pieceIndex[player][square.Rank][square.File]
	last := int8(len(b.PieceSquares[player])) - 1

	if idx != last {
		moved := b.PieceSquares[player][last]
		b.PieceSquares[player][idx] = moved
		b.pieceIndex[player][moved.Rank][moved.File] = idx
	}

	b.PieceSquares[player] = b.PieceSquares[player][:last]
	b.pieceIndex[player][square.Rank][square.File] = noPieceIndex
}

// SetStartingPosition sets the pieces for 4 players.
func (b *Board) SetStartingPosition() {
	pieces := [][]PieceKind{
		{KindPawn, KindPawn, KindPawn, KindPawn, KindPawn, KindPawn, KindPawn, KindPawn},
		{KindRook, KindKnight, KindBishop, KindQueen, KindKing, KindBishop, KindKnight, KindRook},
	}

	for row := range pieces {
		for col, kind := range pieces[row] {
			playerPositions := [][]int{
				{1 - row, 3 + col},
				{10 - col, 1 - row},
				{12 + row, 10 - col},
				{3 + col, 12 + row},
			}

			for i := range playerPositions {
				player := Player(i)
				rank := playerPositions[i][0]
				file := playerPositions[i][1]
				b.PlacePiece(NewPiece(player, kind), Square{rank, file})
			}
		}
	}
}

// Copy returns a deep copy of the board.
func (b *Board) Copy() *Board {
	board := *b // copies Grid and pieceIndex arrays by value
	for player := 0; player < 4; player++ {
		board.PieceSquares[player] = append([]Square(nil), b.PieceSquares[player]...)
	}
	return &board
}

// Move performs a move of a piece on the board.
func (b *Board) Move(move Move) {
	if !b.IsEmpty(move.To) {
		capturedPiece := b.GetPiece(move.To)
		b.removeFromPieceSquares(capturedPiece.Player(), move.To)
	}

	player := b.GetPiece(move.From).Player()
	b.removeFromPieceSquares(player, move.From)
	b.addToPieceSquares(player, move.To)

	b.Grid[move.To.Rank][move.To.File] = b.Grid[move.From.Rank][move.From.File]
	b.Grid[move.From.Rank][move.From.File] = Piece(EmptySquare)
}

// Unmove undoes a move of a piece on the board.
func (b *Board) Unmove(move Move, capturedPiece Piece) {
	b.Grid[move.From.Rank][move.From.File] = b.Grid[move.To.Rank][move.To.File]
	b.Grid[move.To.Rank][move.To.File] = capturedPiece

	player := b.GetPiece(move.From).Player()
	b.removeFromPieceSquares(player, move.To)
	b.addToPieceSquares(player, move.From)

	if !capturedPiece.IsEmpty() {
		b.addToPieceSquares(capturedPiece.Player(), move.To)
	}
}
