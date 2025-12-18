# Flatten Go Doc

`flatten-go-doc` is a CLI tool that scrapes Go package documentation from `pkg.go.dev` and consolidates it into a single, easy-to-read Markdown file. Perfect for offline reading or providing context to LLMs.

## Features

-   **GitHub URL Support**: Accepts GitHub repository URLs directly (e.g., `https://github.com/user/repo`) and resolves them to `pkg.go.dev`.
-   **Source File Downloading**: Automatically identifies and downloads raw Go source files linked in the documentation, appending them to the output for full context.
-   **Retry Logic**: Robust scraping with automatic retries and exponential backoff for transient network issues.
-   **Recursive Scraping**: Automatically finds and scrapes sub-packages.
-   **Markdown Conversion**: Converts documentation and READMEs into clean Markdown.
-   **Sorted Output**: Organizes documentation alphabetically by package URL.
-   **SOLID Architecture**: Refactored into modular components (Scraper, Converter, Source Handler) for better maintainability and extensibility.

## Installation

### Using `go install`

You can install `flatten-go-doc` directly to your `$GOPATH/bin`:

```bash
go install github.com/igorrius/flatten-go-doc@latest
```

### From Source

1.  Clone the repository:
    ```bash
    git clone https://github.com/igorrius/flatten-go-doc.git
    cd flatten-go-doc
    ```
2.  Install locally:
    ```bash
    go install .
    ```

## Usage

The tool takes the `pkg.go.dev` URL or a GitHub repository URL as the first argument. An optional output file path can be provided as the second argument.

```bash
flatten-go-doc <PKG_GO_DEV_URL | GITHUB_URL> [OUTPUT_FILE]
```

### Examples

**Basic usage (outputs to `<package_name_sanitized>.md`):**
```bash
go run . https://pkg.go.dev/github.com/google/go-cmp/cmp
# Creates: github.com_google_go-cmp_cmp.md
```

**Using a GitHub URL:**
```bash
go run . https://github.com/google/go-cmp
# Resolves to pkg.go.dev and creates: github.com_google_go-cmp.md
```

**Specifying an output file:**
```bash
flatten-doc https://pkg.go.dev/github.com/google/go-cmp/cmp my-docs.md
```

## Library Usage

You can also use the scraper as a library in your own Go projects:

```go
import "github.com/igorrius/flatten-go-doc/pkg/flattener"

func main() {
    config := flattener.DefaultConfig()
    f := flattener.New(config)
    results, err := f.Flatten("https://pkg.go.dev/some/package")
    // ... handle results
}
```

## Tip: Using with NotebookLM

`flatten-go-doc` is an excellent companion for Google **NotebookLM**. By consolidating documentation and source code into a single Markdown file, you can:

1.  **Full Library Context**: Upload the generated `.md` file as a source in NotebookLM.
2.  **AI-Powered Guidance**: Ask NotebookLM to "explain how to use this library based on the source code" or "create a tutorial for [specific feature]".
3.  **Code Synthesis**: Since the tool includes the actual `.go` source files, NotebookLM can reason about the implementation details that aren't always visible in the standard documentation, helping you write more accurate and idiomatic code.

## License

This project is licensed under the Apache License 2.0 - see the [LICENSE](LICENSE) file for details.
