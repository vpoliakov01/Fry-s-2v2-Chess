package game

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
	index := b.pieceIndex[player][square.Rank][square.File]
	last := int8(len(b.PieceSquares[player])) - 1

	if index != last {
		moved := b.PieceSquares[player][last]
		b.PieceSquares[player][index] = moved
		b.pieceIndex[player][moved.Rank][moved.File] = index
	}

	b.PieceSquares[player] = b.PieceSquares[player][:last]
	b.pieceIndex[player][square.Rank][square.File] = noPieceIndex
}
