package ai_test

import (
	"fmt"
	"math/rand"
	"runtime"
	"sync"
	"time"

	. "github.com/vpoliakov01/2v2ChessAI/engine/ai"
	. "github.com/vpoliakov01/2v2ChessAI/engine/game"
)

var (
	iterations = int(5e7)
	numCPUs    = runtime.NumCPU()
)

// TestBonuses prints the values for the position bonus maps.
func (s *TestSuite) TestBonuses() {
	bonus := 0.0

	testCases := []struct {
		name string
		f    func(Square) float64
	}{
		{
			name: "GetCenterBonus",
			f:    GetCenterBonus,
		},
		{
			name: "GetEdgeBonus",
			f:    GetEdgeBonus,
		},
		{
			name: "GetBalanceBonus",
			f:    GetBalanceBonus,
		},
		{
			name: "GetCenterBonus+GetBalanceBonus",
			f: func(s Square) float64 {
				return (GetCenterBonus(s) + GetBalanceBonus(s)) / 2
			},
		},
		{
			name: "GetEdgeBonus*GetBalanceBonus",
			f: func(s Square) float64 {
				return (GetEdgeBonus(s) + GetBalanceBonus(s)) / 2
			},
		},
		{
			name: "GetDefenseBonus",
			f: func(s Square) float64 {
				return GetDefenseBonus(s, 1)
			},
		},
		{
			name: "GetAttackBonus",
			f: func(s Square) float64 {
				return GetAttackBonus(s, 1)
			},
		},
		{
			name: "GetBalanceBonus",
			f:    GetBalanceBonus,
		},
	}
	for _, tc := range testCases {
		fmt.Println(tc.name)
		sum := 0.0
		for rank := 0; rank < BoardSize; rank++ {
			for file := 0; file < BoardSize; file++ {
				if isCorner(rank, file) {
					fmt.Printf("     ")
				} else {
					fmt.Printf("%.2f ", tc.f(Square{rank, file})+bonus)
					sum += tc.f(Square{rank, file})
				}
			}
			fmt.Println()
		}
		fmt.Printf("avg: %.2f\n\n", sum/(BoardSize*BoardSize-4*9))
	}
}

// TestPrecomputedBonuses prints the precomputed position bonus values for each piece kind.
func (s *TestSuite) TestPrecomputedBonuses() {
	multiplyByPiece := true

	testCases := []struct {
		name string
		kind PieceKind
	}{
		{name: "Knight", kind: KindKnight},
		{name: "Bishop", kind: KindBishop},
		{name: "Rook", kind: KindRook},
		{name: "Queen", kind: KindQueen},
		{name: "King", kind: KindKing},
	}
	for _, tc := range testCases {
		fmt.Println(tc.name)
		sum := 0.0
		for rank := 0; rank < BoardSize; rank++ {
			for file := 0; file < BoardSize; file++ {
				if isCorner(rank, file) {
					fmt.Printf("     ")
				} else {
					v := StrengthPrecomputed[tc.kind][rank][file]
					if multiplyByPiece {
						v *= Strength[tc.kind]
					}
					if v > 10 {
						fmt.Printf("%.1f ", v)
					} else {
						fmt.Printf("%.2f ", v)
					}
					sum += v
				}
			}
			fmt.Println()
		}
		fmt.Printf("avg: %.2f\n\n", sum/(BoardSize*BoardSize-4*9))
	}
}

// TestPieceStrengths prints the full GetStrength() values for each non-pawn piece kind
// when placed (alone) on every valid square of an otherwise empty board.
func (s *TestSuite) TestPieceStrengths() {
	testCases := []struct {
		name string
		kind PieceKind
	}{
		{name: "Knight", kind: KindKnight},
		{name: "Bishop", kind: KindBishop},
		{name: "Rook", kind: KindRook},
		{name: "Queen", kind: KindQueen},
		{name: "King", kind: KindKing},
	}

	g := NewGame()

	for _, tc := range testCases {
		fmt.Println(tc.name)
		sum := 0.0

		for rank := 0; rank < BoardSize; rank++ {
			for file := 0; file < BoardSize; file++ {
				square := Square{rank, file}
				if !square.IsValid() {
					fmt.Printf("     ")
					continue
				}

				board := g.Board
				board.Clear()
				piece := NewPiece(0, tc.kind)
				board.PlacePiece(piece, square)

				v, _ := GetPieceStrength(g, piece, square)

				fmt.Printf("%.2f ", v)
				sum += v
			}
			fmt.Println()
		}
		fmt.Printf("avg: %.2f\n\n", sum/(BoardSize*BoardSize-4*9))
	}
}

func (s *TestSuite) TestSpeed() {
	testCases := []struct {
		name string
		f    func(Square) float64
		f2   func(Square, Team) float64
		f3   func(Square, Square) float64
	}{
		{
			name: "GetEdgeBonus",
			f:    GetEdgeBonus,
		},
		{
			name: "GetCenterBonus",
			f:    GetCenterBonus,
		},
		{
			name: "GetCenterBonus",
			f:    GetCenterBonus,
		},
		{
			name: "GetEdgeBonus",
			f:    GetEdgeBonus,
		},
		{
			name: "GetBalanceBonus",
			f:    GetBalanceBonus,
		},
		{
			name: "GetDefenseBonus",
			f2:   GetDefenseBonus,
		},
		{
			name: "GetAttackBonus",
			f2:   GetAttackBonus,
		},
	}

	// Warm up
	for i := 0; i < iterations; i++ {
		getRandomSquare()
	}

	start := time.Now()
	for i := 0; i < iterations; i++ {
		getRandomSquare()
	}
	timeToGetRandomSquares := time.Since(start)

	for _, tc := range testCases {
		start := time.Now()
		for i := 0; i < iterations; i++ {
			square := getRandomSquare()
			switch tc.name {
			case "GetCenterBonus":
				tc.f(square)
			case "GetEdgeBonus":
				tc.f(square)
			case "GetBalanceBonus":
				tc.f(square)
			case "GetDefenseBonus":
				tc.f2(square, 1)
			case "GetAttackBonus":
				tc.f2(square, 1)
			}
		}
		name := fmt.Sprintf("%-18s", tc.name+":")
		fmt.Printf("%s %dms\n", name, (time.Since(start) - timeToGetRandomSquares).Milliseconds())
	}
}

func (s *TestSuite) TestPrecomputedSpeed() {
	values := make([]float64, numCPUs)
	squares := make([]Square, numCPUs)
	wg := sync.WaitGroup{}

	var precomputed [BoardSize][BoardSize]float64
	for rank := 0; rank < BoardSize; rank++ {
		for file := 0; file < BoardSize; file++ {
			precomputed[rank][file] = GetBalanceBonus(Square{rank, file})
		}
	}

	// Warm up
	wg.Add(numCPUs)
	for i := 0; i < numCPUs; i++ {
		go func(i int) {
			defer wg.Done()
			for j := 0; j < iterations; j++ {
				squares[i] = getRandomSquare()
			}
		}(i)
	}
	wg.Wait()

	wg.Add(numCPUs)
	start := time.Now()
	for i := 0; i < numCPUs; i++ {
		go func(i int) {
			defer wg.Done()
			for j := 0; j < iterations; j++ {
				squares[i] = getRandomSquare()
			}
		}(i)
	}
	wg.Wait()
	timeToGetRandomSquares := time.Since(start)

	wg.Add(numCPUs)
	start = time.Now()
	for i := 0; i < numCPUs; i++ {
		go func(i int) {
			defer wg.Done()
			for j := 0; j < iterations; j++ {
				square := getRandomSquare()
				values[i] += precomputed[square.Rank][square.File]
			}
		}(i)
	}
	wg.Wait()
	fmt.Printf("Precomputed GetBalanceBonus: %dms\n", (time.Since(start) - timeToGetRandomSquares).Milliseconds())

	wg.Add(numCPUs)
	start = time.Now()
	for i := 0; i < numCPUs; i++ {
		go func(i int) {
			defer wg.Done()
			for j := 0; j < iterations; j++ {
				square := getRandomSquare()
				values[i] += GetBalanceBonus(square)
			}
		}(i)
	}
	wg.Wait()
	fmt.Printf("Function GetBalanceBonus: %dms\n", (time.Since(start) - timeToGetRandomSquares).Milliseconds())
	fmt.Printf("Value: %v\n", values)
	fmt.Printf("Value: %v\n", squares)
}

func getRandomSquare() Square {
	return Square{rand.Intn(BoardSize), rand.Intn(BoardSize)}
}

func isCorner(rank, file int) bool {
	return (rank < 3 && file < 3) ||
		(rank < 3 && file >= BoardSize-3) ||
		(rank >= BoardSize-3 && file < 3) ||
		(rank >= BoardSize-3 && file >= BoardSize-3)
}
