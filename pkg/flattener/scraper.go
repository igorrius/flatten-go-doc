package flattener

import (
	"fmt"
	"log"

	"github.com/gocolly/colly/v2"
)

// Scraper wraps colly.Collector with custom configuration and retry logic.
type Scraper struct {
	Collector *colly.Collector
	config    Config
}

// NewScraper creates a new Scraper instance.
func NewScraper(config Config) (*Scraper, error) {
	c := colly.NewCollector(
		colly.AllowedDomains(config.AllowedDomains...),
		colly.UserAgent(config.UserAgent),
		colly.Async(true),
	)

	// Configure parallelism and delay
	err := c.Limit(&colly.LimitRule{
		DomainGlob:  "*",
		Parallelism: config.Parallelism,
		RandomDelay: config.RandomDelay,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to set limit rule: %w", err)
	}

	scraper := &Scraper{
		Collector: c,
		config:    config,
	}

	scraper.setupRetry()

	return scraper, nil
}

func (s *Scraper) setupRetry() {
	s.Collector.OnError(func(r *colly.Response, err error) {
		// StatusCode is 0 if the request failed before receiving a response
		if r.StatusCode >= 400 || r.StatusCode == 0 {
			retries := 0
			if v := r.Ctx.GetAny("retries"); v != nil {
				retries = v.(int)
			}

			if retries < s.config.MaxRetries {
				log.Printf("Error requesting %s (attempt %d/%d): %v. Retrying...", r.Request.URL, retries+1, s.config.MaxRetries, err)
				r.Ctx.Put("retries", retries+1)
				r.Request.Retry()
			} else {
				log.Printf("Error requesting %s: %v. Max retries reached.", r.Request.URL, err)
			}
		}
	})
}

// Visit starts the scraping process.
func (s *Scraper) Visit(url string) error {
	return s.Collector.Visit(url)
}

// Wait waits for all requests to finish.
func (s *Scraper) Wait() {
	s.Collector.Wait()
}
