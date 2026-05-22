package main

import (
	"net/http"
	"strings"
)

func newHandler(links map[string]string) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet && r.Method != http.MethodHead {
			w.Header().Set("Allow", "GET, HEAD")
			http.Error(w, http.StatusText(http.StatusMethodNotAllowed), http.StatusMethodNotAllowed)
			return
		}

		name, ok := linkNameFromRequest(r)
		if !ok {
			http.NotFound(w, r)
			return
		}

		target, ok := links[name]
		if !ok {
			http.NotFound(w, r)
			return
		}

		http.Redirect(w, r, target, http.StatusFound)
	})
}

func linkNameFromRequest(r *http.Request) (string, bool) {
	if r.URL.RawQuery != "" {
		return "", false
	}
	if r.URL.RawPath != "" {
		return "", false
	}
	if !strings.HasPrefix(r.URL.Path, "/") {
		return "", false
	}

	name := strings.TrimPrefix(r.URL.Path, "/")
	if !isValidLinkName(name) {
		return "", false
	}

	return name, true
}
