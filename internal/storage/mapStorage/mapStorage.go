package mapStorage

import (
	"shortner/internal/domain"
	"shortner/internal/storage/types"
	"sync"
)

type MapStorage struct {
	data         map[string]*domain.Link
	idAliasIndex map[int64]string

	mu sync.RWMutex
}

func NewMapStorage() *MapStorage {
	return &MapStorage{data: make(map[string]*domain.Link),
		idAliasIndex: make(map[int64]string)}
}

func (s *MapStorage) PutLink(link *domain.Link) (string, error) {
	s.mu.Lock()
	val, ok := s.data[link.Alias]
	s.mu.Unlock()

	if !ok {
		return "", types.ErrAlreadyExists
	}
	return val.Alias, nil
}

func (s *MapStorage) GetLink(alias string) (*domain.Link, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	link, ok := s.data[alias]
	if !ok {
		return nil, types.ErrNotFound
	}
	return link, nil
}

func (s *MapStorage) ValuePresent(link *domain.Link) (string, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	for key, v := range s.data {
		if v == link {
			return key, true
		}
	}
	return "", false
}

func (s *MapStorage) KeyPresent(alias string) bool {
	s.mu.RLock()
	defer s.mu.RUnlock()
	_, ok := s.data[alias]
	return ok
}
