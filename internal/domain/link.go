package domain

import (
	"fmt"
	"net/url"
	"time"
)

type Link struct {
	Id          int64
	Alias       string
	OriginalUrl string
	CreatedAt   time.Time
}

func NewLink(originalURL string, shortCode string) (*Link, error) {
	if originalURL == "" {
		return nil, fmt.Errorf("domain: original URL cannot be empty")
	}
	if shortCode == "" {
		return nil, fmt.Errorf("domain: short code cannot be empty")
	}

	if _, err := url.Parse(originalURL); err != nil {
		return nil, fmt.Errorf("domain: invalid URL: %w", err)
	}

	return &Link{
		Alias:       shortCode,
		OriginalUrl: originalURL,
	}, nil
}

//func (l *Link) IsExpired(ttl time.Duration) bool {
//	return time.Since(l.CreatedAt) > ttl
//}
