package service

import (
	"fmt"
	"net/url"
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

func (s *ShorterService) ShortenLink(link string) string {
	shortLink := encodeBase62(getHash(link))
	if v, ok := s.mapStorage.ValuePresent(link); ok {
		return v
	}
	count := 0
	for s.mapStorage.KeyPresent(shortLink) {
		shortLink = s.retryIfCollision(link, count)
	}
	s.mapStorage.PutLink(shortLink, link)
	return "http://localhost:8080/" + shortLink
}

func (s *ShorterService) GetOriginalLink(shortLink string) (string, error) {
	return s.mapStorage.GetLink(shortLink)
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

		//if char >= '0' && char <= '9' {
		//	value = uint64(char - '0')
		//} else if char >= 'A' && char <= 'Z' {
		//	value = uint64(char - 'A' + 10)
		//} else if char >= 'a' && char <= 'z' {
		//	value = uint64(char - 'a' + 36)
		//}

		//var num1 uint64 = 0
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

func ValidateLink(link string) error {
	_, err := url.Parse(link)
	if err != nil {
		return fmt.Errorf("parsing url: %w", err)
	}

	return nil
}

func (s *ShorterService) retryIfCollision(link string, count int) string {
	newShortLink := encodeBase62(getHash(link + strconv.Itoa(count)))
	return newShortLink
}
