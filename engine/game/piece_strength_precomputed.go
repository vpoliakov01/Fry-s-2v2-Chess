package game

var (
	StrengthPrecomputed = [8][BoardSize][BoardSize]float64{}
	PiecePositionBonus  = map[PieceKind]func(Square) float64{
		KindKnight: func(s Square) float64 { return (3*GetBalanceBonus(s)+GetCenterBonus(s))/4 + 0.27 },
		KindBishop: func(s Square) float64 { return 2*GetBalanceBonus(s)/3 + 0.45 },
		KindRook: func(s Square) float64 {
			return (4*GetBalanceBonus(s)+2*GetEdgeBonus(s)+3*GetAttackBonus(s, 1))/16 + 0.6
		},
		KindQueen: func(s Square) float64 { return (2*GetCenterBonus(s)+GetBalanceBonus(s))/6 + 0.7 },
		KindKing:  func(s Square) float64 { return (GetDefenseBonus(s, 1) + GetEdgeBonus(s)) / 2 },
	}
)

func init() {
	for pieceKind := range PiecePositionBonus {
		for rank := 0; rank < BoardSize; rank++ {
			for file := 0; file < BoardSize; file++ {
				StrengthPrecomputed[pieceKind][rank][file] = Strength[pieceKind] * PiecePositionBonus[pieceKind](Square{rank, file})
			}
		}
	}
}
