package main

import "testing"

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
