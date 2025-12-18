# Flatten Go Doc

`flatten-go-doc` is a CLI tool that scrapes Go package documentation from `pkg.go.dev` and consolidates it into a single, easy-to-read Markdown file. Perfect for offline reading or providing context to LLMs.

## Features

-   **GitHub URL Support**: Accepts GitHub repository URLs directly (e.g., `https://github.com/user/repo`) and resolves them to `pkg.go.dev`.
-   **Smart Defaults**: Automatically generates a sanitized output filename from the package name if not provided.
-   **Recursive Scraping**: Automatically finds and scrapes sub-packages.
-   **Markdown Conversion**: Converts documentation and READMEs into clean Markdown.
-   **Sorted Output**: Organizes documentation alphabetically by package URL.
-   **Library First**: Built as a reusable Go library with a CLI wrapper.

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

## License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.
