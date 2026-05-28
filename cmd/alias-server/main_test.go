package main

import (
	"bytes"
	"errors"
	"log"
	"strings"
	"testing"
)

func TestParseCLIArgsUsesDefaultPort(t *testing.T) {
	t.Parallel()

	file := writeTempConfig(t, `{
		"defaultPort": 15555,
		"links": {
			"go": "https://go.dev/"
		}
	}`)

	options, err := parseCLIArgs([]string{file})
	if err != nil {
		t.Fatalf("parseCLIArgs() error = %v", err)
	}
	if options.port != 15555 {
		t.Fatalf("options.port = %d, want 15555", options.port)
	}
	if options.indexLink != nil {
		t.Fatalf("options.indexLink = %q, want nil", *options.indexLink)
	}
}

func TestParseCLIArgsUsesIndexLink(t *testing.T) {
	t.Parallel()

	file := writeTempConfig(t, `{
		"defaultPort": 15555,
		"indexLink": "index",
		"links": {
			"go": "https://go.dev/"
		}
	}`)

	options, err := parseCLIArgs([]string{file})
	if err != nil {
		t.Fatalf("parseCLIArgs() error = %v", err)
	}
	if options.indexLink == nil {
		t.Fatal("options.indexLink = nil, want index")
	}
	if *options.indexLink != "index" {
		t.Fatalf("*options.indexLink = %q, want index", *options.indexLink)
	}
}

func TestParseCLIArgsCLIOverrideDefaultPort(t *testing.T) {
	t.Parallel()

	file := writeTempConfig(t, `{
		"defaultPort": 15555,
		"links": {
			"go": "https://go.dev/"
		}
	}`)

	options, err := parseCLIArgs([]string{"-port", "15556", file})
	if err != nil {
		t.Fatalf("parseCLIArgs() error = %v", err)
	}
	if options.port != 15556 {
		t.Fatalf("options.port = %d, want 15556", options.port)
	}
}

func TestParseCLIArgsAllowsPortWithOmittedDefaultPort(t *testing.T) {
	t.Parallel()

	file := writeTempConfig(t, `{
		"links": {
			"go": "https://go.dev/"
		}
	}`)

	options, err := parseCLIArgs([]string{"-port", "15556", file})
	if err != nil {
		t.Fatalf("parseCLIArgs() error = %v", err)
	}
	if options.port != 15556 {
		t.Fatalf("options.port = %d, want 15556", options.port)
	}
}

func TestParseCLIArgsRejectsInvalidArgs(t *testing.T) {
	t.Parallel()

	file := writeTempConfig(t, `{
		"defaultPort": 15555,
		"links": {
			"go": "https://go.dev/"
		}
	}`)
	fileWithoutDefaultPort := writeTempConfig(t, `{
		"links": {
			"go": "https://go.dev/"
		}
	}`)

	tests := []struct {
		name string
		args []string
	}{
		{name: "missing config file", args: []string{}},
		{name: "multiple config files", args: []string{file, file}},
		{name: "missing cli and default port", args: []string{fileWithoutDefaultPort}},
		{name: "cli port zero", args: []string{"-port", "0", file}},
		{name: "cli port negative", args: []string{"-port", "-1", file}},
		{name: "cli port too high", args: []string{"-port", "65536", file}},
	}

	for _, test := range tests {
		test := test
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			if _, err := parseCLIArgs(test.args); err == nil {
				t.Fatal("parseCLIArgs() error = nil, want error")
			}
		})
	}
}

func TestResolvePort(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name        string
		cliPort     int
		cliPortSet  bool
		defaultPort *int
		want        int
		wantErr     bool
	}{
		{
			name:        "cli wins",
			cliPort:     15556,
			cliPortSet:  true,
			defaultPort: intPtr(15555),
			want:        15556,
		},
		{
			name:        "default port",
			cliPort:     0,
			cliPortSet:  false,
			defaultPort: intPtr(15555),
			want:        15555,
		},
		{
			name:       "missing both",
			cliPort:    0,
			cliPortSet: false,
			wantErr:    true,
		},
		{
			name:        "cli zero",
			cliPort:     0,
			cliPortSet:  true,
			defaultPort: intPtr(15555),
			wantErr:     true,
		},
	}

	for _, test := range tests {
		test := test
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			got, err := resolvePort(test.cliPort, test.cliPortSet, test.defaultPort)
			if test.wantErr {
				if err == nil {
					t.Fatal("resolvePort() error = nil, want error")
				}
				return
			}
			if err != nil {
				t.Fatalf("resolvePort() error = %v", err)
			}
			if got != test.want {
				t.Fatalf("resolvePort() = %d, want %d", got, test.want)
			}
		})
	}
}

func TestPrintServerStartup(t *testing.T) {
	t.Parallel()

	var out bytes.Buffer
	logger := log.New(&out, "", log.LstdFlags)

	printServerStartup(&out, logger, testServerInfo())

	got := out.String()
	for _, want := range []string{
		"Links\n",
		"  http://localhost:15555/go   -> https://go.dev/\n",
		"  http://localhost:15555/pkg  -> https://pkg.go.dev/\n",
		"  http://localhost:15555/repo -> https://github.com/\n",
		"\nShortcuts\n",
		"  c clear   u links   q quit\n",
		"\nEvents\n",
		"serving at http://localhost:15555/\n",
	} {
		if !strings.Contains(got, want) {
			t.Fatalf("printServerStartup() output = %q, want to contain %q", got, want)
		}
	}
	for _, unwanted := range []string{
		"Local:",
		"Links: 3",
		"loaded 3 links",
		"alias-server started",
		"alias-server ready",
		"server listening",
	} {
		if strings.Contains(got, unwanted) {
			t.Fatalf("printServerStartup() output = %q, want no %q", got, unwanted)
		}
	}
	if eventsIndex := strings.Index(got, "Events\n"); eventsIndex == -1 {
		t.Fatalf("printServerStartup() output = %q, want Events header", got)
	} else if servingIndex := strings.Index(got, "serving at http://localhost:15555/\n"); servingIndex == -1 || servingIndex < eventsIndex {
		t.Fatalf("printServerStartup() output = %q, want Events before serving log", got)
	}
	assertInOrder(t, got,
		"http://localhost:15555/go   -> https://go.dev/",
		"http://localhost:15555/pkg  -> https://pkg.go.dev/",
		"http://localhost:15555/repo -> https://github.com/",
	)
	if strings.Contains(got, "\033[") {
		t.Fatalf("printServerStartup() output = %q, want no ANSI codes", got)
	}
}

func TestPrintServerLinks(t *testing.T) {
	t.Parallel()

	var out bytes.Buffer

	printServerLinks(&out, testServerInfo())

	got := out.String()
	for _, want := range []string{
		"Links\n",
		"  http://localhost:15555/go   -> https://go.dev/\n",
		"  http://localhost:15555/pkg  -> https://pkg.go.dev/\n",
		"  http://localhost:15555/repo -> https://github.com/\n",
	} {
		if !strings.Contains(got, want) {
			t.Fatalf("printServerLinks() output = %q, want to contain %q", got, want)
		}
	}
	assertInOrder(t, got,
		"http://localhost:15555/go   -> https://go.dev/",
		"http://localhost:15555/pkg  -> https://pkg.go.dev/",
		"http://localhost:15555/repo -> https://github.com/",
	)
	for _, unwanted := range []string{"Events\n", "Shortcuts\n", "Local:", "Links: 3", "serving at", "loaded"} {
		if strings.Contains(got, unwanted) {
			t.Fatalf("printServerLinks() output = %q, want no %q", got, unwanted)
		}
	}
}

func TestPrintServerShortcuts(t *testing.T) {
	t.Parallel()

	var out bytes.Buffer

	printServerShortcuts(&out, false)

	got := out.String()
	for _, want := range []string{
		"Shortcuts\n",
		"  c clear   u links   q quit\n",
	} {
		if !strings.Contains(got, want) {
			t.Fatalf("printServerShortcuts() output = %q, want to contain %q", got, want)
		}
	}
	for _, unwanted := range []string{"Events\n", "Links\n", "serving at"} {
		if strings.Contains(got, unwanted) {
			t.Fatalf("printServerShortcuts() output = %q, want no %q", got, unwanted)
		}
	}
}

func TestPrintServerStartupWithColor(t *testing.T) {
	t.Parallel()

	var out bytes.Buffer
	logger := log.New(&out, "", log.LstdFlags)
	info := testServerInfo()
	info.color = true

	printServerStartup(&out, logger, info)

	got := out.String()
	for _, want := range []string{
		"\033[1mLinks\033[0m\n",
		"\033[36mhttp://localhost:15555/go\033[0m   \033[2m->\033[0m \033[36mhttps://go.dev/\033[0m\n",
		"\033[1mShortcuts\033[0m\n",
		"  \033[1mc\033[0m clear   \033[1mu\033[0m links   \033[1mq\033[0m quit\n",
		"\033[1mEvents\033[0m\n",
	} {
		if !strings.Contains(got, want) {
			t.Fatalf("printServerStartup() output = %q, want to contain %q", got, want)
		}
	}
}

func TestClearScreenPrintsEscapeAndServerStartup(t *testing.T) {
	t.Parallel()

	var out bytes.Buffer
	logger := log.New(&out, "", log.LstdFlags)

	clearScreen(&out, logger, testServerInfo())

	got := out.String()
	if !strings.HasPrefix(got, "\033[H\033[2J") {
		t.Fatalf("clearScreen() output = %q, want ANSI clear prefix", got)
	}
	if !strings.Contains(got, "serving at http://localhost:15555/\n") {
		t.Fatalf("clearScreen() output = %q, want serving log", got)
	}
	if !strings.Contains(got, "Links\n") {
		t.Fatalf("clearScreen() output = %q, want links header", got)
	}
	if !strings.Contains(got, "Shortcuts\n") {
		t.Fatalf("clearScreen() output = %q, want shortcuts header", got)
	}
	if !strings.Contains(got, "\nEvents\n") {
		t.Fatalf("clearScreen() output = %q, want Events header", got)
	}
}

func TestHandleShortcutsQuits(t *testing.T) {
	t.Parallel()

	var out bytes.Buffer
	logger := log.New(&out, "", log.LstdFlags)
	called := false

	handleShortcuts(strings.NewReader("q\n"), &out, logger, testServerInfo(), func() {
		called = true
	})

	if !called {
		t.Fatal("handleShortcuts() did not call shutdown")
	}
	if out.Len() != 0 {
		t.Fatalf("handleShortcuts() output = %q, want empty", out.String())
	}
}

func TestHandleShortcutsClearsAndPrintsURL(t *testing.T) {
	t.Parallel()

	var out bytes.Buffer
	logger := log.New(&out, "", log.LstdFlags)

	handleShortcuts(strings.NewReader("c\nu\n"), &out, logger, testServerInfo(), func() {
		t.Fatal("shutdown called")
	})

	got := out.String()
	if !strings.Contains(got, "\033[H\033[2J") {
		t.Fatalf("handleShortcuts() output = %q, want clear sequence", got)
	}
	if count := strings.Count(got, "Links\n"); count != 2 {
		t.Fatalf("Links header count = %d, want 2 in output %q", count, got)
	}
	if count := strings.Count(got, "http://localhost:15555/go   -> https://go.dev/\n"); count != 2 {
		t.Fatalf("go link count = %d, want 2 in output %q", count, got)
	}
	if count := strings.Count(got, "Events\n"); count != 1 {
		t.Fatalf("Events header count = %d, want 1 in output %q", count, got)
	}
	if count := strings.Count(got, "serving at http://localhost:15555/\n"); count != 1 {
		t.Fatalf("serving log count = %d, want 1 in output %q", count, got)
	}
	if count := strings.Count(got, "Shortcuts\n"); count != 1 {
		t.Fatalf("Shortcuts count = %d, want 1 in output %q", count, got)
	}
}

func TestHandleShortcutsLogsScannerError(t *testing.T) {
	t.Parallel()

	var out bytes.Buffer
	logger := log.New(&out, "", log.LstdFlags)

	handleShortcuts(errorReader{}, &out, logger, testServerInfo(), func() {
		t.Fatal("shutdown called")
	})

	got := out.String()
	if !strings.Contains(got, "shortcut input error: read failed\n") {
		t.Fatalf("handleShortcuts() log = %q, want scanner error", got)
	}
}

type errorReader struct{}

func (errorReader) Read([]byte) (int, error) {
	return 0, errors.New("read failed")
}

func testServerInfo() serverInfo {
	return serverInfo{
		port: 15555,
		links: map[string]string{
			"repo": "https://github.com/",
			"go":   "https://go.dev/",
			"pkg":  "https://pkg.go.dev/",
		},
	}
}

func assertInOrder(t *testing.T, got string, values ...string) {
	t.Helper()

	previous := -1
	for _, value := range values {
		index := strings.Index(got, value)
		if index == -1 {
			t.Fatalf("output = %q, want to contain %q", got, value)
		}
		if index < previous {
			t.Fatalf("output = %q, want %q after previous value", got, value)
		}
		previous = index
	}
}
