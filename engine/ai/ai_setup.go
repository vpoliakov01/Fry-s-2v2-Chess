package ai

import (
	"errors"
	"fmt"
	"log"
	"math"
	"os"
	"os/signal"
	"runtime"
	"sync"
	"syscall"

	"github.com/vpoliakov01/2v2ChessAI/engine/game"
)

var (
	MaxEvalLimit = int(1e12)
	ErrGameEnded = errors.New("the game has ended")
	ErrNoMoves   = errors.New("no move can be made in this position")
	cpus         = runtime.NumCPU()

	DefaultDepth      = 12
	DefaultSpread     = 8
	DefaultSpreadDrop = 2
	DefaultEvalLimit  = MaxEvalLimit

	shutdownHookOnce sync.Once
)

const (
	DefaultCachePath = "chess.cache"

	mateValue     = 1000.0 // Score of a position whose active player has just been mated.
	mateThreshold = 900.0  // Scores beyond this magnitude are treated as mate scores.
)

type moveScore struct {
	move    game.Move
	posEval float64
	score   float64 // Used for move ordering and pruning.
}

func init() {
	fmt.Printf("Running on %v CPUs (GOMAXPROCS=%v)\n", cpus, runtime.GOMAXPROCS(0))
}

// WithEnableDebug enables debug analytics.
func WithEnableDebug(enableDebug bool) func(*AI) {
	return func(ai *AI) {
		ai.enableDebug = enableDebug
	}
}

// WithEnableDebugLogging enables debug logging.
func WithEnableDebugLogging(enableDebugLogging bool) func(*AI) {
	return func(ai *AI) {
		ai.enableDebugLogging = enableDebugLogging
	}
}

// Stop stops the engine and blocks until the running search has returned.
func (ai *AI) Stop() {
	ai.stopFlag.Store(true)
	for !ai.hasStopped.Load() {
		runtime.Gosched()
	}
}

// ResetCache discards all transposition-table entries. Call when starting a new
// game or loading a position so unrelated old entries don't crowd new ones.
func (ai *AI) ResetCache() {
	if ai.cache != nil {
		ai.cache.Stored.Clear()
		ai.cache.NotStored.Clear()
	}
}

// StoreCache persists the transposition table to DefaultCachePath.
func (ai *AI) StoreCache() error {
	if ai.cache == nil || !ai.enableStoredCache {
		return nil
	}
	return ai.cache.Stored.Store(DefaultCachePath)
}

// LoadCache restores the transposition table from DefaultCachePath. Missing file is not an error.
func (ai *AI) LoadCache() error {
	if ai.cache == nil || !ai.enableStoredCache {
		ai.cache = NewCache()
		return nil
	}
	err := ai.cache.Stored.Load(DefaultCachePath)
	if os.IsNotExist(err) {
		return nil
	}
	return err
}

// installShutdownHook stores ai.cache on SIGINT/SIGTERM and exits.
func (ai *AI) installShutdownHook() {
	shutdownHookOnce.Do(func() {
		c := make(chan os.Signal, 1)
		signal.Notify(c, os.Interrupt, syscall.SIGTERM, syscall.SIGQUIT, syscall.SIGABRT)

		go func() {
			sig := <-c
			ai.Stop()
			if err := ai.StoreCache(); err != nil {
				log.Printf("Failed to store transposition table: %v", err)
			}

			signal.Reset(sig.(syscall.Signal))
			syscall.Kill(syscall.Getpid(), sig.(syscall.Signal))
		}()
	})
}

// loadSharedAlpha returns the current shared root-level alpha as float64.
func (ai *AI) loadSharedAlpha() float64 {
	return math.Float64frombits(ai.sharedAlpha.Load())
}

// raiseSharedAlpha atomically lifts the shared alpha to candidate if candidate is greater.
func (ai *AI) raiseSharedAlpha(candidate float64) {
	for { // Retry until success in case it's overwritten by another worker.
		current := ai.sharedAlpha.Load()
		if candidate <= math.Float64frombits(current) {
			return
		}
		if ai.sharedAlpha.CompareAndSwap(current, math.Float64bits(candidate)) {
			return
		}
	}
}

// fromOpponentScore converts scores between plies, including "mate-in-N" logic.
func fromOpponentScore(score float64) float64 {
	score = -score
	if score > mateThreshold {
		return score - 1
	}
	if score < -mateThreshold {
		return score + 1
	}
	return score
}

// sumEvalsCounts aggregates per-worker eval counts into ai.EvalsCount for external telemetry.
func (ai *AI) sumEvalsCounts() {
	total := 0
	for i := range ai.buffers {
		total += ai.buffers[i].evalsCount
	}
	ai.EvalsCount = total
}
