package main

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestHandlerRedirectsMatchingLink(t *testing.T) {
	t.Parallel()

	req := httptest.NewRequest(http.MethodGet, "/go", nil)
	rec := httptest.NewRecorder()

	newHandler(map[string]string{"go": "https://go.dev/"}).ServeHTTP(rec, req)

	if rec.Code != http.StatusFound {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusFound)
	}
	if got := rec.Header().Get("Location"); got != "https://go.dev/" {
		t.Fatalf("Location = %q, want https://go.dev/", got)
	}
}

func TestHandlerRejectsUnsupportedPaths(t *testing.T) {
	t.Parallel()

	tests := []string{
		"/go?x=1",
		"/a/b",
		"/%67%6f",
		"/unknown",
		"/",
	}

	for _, target := range tests {
		target := target
		t.Run(target, func(t *testing.T) {
			t.Parallel()

			req := httptest.NewRequest(http.MethodGet, target, nil)
			rec := httptest.NewRecorder()

			newHandler(map[string]string{"go": "https://go.dev/"}).ServeHTTP(rec, req)

			if rec.Code != http.StatusNotFound {
				t.Fatalf("status = %d, want %d", rec.Code, http.StatusNotFound)
			}
		})
	}
}

func TestHandlerRedirectsHead(t *testing.T) {
	t.Parallel()

	req := httptest.NewRequest(http.MethodHead, "/go", nil)
	rec := httptest.NewRecorder()

	newHandler(map[string]string{"go": "https://go.dev/"}).ServeHTTP(rec, req)

	if rec.Code != http.StatusFound {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusFound)
	}
	if got := rec.Header().Get("Location"); got != "https://go.dev/" {
		t.Fatalf("Location = %q, want https://go.dev/", got)
	}
}

func TestHandlerRejectsUnsupportedMethod(t *testing.T) {
	t.Parallel()

	req := httptest.NewRequest(http.MethodPost, "/go", nil)
	rec := httptest.NewRecorder()

	newHandler(map[string]string{"go": "https://go.dev/"}).ServeHTTP(rec, req)

	if rec.Code != http.StatusMethodNotAllowed {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusMethodNotAllowed)
	}
	if got := rec.Header().Get("Allow"); got != "GET, HEAD" {
		t.Fatalf("Allow = %q, want GET, HEAD", got)
	}
}
