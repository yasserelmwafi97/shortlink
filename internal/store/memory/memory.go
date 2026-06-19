package memory

import (
	"sync"

	"shortlink/internal/store"
)

const maxAttempts = 10

type Store struct {
	mu          sync.Mutex
	shortToLong map[string]string
	longToShort map[string]string
}

func New() *Store {
	return &Store{
		shortToLong: make(map[string]string),
		longToShort: make(map[string]string),
	}
}

func (s *Store) SaveURL(url string, gen func() (string, error)) (string, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if code, ok := s.longToShort[url]; ok {
		return code, nil
	}

	for i := 0; i < maxAttempts; i++ {
		candidate, err := gen()
		if err != nil {
			return "", err
		}
		if _, taken := s.shortToLong[candidate]; taken {
			continue
		}
		s.shortToLong[candidate] = url
		s.longToShort[url] = candidate
		return candidate, nil
	}

	return "", store.ErrCodeExhausted
}

func (s *Store) Lookup(code string) (string, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	url, ok := s.shortToLong[code]
	if !ok {
		return "", store.ErrNotFound
	}
	return url, nil
}

func (s *Store) Close() error {
	return nil
}
