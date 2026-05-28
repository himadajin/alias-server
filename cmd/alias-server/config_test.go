package main

import (
	"os"
	"strings"
	"testing"
)

func TestLoadConfig(t *testing.T) {
	t.Parallel()

	file := writeTempConfig(t, `{
		"defaultPort": 55555,
		"indexLink": "index",
		"links": {
			"go": "https://go.dev/",
			"pkg": "https://pkg.go.dev/"
		}
	}`)

	config, err := loadConfig(file)
	if err != nil {
		t.Fatalf("loadConfig() error = %v", err)
	}
	if config.DefaultPort == nil {
		t.Fatal("config.DefaultPort = nil, want 55555")
	}
	if *config.DefaultPort != 55555 {
		t.Fatalf("*config.DefaultPort = %d, want 55555", *config.DefaultPort)
	}
	if config.Links["go"] != "https://go.dev/" {
		t.Fatalf("config.Links[go] = %q, want https://go.dev/", config.Links["go"])
	}
	if config.IndexLink == nil {
		t.Fatal("config.IndexLink = nil, want index")
	}
	if *config.IndexLink != "index" {
		t.Fatalf("*config.IndexLink = %q, want index", *config.IndexLink)
	}
}

func TestLoadConfigAllowsOmittedDefaultPortAndIndexLink(t *testing.T) {
	t.Parallel()

	file := writeTempConfig(t, `{
		"links": {
			"go": "https://go.dev/"
		}
	}`)

	config, err := loadConfig(file)
	if err != nil {
		t.Fatalf("loadConfig() error = %v", err)
	}
	if config.DefaultPort != nil {
		t.Fatalf("config.DefaultPort = %d, want nil", *config.DefaultPort)
	}
	if config.IndexLink != nil {
		t.Fatalf("config.IndexLink = %q, want nil", *config.IndexLink)
	}
}

func TestLoadConfigAllowsNullIndexLink(t *testing.T) {
	t.Parallel()

	file := writeTempConfig(t, `{
		"defaultPort": 55555,
		"indexLink": null,
		"links": {
			"go": "https://go.dev/"
		}
	}`)

	config, err := loadConfig(file)
	if err != nil {
		t.Fatalf("loadConfig() error = %v", err)
	}
	if config.IndexLink != nil {
		t.Fatalf("config.IndexLink = %q, want nil", *config.IndexLink)
	}
}

func TestLoadConfigRejectsInvalidJSON(t *testing.T) {
	t.Parallel()

	file := writeTempConfig(t, `{`)

	if _, err := loadConfig(file); err == nil {
		t.Fatal("loadConfig() error = nil, want error")
	}
}

func TestValidateConfigRejectsInvalidConfig(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name   string
		config Config
	}{
		{
			name: "default port too low",
			config: Config{
				DefaultPort: intPtr(0),
				Links:       map[string]string{"go": "https://go.dev/"},
			},
		},
		{
			name: "default port negative",
			config: Config{
				DefaultPort: intPtr(-1),
				Links:       map[string]string{"go": "https://go.dev/"},
			},
		},
		{
			name: "default port too high",
			config: Config{
				DefaultPort: intPtr(65536),
				Links:       map[string]string{"go": "https://go.dev/"},
			},
		},
		{
			name: "empty links",
			config: Config{
				DefaultPort: intPtr(55555),
				Links:       map[string]string{},
			},
		},
		{
			name: "invalid link name",
			config: Config{
				DefaultPort: intPtr(55555),
				Links:       map[string]string{"Go": "https://go.dev/"},
			},
		},
		{
			name: "invalid target URL",
			config: Config{
				DefaultPort: intPtr(55555),
				Links:       map[string]string{"go": "go.dev"},
			},
		},
		{
			name: "invalid index link",
			config: Config{
				DefaultPort: intPtr(55555),
				IndexLink:   stringPtr("Index"),
				Links:       map[string]string{"go": "https://go.dev/"},
			},
		},
		{
			name: "index link conflicts with link name",
			config: Config{
				DefaultPort: intPtr(55555),
				IndexLink:   stringPtr("index"),
				Links: map[string]string{
					"index": "https://example.com/",
					"go":    "https://go.dev/",
				},
			},
		},
	}

	for _, test := range tests {
		test := test
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			if err := validateConfig(test.config); err == nil {
				t.Fatal("validateConfig() error = nil, want error")
			}
		})
	}
}

func TestIsValidLinkName(t *testing.T) {
	t.Parallel()

	valid64 := strings.Repeat("a", 64)
	tests := []struct {
		name string
		want bool
	}{
		{name: "go", want: true},
		{name: "go-dev", want: true},
		{name: "a1", want: true},
		{name: valid64, want: true},
		{name: "Go", want: false},
		{name: "go_dev", want: false},
		{name: "go.dev", want: false},
		{name: "go~dev", want: false},
		{name: "-go", want: false},
		{name: "go-", want: false},
		{name: "", want: false},
		{name: strings.Repeat("a", 65), want: false},
		{name: "メモ", want: false},
		{name: "go%dev", want: false},
		{name: "go/dev", want: false},
		{name: "go?dev", want: false},
		{name: "go#dev", want: false},
	}

	for _, test := range tests {
		test := test
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			if got := isValidLinkName(test.name); got != test.want {
				t.Fatalf("isValidLinkName(%q) = %v, want %v", test.name, got, test.want)
			}
		})
	}
}

func writeTempConfig(t *testing.T, content string) string {
	t.Helper()

	file, err := os.CreateTemp(t.TempDir(), "config-*.json")
	if err != nil {
		t.Fatalf("CreateTemp() error = %v", err)
	}
	defer file.Close()

	if _, err := file.WriteString(content); err != nil {
		t.Fatalf("WriteString() error = %v", err)
	}

	return file.Name()
}

func intPtr(value int) *int {
	return &value
}

func stringPtr(value string) *string {
	return &value
}
