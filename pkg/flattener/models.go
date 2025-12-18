package flattener

import "time"

// PageResult represents a single scraped page's documentation.
type PageResult struct {
	URL     string
	Content string
}

// Config holds configuration for the Flattener.
type Config struct {
	UserAgent      string
	Parallelism    int
	RandomDelay    time.Duration
	AllowedDomains []string
	MaxRetries     int
}

// DefaultConfig returns a default configuration.
func DefaultConfig() Config {
	return Config{
		UserAgent:      "Mozilla/5.0 (compatible; GeminiBot/1.0; +http://gemini.google.com)",
		Parallelism:    2,
		RandomDelay:    500 * time.Millisecond,
		AllowedDomains: []string{"pkg.go.dev"},
		MaxRetries:     3,
	}
}
