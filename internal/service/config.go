package service

type Config struct {
	BaseURL         string
	ShortLinkLength int
	MaxRetries      int
}

func DefaultConfig() Config {
	return Config{
		BaseURL:         "http://localhost:8080",
		ShortLinkLength: 6,
		MaxRetries:      3,
	}
}
