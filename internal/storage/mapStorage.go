package storage

import (
	"fmt"
	"sync"
)

type MapStorage struct {
	data map[string]string
	mu   sync.RWMutex
}

func NewMapStorage() *MapStorage {
	return &MapStorage{data: make(map[string]string)}
}

func (s *MapStorage) PutLink(shortLink string, link string) string {
	s.mu.Lock()
	s.data[shortLink] = link
	s.mu.Unlock()
	return shortLink
}

func (s *MapStorage) GetLink(shortLink string) (string, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	elem, ok := s.data[shortLink]
	if !ok {
		return "", fmt.Errorf("link not found (%s)", shortLink)
	}
	return elem, nil
}

func (s *MapStorage) ValuePresent(link string) (string, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	for key, v := range s.data {
		if v == link {
			return key, true
		}
	}
	return "", false
}

func (s *MapStorage) KeyPresent(shortLink string) bool {
	s.mu.RLock()
	defer s.mu.RUnlock()
	_, ok := s.data[shortLink]
	return ok
}
