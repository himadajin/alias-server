package main

import (
	"encoding/json"
	"fmt"
	"net/url"
	"os"
	"regexp"
)

const (
	minPort = 1
	maxPort = 65535
)

var linkNamePattern = regexp.MustCompile(`^[a-z0-9](?:[a-z0-9-]{0,62}[a-z0-9])?$`)

type Config struct {
	DefaultPort *int              `json:"defaultPort,omitempty"`
	IndexLink   *string           `json:"indexLink,omitempty"`
	Links       map[string]string `json:"links"`
}

func loadConfig(path string) (Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return Config{}, fmt.Errorf("read config: %w", err)
	}

	var config Config
	if err := json.Unmarshal(data, &config); err != nil {
		return Config{}, fmt.Errorf("parse config: %w", err)
	}

	if err := validateConfig(config); err != nil {
		return Config{}, err
	}

	return config, nil
}

func validateConfig(config Config) error {
	if config.DefaultPort != nil && !isValidPort(*config.DefaultPort) {
		return fmt.Errorf("defaultPort must be between %d and %d", minPort, maxPort)
	}
	if len(config.Links) == 0 {
		return fmt.Errorf("links must contain at least one entry")
	}
	if config.IndexLink != nil {
		if !isValidLinkName(*config.IndexLink) {
			return fmt.Errorf("invalid indexLink %q", *config.IndexLink)
		}
		if _, ok := config.Links[*config.IndexLink]; ok {
			return fmt.Errorf("indexLink %q conflicts with link name", *config.IndexLink)
		}
	}

	for name, target := range config.Links {
		if !isValidLinkName(name) {
			return fmt.Errorf("invalid link name %q", name)
		}
		if !isValidTargetURL(target) {
			return fmt.Errorf("invalid target URL for link %q", name)
		}
	}

	return nil
}

func isValidLinkName(name string) bool {
	return linkNamePattern.MatchString(name)
}

func isValidPort(port int) bool {
	return port >= minPort && port <= maxPort
}

func isValidTargetURL(raw string) bool {
	parsed, err := url.Parse(raw)
	if err != nil {
		return false
	}
	return parsed.Scheme != "" && parsed.Host != ""
}
