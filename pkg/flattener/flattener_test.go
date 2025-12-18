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
	}
	
	flattener := New(config)

	// 4. Run Flatten
	targetURL := server.URL + "/pkg"
	results, err := flattener.Flatten(targetURL)
	if err != nil {
		t.Fatalf("Flatten failed: %v", err)
	}

	// 5. Assertions
	if len(results) != 2 {
		t.Errorf("Expected 2 results, got %d", len(results))
	}

	// Helper to find result by URL suffix (since port changes)
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
		if !strings.Contains(rootRes.Content, "This is the root readme.") {
			t.Errorf("Root content missing body. Got:\n%s", rootRes.Content)
		}
	}

	// Check Sub
	subRes := findResult("/pkg/sub")
	if subRes == nil {
		t.Error("Result for subpackage not found")
	} else {
		if !strings.Contains(subRes.Content, "## Subpackage Doc") {
			t.Errorf("Sub content missing header. Got:\n%s", subRes.Content)
		}
		if strings.Contains(subRes.Content, "Index (should be removed)") {
			t.Error("Content cleaning failed: found .Documentation-index content")
		}
	}
}
