package flattener

import (
	"fmt"
	"sort"
	"strings"
	"sync"

	"github.com/PuerkitoBio/goquery"
	"github.com/gocolly/colly/v2"
)

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

	// Components
	scraper, err := NewScraper(f.config)
	if err != nil {
		return nil, err
	}
	converter := NewConverter()

	// Safe storage for results
	var results []PageResult
	var mu sync.Mutex
	var wg sync.WaitGroup
	visitedSources := make(map[string]bool)

	// Handler for Main Content
	scraper.Collector.OnHTML("main", func(e *colly.HTMLElement) {
		docURL := e.Request.URL.String()

		// 1. Documentation & Readme
		docSelection := e.DOM.Find(".UnitReadme, .Documentation-content")
		if docSelection.Length() > 0 {
			// Cleanup
			docSelection.Find(".Documentation-index").Remove()
			docSelection.Find("script").Remove()
			docSelection.Find("style").Remove()
			docSelection.Find(".Documentation-exampleButtonsContainer").Remove()

			markdown := converter.Convert(docSelection)

			// Add a header for the file
			header := fmt.Sprintf("# Package: %s\nInput URL: %s\n\n", docURL, docURL)

			mu.Lock()
			results = append(results, PageResult{
				URL:     docURL,
				Content: header + markdown,
			})
			mu.Unlock()

			// Find Source File Links
			e.DOM.Find(".UnitFiles-fileList li a").Each(func(_ int, s *goquery.Selection) {
				href, exists := s.Attr("href")
				if exists {
					fullURL := e.Request.AbsoluteURL(href)
					
					mu.Lock()
					if visited := visitedSources[fullURL]; !visited && IsSourceLink(fullURL) {
						visitedSources[fullURL] = true
						mu.Unlock()

						// Download source file
						// We do this in a goroutine to not block the collector
						wg.Add(1)
						go func(u string) {
							defer wg.Done()
							content, err := DownloadSource(u, f.config.MaxRetries)
							if err == nil {
								header := fmt.Sprintf("# Source File: %s\nURL: %s\n\n```go\n", u, u)
								footer := "\n```\n"
								
								mu.Lock()
								results = append(results, PageResult{
									URL:     u,
									Content: header + content + footer,
								})
								mu.Unlock()
							}
						}(fullURL)
					} else {
						mu.Unlock()
					}
				}
			})
		}
	})

	// Handle Subdirectories (Children links)
	scraper.Collector.OnHTML("div.UnitDirectories table tr td a", func(e *colly.HTMLElement) {
		link := e.Attr("href")
		absoluteURL := e.Request.AbsoluteURL(link)

		// Only visit if it is a child of the root URL
		if strings.HasPrefix(absoluteURL, rootURL) {
			e.Request.Visit(link)
		}
	})

	// Start the scraping
	if err := scraper.Visit(rootURL); err != nil {
		return nil, fmt.Errorf("failed to start scraping: %w", err)
	}

	// Wait for all requests to finish
	scraper.Wait()
	wg.Wait()

	// Sort results by URL
	sort.Slice(results, func(i, j int) bool {
		return results[i].URL < results[j].URL
	})

	return results, nil
}
