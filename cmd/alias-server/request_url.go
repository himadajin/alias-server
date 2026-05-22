package main

import (
	"net/url"
	"strings"
)

type requestURL struct {
	path     string
	rawPath  string
	rawQuery string
	linkName string
	valid    bool
}

func interpretRequestURL(raw *url.URL) requestURL {
	interpreted := requestURL{
		path:     raw.Path,
		rawPath:  raw.RawPath,
		rawQuery: raw.RawQuery,
	}

	if interpreted.rawQuery != "" {
		return interpreted
	}
	if interpreted.rawPath != "" {
		return interpreted
	}
	if !strings.HasPrefix(interpreted.path, "/") {
		return interpreted
	}

	name := strings.TrimPrefix(interpreted.path, "/")
	if !isValidLinkName(name) {
		return interpreted
	}

	interpreted.linkName = name
	interpreted.valid = true
	return interpreted
}
