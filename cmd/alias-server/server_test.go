package main

import (
	"bytes"
	"log"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestHandlerRedirectsMatchingLink(t *testing.T) {
	t.Parallel()

	req := httptest.NewRequest(http.MethodGet, "/go", nil)
	rec := httptest.NewRecorder()

	newHandler(map[string]string{"go": "https://go.dev/"}, nil).ServeHTTP(rec, req)

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

			newHandler(map[string]string{"go": "https://go.dev/"}, nil).ServeHTTP(rec, req)

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

	newHandler(map[string]string{"go": "https://go.dev/"}, nil).ServeHTTP(rec, req)

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

	newHandler(map[string]string{"go": "https://go.dev/"}, nil).ServeHTTP(rec, req)

	if rec.Code != http.StatusMethodNotAllowed {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusMethodNotAllowed)
	}
	if got := rec.Header().Get("Allow"); got != "GET, HEAD" {
		t.Fatalf("Allow = %q, want GET, HEAD", got)
	}
}

func TestHandlerRendersLinkIndex(t *testing.T) {
	t.Parallel()

	req := httptest.NewRequest(http.MethodGet, "/index", nil)
	rec := httptest.NewRecorder()

	newHandler(map[string]string{
		"pkg": "https://pkg.go.dev/",
		"go":  "https://go.dev/",
	}, stringPtr("index")).ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusOK)
	}
	if got := rec.Header().Get("Content-Type"); got != "text/html; charset=utf-8" {
		t.Fatalf("Content-Type = %q, want text/html; charset=utf-8", got)
	}

	got := rec.Body.String()
	for _, want := range []string{
		"<h1>Links</h1>",
		`<a href="/go">go</a>`,
		`<a href="https://go.dev/">https://go.dev/</a>`,
		`<a href="/pkg">pkg</a>`,
		`<a href="https://pkg.go.dev/">https://pkg.go.dev/</a>`,
	} {
		if !strings.Contains(got, want) {
			t.Fatalf("body = %q, want to contain %q", got, want)
		}
	}
	assertInOrder(t, got, `<a href="/go">go</a>`, `<a href="/pkg">pkg</a>`)
}

func TestHandlerRendersLinkIndexHead(t *testing.T) {
	t.Parallel()

	req := httptest.NewRequest(http.MethodHead, "/index", nil)
	rec := httptest.NewRecorder()

	newHandler(map[string]string{"go": "https://go.dev/"}, stringPtr("index")).ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusOK)
	}
	if got := rec.Header().Get("Content-Type"); got != "text/html; charset=utf-8" {
		t.Fatalf("Content-Type = %q, want text/html; charset=utf-8", got)
	}
	if rec.Body.Len() != 0 {
		t.Fatalf("body = %q, want empty", rec.Body.String())
	}
}

func TestHandlerRejectsIndexWhenDisabled(t *testing.T) {
	t.Parallel()

	req := httptest.NewRequest(http.MethodGet, "/index", nil)
	rec := httptest.NewRecorder()

	newHandler(map[string]string{"go": "https://go.dev/"}, nil).ServeHTTP(rec, req)

	if rec.Code != http.StatusNotFound {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusNotFound)
	}
}

func TestHandlerRejectsUnsupportedIndexPaths(t *testing.T) {
	t.Parallel()

	tests := []string{
		"/index?x=1",
		"/index/a",
		"/%69%6e%64%65%78",
	}

	for _, target := range tests {
		target := target
		t.Run(target, func(t *testing.T) {
			t.Parallel()

			req := httptest.NewRequest(http.MethodGet, target, nil)
			rec := httptest.NewRecorder()

			newHandler(map[string]string{"go": "https://go.dev/"}, stringPtr("index")).ServeHTTP(rec, req)

			if rec.Code != http.StatusNotFound {
				t.Fatalf("status = %d, want %d", rec.Code, http.StatusNotFound)
			}
		})
	}
}

func TestRequestLoggingRedirect(t *testing.T) {
	t.Parallel()

	var out bytes.Buffer
	logger := log.New(&out, "", log.LstdFlags)
	req := httptest.NewRequest(http.MethodGet, "/go", nil)
	rec := httptest.NewRecorder()

	withRequestLogging(newHandler(map[string]string{"go": "https://go.dev/"}, nil), logger).ServeHTTP(rec, req)

	got := out.String()
	for _, want := range []string{
		"GET /go 302 ",
		" -> https://go.dev/",
	} {
		if !strings.Contains(got, want) {
			t.Fatalf("request log = %q, want to contain %q", got, want)
		}
	}
}

func TestRequestLoggingNotFound(t *testing.T) {
	t.Parallel()

	var out bytes.Buffer
	logger := log.New(&out, "", log.LstdFlags)
	req := httptest.NewRequest(http.MethodGet, "/unknown", nil)
	rec := httptest.NewRecorder()

	withRequestLogging(newHandler(map[string]string{"go": "https://go.dev/"}, nil), logger).ServeHTTP(rec, req)

	got := out.String()
	if !strings.Contains(got, "GET /unknown 404 ") {
		t.Fatalf("request log = %q, want 404 log", got)
	}
	if strings.Contains(got, " -> ") {
		t.Fatalf("request log = %q, want no redirect target", got)
	}
}

func TestRequestLoggingMethodNotAllowed(t *testing.T) {
	t.Parallel()

	var out bytes.Buffer
	logger := log.New(&out, "", log.LstdFlags)
	req := httptest.NewRequest(http.MethodPost, "/go", nil)
	rec := httptest.NewRecorder()

	withRequestLogging(newHandler(map[string]string{"go": "https://go.dev/"}, nil), logger).ServeHTTP(rec, req)

	got := out.String()
	if !strings.Contains(got, "POST /go 405 ") {
		t.Fatalf("request log = %q, want 405 log", got)
	}
}

func TestRequestLoggingLinkIndex(t *testing.T) {
	t.Parallel()

	var out bytes.Buffer
	logger := log.New(&out, "", log.LstdFlags)
	req := httptest.NewRequest(http.MethodGet, "/index", nil)
	rec := httptest.NewRecorder()

	withRequestLogging(newHandler(map[string]string{"go": "https://go.dev/"}, stringPtr("index")), logger).ServeHTTP(rec, req)

	got := out.String()
	if !strings.Contains(got, "GET /index 200 ") {
		t.Fatalf("request log = %q, want 200 log", got)
	}
	if strings.Contains(got, " -> ") {
		t.Fatalf("request log = %q, want no redirect target", got)
	}
}
