package main

import (
	"flag"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"os"
	"strings"

	"github.com/igorrius/flatten-go-doc/pkg/flattener"
)

func main() {
	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage: %s [options] <url> [output-file]\n", os.Args[0])
		fmt.Fprintf(os.Stderr, "Arguments:\n")
		fmt.Fprintf(os.Stderr, "  <url>          The URL of the Go package documentation (e.g., https://pkg.go.dev/github.com/cinar/indicator/v2)\n")
		fmt.Fprintf(os.Stderr, "                 Or a GitHub repository URL (e.g., https://github.com/cinar/indicator)\n")
		fmt.Fprintf(os.Stderr, "  [output-file]  Output markdown file path (default: <package_name_sanitized>.md)\n")
		fmt.Fprintf(os.Stderr, "Options:\n")
		flag.PrintDefaults()
	}
	flag.Parse()

	args := flag.Args()
	if len(args) < 1 {
		fmt.Println("Error: Please provide a URL as the first argument.")
		flag.Usage()
		os.Exit(1)
	}

	rawURL := args[0]
	targetURL := rawURL
	isGitHub := false

	// Attempt to parse and transform GitHub URLs
	u, err := url.Parse(rawURL)
	if err == nil {
		if strings.HasSuffix(u.Host, "github.com") {
			// e.g. https://github.com/user/repo -> https://pkg.go.dev/github.com/user/repo
			// u.Path includes the leading slash
			targetURL = "https://pkg.go.dev/" + u.Host + u.Path
			isGitHub = true
		} else if u.Scheme == "" && strings.HasPrefix(rawURL, "github.com/") {
			// e.g. github.com/user/repo -> https://pkg.go.dev/github.com/user/repo
			targetURL = "https://pkg.go.dev/" + rawURL
			isGitHub = true
		}
	}

	if isGitHub {
		fmt.Printf("Detected GitHub URL. Transformed to: %s\n", targetURL)
		fmt.Println("Verifying package availability on pkg.go.dev...")

		resp, err := http.Get(targetURL)
		if err != nil {
			fmt.Printf("Error checking package availability: %v\n", err)
			os.Exit(1)
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			fmt.Printf("Error: Package not available on pkg.go.dev (status: %d)\n", resp.StatusCode)
			os.Exit(1)
		}
	}

	outFile := ""
	if len(args) >= 2 {
		outFile = args[1]
	} else {
		// Generate default filename from targetURL
		parsedURL, err := url.Parse(targetURL)
		if err == nil {
			path := strings.TrimPrefix(parsedURL.Path, "/")
			if path == "" {
				path = "documentation"
			}
			// Sanitize the path for use as a filename
			replacer := strings.NewReplacer(
				"/", "_",
				"\\", "_",
				":", "_",
				"*", "_",
				"?", "_",
				"\"", "_",
				"<", "_",
				">", "_",
				"|", "_",
			)
			outFile = replacer.Replace(path)
		} else {
			outFile = "documentation"
		}
	}

	if !strings.HasSuffix(strings.ToLower(outFile), ".md") {
		outFile += ".md"
	}

	fmt.Printf("Starting scraping for: %s\n", targetURL)

	// Configure the flattener
	config := flattener.DefaultConfig()
	// You could expose config options via flags here (e.g., parallelism) if desired.

	f := flattener.New(config)
	results, err := f.Flatten(targetURL)
	if err != nil {
		log.Fatalf("Error flattening documentation: %v", err)
	}

	if len(results) == 0 {
		fmt.Println("Warning: No documentation found.")
		return
	}

	// Write to file
	file, err := os.Create(outFile)
	if err != nil {
		log.Fatalf("Error creating output file: %v", err)
	}
	defer file.Close()

	for _, res := range results {
		fmt.Printf("Writing: %s\n", res.URL)
		_, err := file.WriteString(res.Content)
		if err != nil {
			log.Printf("Error writing content for %s: %v", res.URL, err)
		}
		_, err = file.WriteString("\n\n---\n\n") // Separator
		if err != nil {
			log.Printf("Error writing separator: %v", err)
		}
	}

	fmt.Printf("\nSuccessfully saved documentation to %s\n", outFile)
}
