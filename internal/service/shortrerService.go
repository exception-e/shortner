package service

import (
	"fmt"
	"log/slog"
	"strconv"

	storageTypes "shortner/internal/storage/types"

	"github.com/spaolacci/murmur3"
)

const base62Alphabet = "0123456789ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz"

type ShortnerService struct {
	mapStorage storageTypes.LinkStorage
	logger     *slog.Logger
}

func NewShortnerService(mapStorage storageTypes.LinkStorage, logger *slog.Logger) *ShortnerService {
	componentLogger := logger.With(slog.String("component", "shortnerService"))
	return &ShortnerService{mapStorage: mapStorage, logger: componentLogger}
}

func (s *ShortnerService) ShortenLink(link string) (string, error) {
	s.logger.Info("Shortening link", slog.String("link", link))
	if v, ok := s.mapStorage.ValuePresent(link); ok {
		return "http://localhost:8080/" + v, nil
	}

	alias := encodeBase62(getHash(link))
	count := 0
	for s.mapStorage.KeyPresent(alias) {
		if count > 50 {
			s.logger.Error("Failed to generate unique short link")
			return "", fmt.Errorf("service: failed to generate unique alias for %s after %d attempts", link, count)
		}
		alias = s.addSalt(link, count)
		count++
	}
	_, err := s.mapStorage.PutLink(alias, link)
	if err != nil {
		return "", fmt.Errorf("service failed to save url %s: %w", link, err)
	}
	s.logger.Info("Link shortened and saved", slog.String("alias", alias))
	return "http://localhost:8080/" + alias, nil
}

func (s *ShortnerService) GetOriginalLink(alias string) (string, error) {
	s.logger.Info("Getting original link", slog.String("shortLink", alias))
	link, err := s.mapStorage.GetLink(alias)
	if err != nil {
		return "", fmt.Errorf("service: failed to get original url for alias %s: %w", alias, err)
	}
	s.logger.Info("Original link", slog.String("link", link))
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

func (s *ShortnerService) addSalt(link string, count int) string {
	newShortLink := encodeBase62(getHash(link + strconv.Itoa(count)))
	return newShortLink
}
