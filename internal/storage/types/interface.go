package types

type LinkStorage interface {
	PutLink(shortLink string, link string) string
	GetLink(shortLink string) (string, error)
	ValuePresent(link string) (string, bool)
	KeyPresent(shortLink string) bool
}
