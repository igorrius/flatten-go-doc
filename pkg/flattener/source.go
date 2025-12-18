package flattener

import (
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
)

// IsSourceLink checks if a link points to a source file.
func IsSourceLink(url string) bool {
	return strings.HasSuffix(url, ".go")
}

// GetRawURL converts a VCS provider URL to a raw content URL if possible.
func GetRawURL(originalURL string) string {
	if strings.Contains(originalURL, "github.com") && strings.Contains(originalURL, "/blob/") {
		// Convert https://github.com/user/repo/blob/branch/path/to/file.go
		// To https://raw.githubusercontent.com/user/repo/branch/path/to/file.go
		rawURL := strings.Replace(originalURL, "github.com", "raw.githubusercontent.com", 1)
		rawURL = strings.Replace(rawURL, "/blob/", "/", 1)
		return rawURL
	}
	// Add other providers if needed (e.g. gitlab, bitbucket)
	return originalURL
}

// DownloadSource downloads the raw content of a source file with retries.
func DownloadSource(url string, maxRetries int) (string, error) {
	rawURL := GetRawURL(url)
	
	var lastErr error
	for i := 0; i <= maxRetries; i++ {
		if i > 0 {
			time.Sleep(time.Duration(i) * 500 * time.Millisecond) // Exponential-ish backoff
		}

		resp, err := http.Get(rawURL)
		if err != nil {
			lastErr = err
			continue
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			lastErr = fmt.Errorf("unexpected status code: %d", resp.StatusCode)
			continue
		}

		body, err := io.ReadAll(resp.Body)
		if err != nil {
			lastErr = err
			continue
		}

		return string(body), nil
	}

	return "", fmt.Errorf("failed after %d retries: %w", maxRetries, lastErr)
}