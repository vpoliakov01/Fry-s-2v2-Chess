package ai

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/vpoliakov01/2v2ChessAI/engine/game"
)

func TestTranspositionTableStoreLoad(t *testing.T) {
	path := filepath.Join(t.TempDir(), "tt.bin")

	cache := NewTranspositionTable()

	type input struct {
		key   uint64
		move  game.Move
		score float64
		depth int8
		bound uint8
	}

	inputs := []input{
		{0x1111111111111111, game.Move{From: game.Square{File: 2, Rank: 1}, To: game.Square{File: 4, Rank: 3}}, 1.25, 5, BoundExact},
		{0xDEADBEEFCAFEBABE, game.Move{From: game.Square{File: 13, Rank: 13}, To: game.Square{File: 12, Rank: 12}}, -99.5, 12, BoundLower},
		{0xFFFFFFFFFFFFFFFF, game.Move{From: game.Square{File: 15, Rank: 0}, To: game.Square{File: 0, Rank: 15}}, 0.0, 0, BoundUpper},
		{42, game.Move{From: game.Square{File: 7, Rank: 7}, To: game.Square{File: 8, Rank: 8}}, 3.5, -3, BoundExact},
	}

	for _, in := range inputs {
		cache.Set(in.key, in.move, in.score, in.depth, in.bound)
	}

	if err := cache.Store(path); err != nil {
		t.Fatalf("Store: %v", err)
	}

	loaded := NewTranspositionTable()
	if err := loaded.Load(path); err != nil {
		t.Fatalf("Load: %v", err)
	}

	for _, in := range inputs {
		e, ok := loaded.Get(in.key)
		if !ok {
			t.Errorf("key %x: missing after Load", in.key)
			continue
		}

		if float64(e.score) != in.score {
			t.Errorf("key %x: score got %v want %v", in.key, e.score, in.score)
		}
		if e.depth != in.depth {
			t.Errorf("key %x: depth got %d want %d", in.key, e.depth, in.depth)
		}
		if e.bound != in.bound {
			t.Errorf("key %x: bound got %d want %d", in.key, e.bound, in.bound)
		}
		if e.move() != in.move {
			t.Errorf("key %x: move got %v want %v", in.key, e.move(), in.move)
		}
	}
}

func TestTranspositionTableLoadOverwrites(t *testing.T) {
	path := filepath.Join(t.TempDir(), "tt.bin")

	src := NewTranspositionTable()
	src.Set(0xAAAA, game.Move{From: game.Square{File: 1, Rank: 1}, To: game.Square{File: 2, Rank: 2}}, 1.0, 1, BoundExact)
	if err := src.Store(path); err != nil {
		t.Fatalf("Store: %v", err)
	}

	dst := NewTranspositionTable()
	dst.Set(0xBBBB, game.Move{From: game.Square{File: 3, Rank: 3}, To: game.Square{File: 4, Rank: 4}}, 2.0, 2, BoundLower)

	if err := dst.Load(path); err != nil {
		t.Fatalf("Load: %v", err)
	}

	if _, ok := dst.Get(0xBBBB); ok {
		t.Errorf("Load did not clear pre-existing entries")
	}
	if _, ok := dst.Get(0xAAAA); !ok {
		t.Errorf("Load did not restore stored entry")
	}
}

func TestTranspositionTableLoadMissingFile(t *testing.T) {
	tt := NewTranspositionTable()
	err := tt.Load(filepath.Join(t.TempDir(), "missing.bin"))
	if !os.IsNotExist(err) {
		t.Fatalf("expected IsNotExist, got %v", err)
	}
}

func TestTranspositionTableLoadRejectsBadMagic(t *testing.T) {
	path := filepath.Join(t.TempDir(), "tt.bin")
	if err := os.WriteFile(path, make([]byte, 32), 0o644); err != nil {
		t.Fatal(err)
	}
	tt := NewTranspositionTable()
	if err := tt.Load(path); err == nil {
		t.Fatalf("expected error for bad magic, got nil")
	}
}
