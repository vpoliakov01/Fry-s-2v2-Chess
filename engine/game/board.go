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
	Kings        [4]Square                     `json:"-"` // Per-player square of the king
	PieceSquares [4][]Square                   `json:"-"` // Per-player slice of occupied squares
	pieceIndex   [4][BoardSize][BoardSize]int8 `json:"-"` // Maps each occupied square to its index in PieceSquares for O(1) removal via swap-with-last

}

// NewBoard creates a new board.
func NewBoard() *Board {
	b := Board{}

	for rank := 0; rank < BoardSize; rank++ {
		for file := 0; file < BoardSize; file++ {
			square := Square{Rank: rank, File: file}

			if square.IsValid() {
				b.Grid[rank][file] = Piece(EmptySquare)

				for player := 0; player < 4; player++ {
					b.pieceIndex[player][rank][file] = noPieceIndex
				}
			} else {
				b.Grid[rank][file] = Piece(InactiveSquare)
			}
		}
	}

	for player := 0; player < 4; player++ {
		b.PieceSquares[player] = make([]Square, 0, MaxPiecesPerPlayer)
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

	if piece.Kind() == KindKing {
		b.Kings[piece.Player()] = square
	}
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
				b.PlacePiece(NewPiece(player, kind), Square{Rank: rank, File: file})
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

	piece := b.GetPiece(move.From)
	player := piece.Player()

	if piece.Kind() == KindKing {
		b.Kings[player] = move.To
	}

	b.removeFromPieceSquares(player, move.From)
	b.addToPieceSquares(player, move.To)

	b.Grid[move.To.Rank][move.To.File] = b.Grid[move.From.Rank][move.From.File]
	b.Grid[move.From.Rank][move.From.File] = Piece(EmptySquare)
}

// Unmove undoes a move of a piece on the board.
func (b *Board) Unmove(move Move, capturedPiece Piece) {
	piece := b.GetPiece(move.To)
	player := piece.Player()

	b.Grid[move.From.Rank][move.From.File] = b.Grid[move.To.Rank][move.To.File]
	b.Grid[move.To.Rank][move.To.File] = capturedPiece

	b.removeFromPieceSquares(player, move.To)
	b.addToPieceSquares(player, move.From)

	if piece.Kind() == KindKing {
		b.Kings[player] = move.From
	}

	if !capturedPiece.IsEmpty() {
		b.addToPieceSquares(capturedPiece.Player(), move.To)

	}
}
