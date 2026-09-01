package types

import "shortner/internal/domain"

// TODO: godocs
// TODO: ctx
type LinkStorage interface {
	PutLink(*domain.Link) (string, error)
	GetLink(shortLink string) (*domain.Link, error)
	// TODO: IsPresent
	ValuePresent(link string) (string, bool)
	KeyPresent(shortLink string) bool
}
