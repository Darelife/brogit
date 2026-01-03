package store

import (
	"encoding/json"
	"fmt"
	"os"
	"sync"
	"time"

	"brogit/pkg/api"
)

const (
	MaxBufferSize = 10 * 1024 * 1024 // 10 MB
	FlushInterval = 5 * time.Second
	StorageFile   = "pending_diffs.jsonl"
)

type Store struct {
	mu         sync.Mutex
	buffer     []api.DiffEntry
	bufferSize int
	filePath   string
	lastFlush  time.Time
}

func NewStore(filePath string) *Store {
	if filePath == "" {
		filePath = StorageFile
	}
	s := &Store{
		filePath:  filePath,
		lastFlush: time.Now(),
	}
	go s.startTicker()
	return s
}

func (s *Store) Add(entry api.DiffEntry) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	// Approximate size of entry
	bytes, err := json.Marshal(entry)
	if err != nil {
		return err
	}
	size := len(bytes)

	s.buffer = append(s.buffer, entry)
	s.bufferSize += size

	if s.bufferSize >= MaxBufferSize {
		return s.flushLocked()
	}

	return nil
}

func (s *Store) startTicker() {
	ticker := time.NewTicker(FlushInterval)
	for range ticker.C {
		s.mu.Lock()
		if time.Since(s.lastFlush) >= FlushInterval && len(s.buffer) > 0 {
			_ = s.flushLocked()
		}
		s.mu.Unlock()
	}
}

// flushLocked assumes the caller holds the lock
func (s *Store) flushLocked() error {
	if len(s.buffer) == 0 {
		return nil
	}

	f, err := os.OpenFile(s.filePath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return fmt.Errorf("failed to open storage file: %w", err)
	}
	defer f.Close()

	encoder := json.NewEncoder(f)
	for _, entry := range s.buffer {
		if err := encoder.Encode(entry); err != nil {
			return fmt.Errorf("failed to encode entry: %w", err)
		}
	}

	s.buffer = nil
	s.bufferSize = 0
	s.lastFlush = time.Now()
	fmt.Println("Flushed buffer to disk")
	return nil
}

// GetAll returns all entries from the file and current buffer
// NOTE: This is expensive and mostly for the commit phase
func (s *Store) GetAll() ([]api.DiffEntry, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	// First flush whatever is in memory
	if err := s.flushLocked(); err != nil {
		return nil, err
	}

	// Read from file
	f, err := os.Open(s.filePath)
	if os.IsNotExist(err) {
		return []api.DiffEntry{}, nil
	}
	if err != nil {
		return nil, err
	}
	defer f.Close()

	var entries []api.DiffEntry
	decoder := json.NewDecoder(f)
	for decoder.More() {
		var entry api.DiffEntry
		if err := decoder.Decode(&entry); err != nil {
			return nil, err
		}
		entries = append(entries, entry)
	}

	return entries, nil
}

func (s *Store) Clear() error {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.buffer = nil
	s.bufferSize = 0
	
	// Truncate file
	if err := os.Truncate(s.filePath, 0); err != nil {
		return err
	}
	
	return nil
}
