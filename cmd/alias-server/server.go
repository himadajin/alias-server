package main

import (
	"log"
	"net/http"
	"time"
)

func newHandler(links map[string]string) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet && r.Method != http.MethodHead {
			w.Header().Set("Allow", "GET, HEAD")
			http.Error(w, http.StatusText(http.StatusMethodNotAllowed), http.StatusMethodNotAllowed)
			return
		}

		requestURL := interpretRequestURL(r.URL)
		if !requestURL.valid {
			http.NotFound(w, r)
			return
		}

		target, ok := links[requestURL.linkName]
		if !ok {
			http.NotFound(w, r)
			return
		}

		http.Redirect(w, r, target, http.StatusFound)
	})
}

func withRequestLogging(next http.Handler, logger *log.Logger) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		rec := &statusRecorder{
			ResponseWriter: w,
			status:         http.StatusOK,
		}

		next.ServeHTTP(rec, r)

		message := "%s %s %d %s"
		args := []any{
			r.Method,
			r.URL.RequestURI(),
			rec.status,
			time.Since(start).Round(time.Microsecond),
		}
		if location := rec.Header().Get("Location"); location != "" {
			message += " -> %s"
			args = append(args, location)
		}
		logger.Printf(message, args...)
	})
}

type statusRecorder struct {
	http.ResponseWriter
	status int
}

func (r *statusRecorder) WriteHeader(status int) {
	r.status = status
	r.ResponseWriter.WriteHeader(status)
}
