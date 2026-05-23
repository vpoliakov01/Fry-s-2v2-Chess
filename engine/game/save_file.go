package game

import (
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"regexp"
	"strconv"
	"strings"
)

var moveRegex = regexp.MustCompile(`[QRBNK]?([a-n])([1-9][0-4]?){1,2}[-x]?[QRBNK]?([a-n])([1-9][0-4]?)(=[QRBN])?[+#]?`)

// JSON returns json of the game session object.
func (g *GameSession) JSON() ([]byte, error) {
	return json.Marshal(g)
}

// PGN returns the game session in pgn notation.
func (g *GameSession) PGN() string {
	pgn := ""
	for i := 0; i < len(g.PastMoves); i += 4 {
		if i > 0 && i%4 == 0 {
			pgn += "\n"
		}
		pgn += fmt.Sprintf("%v.", i/4+1)
		for j := 0; j < 4 && i+j < len(g.PastMoves); j++ {
			pgn += fmt.Sprintf(" %v", g.PastMoves[i+j])
		}
	}
	return pgn
}

// LoadJSON returns the game session defined by the json.
func LoadJSON(bytes []byte) (*GameSession, error) {
	g := GameSession{}

	err := json.Unmarshal(bytes, &g)
	if err != nil {
		return nil, err
	}

	g.Board.SetPieceSquares()
	g.ComputeHash()

	return &g, nil
}

// LoadPGN returns the moves (pgn notation) specified in the file.
func LoadPGN(pgn string) (*GameSession, error) {
	moves, err := ParsePGN(pgn)
	if err != nil {
		return nil, err
	}

	g := NewGameSession()
	for _, move := range moves {
		g.Play(move)
	}

	return g, nil
}

// ParseMove parses a move from a string.
func ParseMove(m string) (*Move, error) {
	matches := moveRegex.FindStringSubmatch(m)

	if len(matches) < 5 {
		return nil, fmt.Errorf("bad io %v", matches)
	}

	fromFile := int(matches[1][0]) - int('a')
	fromRank, err := strconv.Atoi(matches[2])
	if err != nil {
		return nil, err
	}
	fromRank--

	toFile := int(matches[3][0]) - int('a')
	toRank, err := strconv.Atoi(matches[4])
	if err != nil {
		return nil, err
	}
	toRank--

	return &Move{
		From: Square{fromRank, fromFile},
		To:   Square{toRank, toFile},
	}, nil
}

// ParsePGN parses pgn from a string.
func ParsePGN(pgn string) ([]Move, error) {
	lines := strings.Split(pgn, "\n")
	moves := []Move{}

	for _, line := range lines {
		if len(line) < 4 || line[0] == '[' {
			continue
		}

		turnMovesStr := strings.Split(line, ". ")[1]

		for _, moveStr := range strings.Split(turnMovesStr, " ") {
			move, err := ParseMove(moveStr)
			if err != nil {
				return nil, err
			}
			moves = append(moves, *move)
		}

	}
	return moves, nil
}

// LoadFile attempts to load a game from pgn and if it fails,
// it attempts to load it from json.
func LoadFile(file string) (*GameSession, error) {
	bytes, err := ioutil.ReadFile(file)
	if err != nil {
		return nil, err
	}

	g, err := LoadPGN(string(bytes))
	if err == nil {
		return g, nil
	}

	fmt.Printf("Failed to load pgn: %v\nAttempting to load json\n", err)

	g, err = LoadJSON(bytes)
	if err != nil {
		fmt.Printf("Failed to load json: %v\n", err)
		return nil, err
	}

	return g, nil
}

// SaveToFile saves the game session to a file.
func SaveToFile(g *GameSession) (string, error) {
	bytes := []byte(g.PGN())

	hash := sha256.Sum256(bytes)
	file := fmt.Sprintf("%x.save", hash[0:2])

	err := ioutil.WriteFile(file, bytes, 0o644)
	if err != nil {
		return "", err
	}

	return file, nil
}

// SetupBoard loads a game from a file if it exists, otherwise it creates a new game.
func SetupBoard(loadFile string) *GameSession {
	if loadFile != "" {
		g, err := LoadFile(loadFile)
		if err != nil {
			panic(err)
		}
		return g
	}

	return NewGameSession()
}
