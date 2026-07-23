package service

import (
	"fmt"
	"strconv"

	storageTypes "shortner/internal/storage/types"

	"github.com/spaolacci/murmur3"
)

const base62Alphabet = "0123456789ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz"

type ShorterService struct {
	mapStorage storageTypes.LinkStorage
}

func NewShortnerService(mapStorage storageTypes.LinkStorage) *ShorterService {
	return &ShorterService{mapStorage: mapStorage}
}

func (s *ShorterService) ShortenLink(link string) (string, error) {
	if v, ok := s.mapStorage.ValuePresent(link); ok {
		return "http://localhost:8080/" + v, nil
	}

	shortLink := encodeBase62(getHash(link))
	count := 0
	for s.mapStorage.KeyPresent(shortLink) {
		if count > 50 {
			return "", fmt.Errorf("service: failed to generate unique short link after %d attempts", count)
		}
		shortLink = s.retryIfCollision(link, count)
		count++
	}
	s.mapStorage.PutLink(shortLink, link)
	return "http://localhost:8080/" + shortLink, nil
}

func (s *ShorterService) GetOriginalLink(shortLink string) (string, error) {
	link, err := s.mapStorage.GetLink(shortLink)
	if err != nil {
		return "", fmt.Errorf("service: failed to get original url: %w", err)
	}
	return link, nil
}

func getHash(link string) uint64 {
	return uint64(murmur3.Sum32([]byte(link)))
}

func encodeBase62(hash uint64) string {
	if hash == 0 {
		return "0"
	}
	var byteArr []byte

	for hash > 0 {
		c := base62Alphabet[hash%62]
		byteArr = append([]byte{c}, byteArr...)
		hash = hash / 62
	}
	return string(byteArr)
}

func decodeBase62(link string) uint64 {
	var num uint64 = 0
	for _, ch := range link {
		var num1 uint64 = 0
		if ch >= '0' && ch <= '9' {
			num1 = uint64(ch - '0')
		}
		if ch >= 'A' && ch <= 'Z' {
			num1 = uint64(ch) - 'A' + 10
		}
		if ch >= 'a' && ch <= 'z' {
			num1 = uint64(ch) - 'a' + 36
		}
		num = num*62 + num1
	}
	return num
}

func (s *ShorterService) retryIfCollision(link string, count int) string {
	newShortLink := encodeBase62(getHash(link + strconv.Itoa(count)))
	return newShortLink
}
