package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"sort"
	"strings"
	"sync"
	"time"

	md "github.com/JohannesKaufmann/html-to-markdown"
	"github.com/gocolly/colly/v2"
)

type PageResult struct {
	URL     string
	Content string
}

func main() {
	urlFlag := flag.String("url", "", "The URL of the Go package documentation (e.g., https://pkg.go.dev/github.com/cinar/indicator/v2)")
	outFlag := flag.String("out", "documentation.md", "Output markdown file path")
	flag.Parse()

	if *urlFlag == "" {
		log.Fatal("Please provide a URL using the -url flag")
	}

	rootURL := strings.TrimSuffix(*urlFlag, "/")
	
	// Prepare the Markdown converter
	converter := md.NewConverter("", true, nil)

	// Safe storage for results
	var results []PageResult
	var mu sync.Mutex

	// Initialize Colly
	c := colly.NewCollector(
		colly.AllowedDomains("pkg.go.dev"),
		colly.UserAgent("Mozilla/5.0 (compatible; GeminiBot/1.0; +http://gemini.google.com)"),
		colly.Async(true),
	)

	// Limit parallelism to be polite
	c.Limit(&colly.LimitRule{
		DomainGlob:  "*",
		Parallelism: 2,
		RandomDelay: 500 * time.Millisecond,
	})

	// 1. Handle Documentation Content
	c.OnHTML("main", func(e *colly.HTMLElement) {
		docURL := e.Request.URL.String()
		fmt.Printf("Scraping: %s\n", docURL)

		// Create a clean DOM for extraction
		// We mainly want the Readme and the Documentation section
		// Selectors updated to match current pkg.go.dev structure (divs mostly)
		selection := e.DOM.Find(".UnitReadme, .Documentation")
		
		// Remove noise
		selection.Find(".Documentation-index").Remove() // Remove the index links, just keep content
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
	// The selector for the directory list in pkg.go.dev
	// Usually found in a "Directories" section. 
	// The markup is often a table within `div.UnitDirectories`.
	c.OnHTML("div.UnitDirectories table tbody tr td a", func(e *colly.HTMLElement) {
		link := e.Attr("href")
		// pkg.go.dev links are often relative like "/github.com/..."
		absoluteURL := e.Request.AbsoluteURL(link)

		// Only visit if it is a child of the root URL (or the module path)
		// and hasn't been visited (Colly handles visited check if configured, but let's be explicit about scope)
		if strings.HasPrefix(absoluteURL, rootURL) {
			// e.Request.Visit(absoluteURL) -> For Async, we use c.Visit, but inside OnHTML we should use e.Request.Visit 
			// However, in Async mode, it's safer to just queue it.
			e.Request.Visit(link)
		}
	})

	// Start the scraping
	err := c.Visit(rootURL)
	if err != nil {
		log.Fatal(err)
	}

	// Wait for all requests to finish
	c.Wait()

	// Sort results by URL length (shortest first -> root) or alphabetically
	sort.Slice(results, func(i, j int) bool {
		return results[i].URL < results[j].URL
	})

	// Write to file
	f, err := os.Create(*outFlag)
	if err != nil {
		log.Fatal(err)
	}
	defer f.Close()

	for _, res := range results {
		_, err := f.WriteString(res.Content)
		if err != nil {
			log.Printf("Error writing content: %v", err)
		}
		_, err = f.WriteString("\n\n---\n\n") // Separator
		if err != nil {
			log.Printf("Error writing separator: %v", err)
		}
	}

	fmt.Printf("Successfully saved documentation to %s\n", *outFlag)
}
