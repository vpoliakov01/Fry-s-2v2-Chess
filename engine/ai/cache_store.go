package ai

import (
	"bufio"
	"encoding/binary"
	"fmt"
	"io"
	"log"
	"os"
	"time"
)

const (
	CacheTempPath = "/tmp/2v2chessai_cache.bin"
)

// On-disk format for the transposition table. Only non-empty entries are written.
// Header (32 bits):  headerType(4) | version(1) | sizeBits(1) | count(4) | reserved(22)
// Entry (80 bits):   index(32) | key(32) | score(16) | depth(8) | from(8) | to(8) | bound(8)
const (
	cacheHeaderType = "TTBL"
	cacheVersion    = uint8(1)
	cacheEntryBytes = 14
)

// Store writes the non-empty entries of t to path.
func (t *TranspositionTable) Store(path string) error {
	start := time.Now()

	for i := range t.locks {
		t.locks[i].Lock()
	}
	defer func() {
		for i := range t.locks {
			t.locks[i].Unlock()
		}
	}()

	f, err := os.Create(CacheTempPath)
	if err != nil {
		return err
	}
	defer f.Close()

	w := bufio.NewWriter(f)

	count := uint32(0)
	for i := range t.entries {
		if t.entries[i].Key != 0 {
			count++
		}
	}

	header := [32]byte{}
	copy(header[0:4], cacheHeaderType)
	header[4] = cacheVersion
	header[5] = uint8(TTSizeBits)
	binary.LittleEndian.PutUint32(header[6:10], count)

	_, err = w.Write(header[:])
	if err != nil {
		return err
	}

	buf := [cacheEntryBytes]byte{}
	for i := range t.entries {
		e := &t.entries[i]
		if e.Key == 0 {
			continue
		}

		key32 := uint64(e.Key)

		binary.LittleEndian.PutUint32(buf[0:4], uint32(i))
		binary.LittleEndian.PutUint32(buf[4:8], uint32(key32))
		binary.LittleEndian.PutUint16(buf[8:10], uint16(e.Score))
		buf[10] = uint8(e.Depth)
		buf[11] = e.From
		buf[12] = e.To
		buf[13] = e.Bound

		if _, err := w.Write(buf[:]); err != nil {
			return err
		}
	}

	err = w.Flush()
	if err != nil {
		return err
	}

	err = os.Rename(CacheTempPath, path)
	if err != nil {
		return err
	}

	log.Printf("Transposition table stored to %s in %v (%.2f%% full)", path, time.Since(start), float64(count)/float64(TTSize)*100)
	return nil
}

// Load restores entries previously written by Store. Existing entries are cleared first.
func (t *TranspositionTable) Load(path string) error {
	start := time.Now()

	f, err := os.Open(path)
	if err != nil {
		return err
	}
	defer f.Close()

	r := bufio.NewReader(f)

	header := [32]byte{}
	if _, err := io.ReadFull(r, header[:]); err != nil {
		return err
	}
	if string(header[0:4]) != cacheHeaderType {
		return fmt.Errorf("invalid magic %q", header[0:4])
	}
	if header[4] != cacheVersion {
		return fmt.Errorf("unsupported version %d", header[4])
	}
	if header[5] != uint8(TTSizeBits) {
		return fmt.Errorf("size bits mismatch: file %d, expected %d", header[5], TTSizeBits)
	}

	count := binary.LittleEndian.Uint32(header[6:10])

	for i := range t.locks {
		t.locks[i].Lock()
	}
	defer func() {
		for i := range t.locks {
			t.locks[i].Unlock()
		}
	}()

	clear(t.entries)

	buf := [cacheEntryBytes]byte{}
	for i := uint32(0); i < count; i++ {
		if _, err := io.ReadFull(r, buf[:]); err != nil {
			return err
		}

		index := binary.LittleEndian.Uint32(buf[0:4])

		t.entries[index] = Entry{
			Key:   binary.LittleEndian.Uint32(buf[4:8]),
			Score: int16(binary.LittleEndian.Uint16(buf[8:10])),
			Depth: int8(buf[10]),
			From:  buf[11],
			To:    buf[12],
			Bound: buf[13],
		}
	}

	log.Printf("Transposition table loaded from %s in %v (%.2f%% full)", path, time.Since(start), float64(count)/float64(TTSize)*100)
	return nil
}
