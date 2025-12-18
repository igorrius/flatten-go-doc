package flattener

import (
	"fmt"
	"sort"
	"strings"
	"sync"
	"time"

	md "github.com/JohannesKaufmann/html-to-markdown"
	"github.com/gocolly/colly/v2"
)

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
}

// DefaultConfig returns a default configuration.
func DefaultConfig() Config {
	return Config{
		UserAgent:      "Mozilla/5.0 (compatible; GeminiBot/1.0; +http://gemini.google.com)",
		Parallelism:    2,
		RandomDelay:    500 * time.Millisecond,
		AllowedDomains: []string{"pkg.go.dev"},
	}
}

// Flattener handles the scraping and conversion of documentation.
type Flattener struct {
	config Config
}

// New creates a new Flattener with the given configuration.
func New(config Config) *Flattener {
	return &Flattener{
		config: config,
	}
}

// Flatten scrapes the given root URL and its sub-packages, returning a sorted list of PageResults.
func (f *Flattener) Flatten(rootURL string) ([]PageResult, error) {
	rootURL = strings.TrimSuffix(rootURL, "/")

	// Prepare the Markdown converter
	converter := md.NewConverter("", true, nil)

	// Safe storage for results
	var results []PageResult
	var mu sync.Mutex

	// Initialize Colly
	c := colly.NewCollector(
		colly.AllowedDomains(f.config.AllowedDomains...),
		colly.UserAgent(f.config.UserAgent),
		colly.Async(true),
	)

	// Configure parallelism and delay
	err := c.Limit(&colly.LimitRule{
		DomainGlob:  "*",
		Parallelism: f.config.Parallelism,
		RandomDelay: f.config.RandomDelay,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to set limit rule: %w", err)
	}

	// 1. Handle Documentation Content
	c.OnHTML("main", func(e *colly.HTMLElement) {
		docURL := e.Request.URL.String()
		
		// Selectors updated to match pkg.go.dev structure
		selection := e.DOM.Find(".UnitReadme, .Documentation")

		// Remove noise
		selection.Find(".Documentation-index").Remove()
		selection.Find("script").Remove()
		selection.Find("style").Remove()

		markdown := converter.Convert(selection)

		// Add a header for the file
		header := fmt.Sprintf("# Package: %s\nInput URL: %s\n\n", docURL, docURL)

		mu.Lock()
		results = append(results, PageResult{
			URL:     docURL,
			Content: header + markdown,
		})
		mu.Unlock()
	})

	// 2. Handle Subdirectories (Children links)
	c.OnHTML("div.UnitDirectories table tbody tr td a", func(e *colly.HTMLElement) {
		link := e.Attr("href")
		absoluteURL := e.Request.AbsoluteURL(link)

		// Only visit if it is a child of the root URL
		if strings.HasPrefix(absoluteURL, rootURL) {
			e.Request.Visit(link)
		}
	})

	// Start the scraping
	if err := c.Visit(rootURL); err != nil {
		return nil, fmt.Errorf("failed to start scraping: %w", err)
	}

	// Wait for all requests to finish
	c.Wait()

	// Sort results by URL length (shortest first -> root) or alphabetically.
	// Here we stick to alphabetical to correspond with original behavior's sort intention,
	// but generally sorting by URL makes sense for a linear read.
	sort.Slice(results, func(i, j int) bool {
		return results[i].URL < results[j].URL
	})

	return results, nil
}
