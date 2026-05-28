package ai

import (
	"sync"

	"github.com/vpoliakov01/2v2ChessAI/engine/game"
)

const (
	TTSizeBits       = 28
	TTSize           = 1 << TTSizeBits
	TTIndexMask      = TTSize - 1
	TTLockShardCount = 256
	TTLockShardMask  = TTLockShardCount - 1

	BoundExact uint8 = 0
	BoundLower uint8 = 1 // fail-high: score is a lower bound
	BoundUpper uint8 = 2 // fail-low: score is an upper bound

	StoringDepthThreshold = 24 // Don't store positions with moveNumber > this value.
)

// entry is one slot in the transposition table.
// fromIndex and toIndex pack a Square into a byte (high nibble Rank, low nibble File).
type entry struct {
	key   uint32
	score int16
	depth int8 // remaining depth searched below this node when stored
	from  uint8
	to    uint8
	bound uint8
}

// TranspositionTable is a transposition table shared across workers, sharded by mutex to keep contention low.
type TranspositionTable struct {
	entries []entry
	locks   [TTLockShardCount]sync.Mutex
}

// Cache is a wrapper around the transposition table that allows for storing based on depth.
type Cache struct {
	Stored    *TranspositionTable
	NotStored *TranspositionTable
}

// NewCache creates a new cache.
func NewCache() *Cache {
	return &Cache{
		Stored:    NewTranspositionTable(),
		NotStored: NewTranspositionTable(),
	}
}

// Get returns the entry for key; ok is false if the slot holds a different key.
func (c *Cache) Get(key uint64) (cachedEntry entry, ok bool) {
	e, ok := c.Stored.Get(key)
	if ok {
		return e, true
	}
	return c.NotStored.Get(key)
}

// Set writes an entry using depth-preferred replacement.
func (c *Cache) Set(key uint64, move game.Move, score float64, depth int8, bound uint8, moveNumber int) {
	if moveNumber <= StoringDepthThreshold {
		c.Stored.Set(key, move, score, depth, bound)
	} else {
		c.NotStored.Set(key, move, score, depth, bound)
	}
}

// NewTranspositionTable allocates a fresh transposition table.
func NewTranspositionTable() *TranspositionTable {
	return &TranspositionTable{
		entries: make([]entry, TTSize),
	}
}

// Clear zeroes every entry. Intended for fresh-game resets so old positions
// don't compete for slots with new ones.
func (t *TranspositionTable) Clear() {
	for i := range t.locks {
		t.locks[i].Lock()
	}

	clear(t.entries)

	for i := range t.locks {
		t.locks[i].Unlock()
	}
}

// Get returns the entry for key; ok is false if the slot holds a different key.
func (t *TranspositionTable) Get(key uint64) (cachedEntry entry, ok bool) {
	index := key & TTIndexMask
	lock := &t.locks[index&TTLockShardMask]

	lock.Lock()
	e := t.entries[index]
	lock.Unlock()

	key32 := uint32(key >> 32)

	if e.key != key32 {
		return entry{}, false
	}
	return e, true
}

// Set writes an entry using depth-preferred replacement.
func (t *TranspositionTable) Set(key uint64, move game.Move, score float64, depth int8, bound uint8) {
	index := key & TTIndexMask
	lock := &t.locks[index&TTLockShardMask]

	key32 := uint32(key >> 32)
	score16 := int16(score * 100) // Store score as centipawns.

	lock.Lock()
	existing := &t.entries[index]
	if existing.key != key32 || existing.depth <= depth {
		*existing = entry{
			key:   key32,
			score: score16,
			depth: depth,
			from:  packSquare(move.From),
			to:    packSquare(move.To),
			bound: bound,
		}
	}
	lock.Unlock()
}

// boundOf classifies a search result given the original alpha and the (possibly tightened) beta.
func boundOf(bestScore, alphaOrig, beta float64) uint8 {
	switch {
	case bestScore <= alphaOrig:
		return BoundUpper
	case bestScore >= beta:
		return BoundLower
	default:
		return BoundExact
	}
}

// canCutoff reports whether a stored entry's bound permits its score to be used as a cutoff.
func canCutoff(score, alpha, beta float64, bound uint8) bool {
	switch bound {
	case BoundExact:
		return true
	case BoundLower:
		return score >= beta
	case BoundUpper:
		return score <= alpha
	}
	return false
}

// packSquare encodes a Square into a single byte: high nibble Rank, low nibble File.
func packSquare(s game.Square) uint8 {
	return uint8(s.Rank<<4 | s.File)
}

// unpackSquare reverses packSquare.
func unpackSquare(b uint8) game.Square {
	return game.Square{File: int(b & 0x0F), Rank: int(b >> 4)}
}

// move returns the entry's move as a game.Move.
func (e entry) move() game.Move {
	return game.Move{From: unpackSquare(e.from), To: unpackSquare(e.to)}
}

// eval returns the entry's score as a float64.
func (e entry) eval() float64 {
	return float64(e.score) / 100
}
