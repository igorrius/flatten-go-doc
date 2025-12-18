# Flatten Go Doc

## Project Overview

**flatten-go-doc** is a CLI tool written in Go designed to scrape package documentation from `pkg.go.dev` and consolidate it into a single Markdown file. This is particularly useful for creating offline documentation or feeding context into LLMs.

The tool starts at a specified root URL, scrapes the README and main documentation content, and then recursively follows links to subdirectories/sub-packages within that module, appending their content to the output.

## Architecture & Technologies

*   **Language:** Go (1.25.5+)
*   **Scraping Engine:** [Colly](https://github.com/gocolly/colly) (v2) - Handles async crawling, domain restrictions, and politeness delays.
*   **HTML Conversion:** [html-to-markdown](https://github.com/JohannesKaufmann/html-to-markdown) - Converts the scraped HTML documentation into clean Markdown.
*   **Concurrency:** Uses Go routines (via Colly's async mode) with a mutex to safely aggregate results.

*   `main.go`: The CLI application entry point.
*   `pkg/flattener/flattener.go`: The core library containing the `Flattener` struct and scraping logic.
*   `README.md`: General project documentation and usage guide.
*   `go.mod` / `go.sum`: Dependency management.

## Building and Running

### Prerequisites
*   Go installed on your system.

### Installation via `go install`
You can install the tool to your `$GOPATH/bin`:
```bash
go install github.com/igorrius/flatten-go-doc@latest
```

### Running Directly
You can run the tool directly using `go run`:

```bash
go run . <PKG_GO_DEV_URL> [OUTPUT_FILE]
```

**Example:**
```bash
go run . https://pkg.go.dev/github.com/cinar/indicator/v2
```
*   **&lt;PKG_GO_DEV_URL&gt;** (Required): The full URL to the package on `pkg.go.dev`.
*   **[OUTPUT_FILE]** (Optional): The path for the output Markdown file. Defaults to `documentation.md`.

### Building
To create a standalone binary:

```bash
go build -o flatten-go-doc .
./flatten-go-doc https://pkg.go.dev/github.com/some/package
```

## Logic Flow

1.  **Input:** Takes a `pkg.go.dev` URL.
2.  **Scrape:** 
    *   Visits the URL.
    *   Extracts content from `.UnitReadme` and `.Documentation`.
    *   Converts HTML to Markdown.
3.  **Crawl:** Finds links in `.UnitDirectories` (sub-packages) and visits them if they share the root prefix.
4.  **Output:** 
    *   Waits for all requests to complete.
    *   Sorts the pages alphabetically by URL.
    *   Writes all concatenated content to the specified output file, separated by `---`.

## Development Conventions

*   **Formatting:** Standard `go fmt`.
*   **Error Handling:** Simple logging via `log.Fatal` for critical errors.
*   **Politeness:** The scraper is configured with a 500ms random delay and limited parallelism (2 threads) to respect `pkg.go.dev`.
