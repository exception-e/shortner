package storage

import (
	"fmt"
)

type MapStorage struct {
	data map[string]string
}

func NewMapStorage() *MapStorage {
	return &MapStorage{data: make(map[string]string)}
}

func (s *MapStorage) PutLink(shortLink string, link string) string {
	s.data[shortLink] = link
	return shortLink
}

func (s *MapStorage) GetLink(shortLink string) (string, error) {
	elem, ok := s.data[shortLink]
	if !ok {
		return "", fmt.Errorf("link not found (%s)", shortLink)
	}

	return elem, nil
}

func (s *MapStorage) ValuePresent(link string) (string, bool) {
	for key, v := range s.data {
		if v == link {
			return key, true
		}
	}

	return "", false
}

func (s *MapStorage) KeyPresent(shortLink string) bool {
	_, ok := s.data[shortLink]
	return ok
}
