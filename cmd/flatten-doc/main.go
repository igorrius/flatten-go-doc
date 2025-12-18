package main

import (
	"flag"
	"fmt"
	"log"
	"os"

	"flatten-go-doc/pkg/flattener"
)

func main() {
	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage: %s [options] <url> [output-file]\n", os.Args[0])
		fmt.Fprintf(os.Stderr, "Arguments:\n")
		fmt.Fprintf(os.Stderr, "  <url>          The URL of the Go package documentation (e.g., https://pkg.go.dev/github.com/cinar/indicator/v2)\n")
		fmt.Fprintf(os.Stderr, "  [output-file]  Output markdown file path (default: documentation.md)\n")
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

	url := args[0]
	outFile := "documentation.md"
	if len(args) >= 2 {
		outFile = args[1]
	}

	fmt.Printf("Starting scraping for: %s\n", url)

	// Configure the flattener
	config := flattener.DefaultConfig()
	// You could expose config options via flags here (e.g., parallelism) if desired.

	f := flattener.New(config)
	results, err := f.Flatten(url)
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
