package main

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestInterpretRequestURL(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name         string
		target       string
		wantPath     string
		wantRawPath  string
		wantRawQuery string
		wantLinkName string
		wantValid    bool
	}{
		{
			name:         "link",
			target:       "/go",
			wantPath:     "/go",
			wantLinkName: "go",
			wantValid:    true,
		},
		{
			name:         "query",
			target:       "/go?x=1",
			wantPath:     "/go",
			wantRawQuery: "x=1",
		},
		{
			name:     "nested path",
			target:   "/a/b",
			wantPath: "/a/b",
		},
		{
			name:        "encoded path",
			target:      "/%67%6f",
			wantPath:    "/go",
			wantRawPath: "/%67%6f",
		},
		{
			name:     "root",
			target:   "/",
			wantPath: "/",
		},
		{
			name:     "invalid link name",
			target:   "/Go",
			wantPath: "/Go",
		},
	}

	for _, test := range tests {
		test := test
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			req := httptest.NewRequest(http.MethodGet, test.target, nil)

			got := interpretRequestURL(req.URL)

			if got.path != test.wantPath {
				t.Fatalf("path = %q, want %q", got.path, test.wantPath)
			}
			if got.rawPath != test.wantRawPath {
				t.Fatalf("rawPath = %q, want %q", got.rawPath, test.wantRawPath)
			}
			if got.rawQuery != test.wantRawQuery {
				t.Fatalf("rawQuery = %q, want %q", got.rawQuery, test.wantRawQuery)
			}
			if got.linkName != test.wantLinkName {
				t.Fatalf("linkName = %q, want %q", got.linkName, test.wantLinkName)
			}
			if got.valid != test.wantValid {
				t.Fatalf("valid = %v, want %v", got.valid, test.wantValid)
			}
		})
	}
}
