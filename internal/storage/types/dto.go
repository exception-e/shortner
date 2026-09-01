package types

import "time"

type CreateLinkDTO struct {
	Alias       string
	OriginalURL string
}

type LinkDTO struct {
	ID          int64
	Alias       string
	OriginalURL string
	CreatedAt   time.Time
}
