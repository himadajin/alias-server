package main

import (
	"bytes"
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

	printServerStartup(&out, logger, serverInfo{port: 15555, linkCount: 3})

	got := out.String()
	for _, want := range []string{
		"  Shortcuts: c clear, u url, q quit\n",
		"\nEvents:\n",
		"serving at http://localhost:15555/\n",
		"loaded 3 links\n",
	} {
		if !strings.Contains(got, want) {
			t.Fatalf("printServerStartup() output = %q, want to contain %q", got, want)
		}
	}
	for _, unwanted := range []string{
		"Local:",
		"Links:",
		"alias-server started",
		"alias-server ready",
		"server listening",
	} {
		if strings.Contains(got, unwanted) {
			t.Fatalf("printServerStartup() output = %q, want no %q", got, unwanted)
		}
	}
	if eventsIndex := strings.Index(got, "Events:\n"); eventsIndex == -1 {
		t.Fatalf("printServerStartup() output = %q, want Events header", got)
	} else if servingIndex := strings.Index(got, "serving at http://localhost:15555/\n"); servingIndex == -1 || servingIndex < eventsIndex {
		t.Fatalf("printServerStartup() output = %q, want Events before serving log", got)
	}
}

func TestPrintServerInfo(t *testing.T) {
	t.Parallel()

	var out bytes.Buffer

	printServerInfo(&out, serverInfo{port: 15555, linkCount: 3})

	got := out.String()
	for _, want := range []string{
		"  Shortcuts: c clear, u url, q quit\n",
	} {
		if !strings.Contains(got, want) {
			t.Fatalf("printServerInfo() output = %q, want to contain %q", got, want)
		}
	}
	for _, unwanted := range []string{"Events:", "Local:", "Links:", "serving at"} {
		if strings.Contains(got, unwanted) {
			t.Fatalf("printServerInfo() output = %q, want no %q", got, unwanted)
		}
	}
}

func TestClearScreenPrintsEscapeAndServerStartup(t *testing.T) {
	t.Parallel()

	var out bytes.Buffer
	logger := log.New(&out, "", log.LstdFlags)

	clearScreen(&out, logger, serverInfo{port: 15555, linkCount: 1})

	got := out.String()
	if !strings.HasPrefix(got, "\033[H\033[2J") {
		t.Fatalf("clearScreen() output = %q, want ANSI clear prefix", got)
	}
	if !strings.Contains(got, "serving at http://localhost:15555/\n") {
		t.Fatalf("clearScreen() output = %q, want serving log", got)
	}
	if !strings.Contains(got, "loaded 1 links\n") {
		t.Fatalf("clearScreen() output = %q, want link count log", got)
	}
	if !strings.Contains(got, "\nEvents:\n") {
		t.Fatalf("clearScreen() output = %q, want Events header", got)
	}
}

func TestHandleShortcutsQuits(t *testing.T) {
	t.Parallel()

	var out bytes.Buffer
	logger := log.New(&out, "", log.LstdFlags)
	called := false

	handleShortcuts(strings.NewReader("q\n"), &out, logger, serverInfo{port: 15555, linkCount: 1}, func() {
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

	handleShortcuts(strings.NewReader("c\nu\n"), &out, logger, serverInfo{port: 15555, linkCount: 1}, func() {
		t.Fatal("shutdown called")
	})

	got := out.String()
	if !strings.Contains(got, "\033[H\033[2J") {
		t.Fatalf("handleShortcuts() output = %q, want clear sequence", got)
	}
	if count := strings.Count(got, "serving at http://localhost:15555/\n"); count != 2 {
		t.Fatalf("serving log count = %d, want 2 in output %q", count, got)
	}
	if count := strings.Count(got, "Events:\n"); count != 1 {
		t.Fatalf("Events header count = %d, want 1 in output %q", count, got)
	}
	if count := strings.Count(got, "  Shortcuts: c clear, u url, q quit\n"); count != 1 {
		t.Fatalf("Shortcuts count = %d, want 1 in output %q", count, got)
	}
}
