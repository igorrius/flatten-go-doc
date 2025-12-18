package flattener

import (
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"
)

func TestFlatten(t *testing.T) {
	// 1. Setup Mock Server
	mux := http.NewServeMux()
	
	// Root package page
	mux.HandleFunc("/pkg", func(w http.ResponseWriter, r *http.Request) {
		html := `
		<html>
			<body>
				<main>
				<div class="UnitReadme">
					<h1>Root Package</h1>
					<p>This is the root readme.</p>
				</div>
				
				<div class="UnitFiles js-unitFiles">
					<ul class="UnitFiles-fileList">
						<li><a href="/pkg/file.go">file.go</a></li>
					</ul>
				</div>

				<div class="UnitDirectories">
					<table>
						<tr>
							<td><a href="/pkg/sub">Subpackage</a></td>
						</tr>
					</table>
				</div>
				</main>
			</body>
		</html>
		`
		w.Write([]byte(html))
	})

	// Sub package page
	mux.HandleFunc("/pkg/sub", func(w http.ResponseWriter, r *http.Request) {
		html := `
		<html>
			<body>
				<main>
				<div class="Documentation-content">
					<h2>Subpackage Doc</h2>
					<p>This is the subpackage documentation.</p>
					<div class="Documentation-index">Index (should be removed)</div>
				</div>
				</main>
			</body>
		</html>
		`
		w.Write([]byte(html))
	})

	// Source file page (RAW)
	mux.HandleFunc("/pkg/file.go", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("package pkg\n\nfunc Foo() {}"))
	})

	server := httptest.NewServer(mux)
	defer server.Close()

	// 2. Parse Server URL to get host for AllowedDomains
	u, err := url.Parse(server.URL)
	if err != nil {
		t.Fatalf("Failed to parse server URL: %v", err)
	}

	// 3. Configure Flattener
	config := Config{
		UserAgent:      "TestAgent",
		Parallelism:    1,
		RandomDelay:    0, // No delay for tests
		AllowedDomains: []string{u.Hostname()},
		MaxRetries:     1,
	}
	
	flattener := New(config)

	// 4. Run Flatten
	targetURL := server.URL + "/pkg"
	results, err := flattener.Flatten(targetURL)
	if err != nil {
		t.Fatalf("Flatten failed: %v", err)
	}

	// 5. Assertions
	// Expected: Root, Sub, SourceFile -> 3 results
	if len(results) != 3 {
		t.Errorf("Expected 3 results, got %d", len(results))
		for _, r := range results {
			t.Logf("Result: %s", r.URL)
		}
	}

	// Helper to find result by URL suffix
	findResult := func(suffix string) *PageResult {
		for _, r := range results {
			if strings.HasSuffix(r.URL, suffix) {
				return &r
			}
		}
		return nil
	}

	// Check Root
	rootRes := findResult("/pkg")
	if rootRes == nil {
		t.Error("Result for root package not found")
	} else {
		if !strings.Contains(rootRes.Content, "# Root Package") {
			t.Errorf("Root content missing header. Got:\n%s", rootRes.Content)
		}
	}

	// Check Sub
	subRes := findResult("/pkg/sub")
	if subRes == nil {
		t.Error("Result for subpackage not found")
	}

	// Check Source File
	sourceRes := findResult("/pkg/file.go")
	if sourceRes == nil {
		t.Error("Result for source file not found")
	} else {
		if !strings.Contains(sourceRes.Content, "func Foo() {}") {
			t.Errorf("Source content missing code. Got:\n%s", sourceRes.Content)
		}
		if !strings.Contains(sourceRes.Content, "```go") {
			t.Error("Source content missing code block formatting")
		}
	}
}