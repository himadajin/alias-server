package main

import (
	"bytes"
	"html/template"
	"log"
	"net/http"
	"time"
)

var linkIndexTemplate = template.Must(template.New("linkIndex").Parse(`<!doctype html>
<html lang="en">
<head>
  <meta charset="utf-8">
  <title>Links</title>

  <style>
    body {
      font-family: system-ui, sans-serif;
      line-height: 1.5;
      margin: 2rem;
      max-width: 60rem;
    }

    table {
      border-collapse: collapse;
      width: 100%;
    }

    th,
    td {
      border-bottom: 1px solid #ddd;
      padding: 0.5rem;
      text-align: left;
    }

    a {
      color: #0645ad;
    }
  </style>
</head>
<body>
  <h1>Links</h1>
  <table>
    <thead>
      <tr>
        <th>Link</th>
        <th>Target</th>
      </tr>
    </thead>
    <tbody>
      {{range .Links}}
      <tr>
        <td><a href="/{{.Name}}">{{.Name}}</a></td>
        <td><a href="{{.Target}}">{{.Target}}</a></td>
      </tr>
      {{end}}
    </tbody>
  </table>
</body>
</html>
`))

type linkIndexData struct {
	Links []linkIndexLink
}

type linkIndexLink struct {
	Name   string
	Target string
}

func newHandler(links map[string]string, indexLink *string) http.Handler {
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

		if indexLink != nil && requestURL.linkName == *indexLink {
			serveLinkIndex(w, r, links)
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

func serveLinkIndex(w http.ResponseWriter, r *http.Request, links map[string]string) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	if r.Method == http.MethodHead {
		return
	}

	var body bytes.Buffer
	if err := linkIndexTemplate.Execute(&body, newLinkIndexData(links)); err != nil {
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}
	body.WriteTo(w)
}

func newLinkIndexData(links map[string]string) linkIndexData {
	names := sortedLinkNames(links)
	data := linkIndexData{
		Links: make([]linkIndexLink, 0, len(names)),
	}
	for _, name := range names {
		data.Links = append(data.Links, linkIndexLink{
			Name:   name,
			Target: links[name],
		})
	}
	return data
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
