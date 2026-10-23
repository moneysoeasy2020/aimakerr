package gitlab

import (
	"encoding/json"
	"net/http"
	"testing"
)

func TestRender(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name       string
		options    *RenderOptions
		wantHTML   string
		statusCode int
	}{
		{
			name: "Basic Markdown",
			options: &RenderOptions{
				Text: Ptr("# Testing"),
			},
			wantHTML:   "<h1>Testing</h1>",
			statusCode: http.StatusOK,
		},
		{
			name: "With GFM and project",
			options: &RenderOptions{
				Text:                    Ptr("**bold**"),
				GitlabFlavouredMarkdown: Ptr(true),
				Project:                 Ptr("group/project"),
			},
			wantHTML:   "<p><strong>bold</strong></p>",
			statusCode: http.StatusOK,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			mux, client := setup(t)

			mux.HandleFunc("/api/v4/markdown", func(w http.ResponseWriter, r *http.Request) {
				testMethod(t, r, http.MethodPost)

				var body map[string]any
				if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
					t.Fatalf("Failed to decode request body: %v", err)
				}
				if body["text"] == nil {
					t.Errorf("Expected 'text' field in request body, got nil")
				}

				w.WriteHeader(tc.statusCode)
				_ = json.NewEncoder(w).Encode(Markdown{HTML: tc.wantHTML})
			})

			md, resp, err := client.Markdown.Render(tc.options)
			if err != nil {
				t.Fatalf("Render failed: %v", err)
			}
			if resp.StatusCode != tc.statusCode {
				t.Fatalf("Expected status %d, got %d", tc.statusCode, resp.StatusCode)
			}
			if md == nil || md.HTML != tc.wantHTML {
				t.Fatalf("Expected HTML %q, got %q", tc.wantHTML, md.HTML)
			}
		})
	}
}
