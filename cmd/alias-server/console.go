package main

import (
	"bufio"
	"fmt"
	"io"
	"log"
	"os"
	"sort"
	"strings"
)

type serverInfo struct {
	port  int
	links map[string]string
	color bool
}

func printServerStartup(w io.Writer, logger *log.Logger, info serverInfo) {
	printServerLinks(w, info)
	fmt.Fprintln(w)
	printServerShortcuts(w, info.color)
	fmt.Fprintln(w)
	fmt.Fprintln(w, styleBold("Events", info.color))
	logServerURL(logger, info)
}

func printServerLinks(w io.Writer, info serverInfo) {
	fmt.Fprintln(w, styleBold("Links", info.color))
	names := sortedLinkNames(info.links)
	width := maxLocalURLWidth(info, names)
	for _, name := range names {
		localURL := fmt.Sprintf("http://localhost:%d/%s", info.port, name)
		targetURL := info.links[name]
		fmt.Fprintf(
			w,
			"  %s%s %s %s\n",
			styleCyan(localURL, info.color),
			strings.Repeat(" ", width-len(localURL)),
			styleDim("->", info.color),
			styleCyan(targetURL, info.color),
		)
	}
}

func printServerShortcuts(w io.Writer, color bool) {
	fmt.Fprintln(w, styleBold("Shortcuts", color))
	fmt.Fprintf(
		w,
		"  %s clear   %s links   %s quit\n",
		styleBold("c", color),
		styleBold("u", color),
		styleBold("q", color),
	)
}

func logServerURL(logger *log.Logger, info serverInfo) {
	logger.Printf("serving at http://localhost:%d/", info.port)
}

func clearScreen(w io.Writer, logger *log.Logger, info serverInfo) {
	fmt.Fprint(w, "\033[H\033[2J")
	printServerStartup(w, logger, info)
}

func handleShortcuts(r io.Reader, w io.Writer, logger *log.Logger, info serverInfo, shutdown func()) {
	scanner := bufio.NewScanner(r)
	for scanner.Scan() {
		switch scanner.Text() {
		case "c":
			clearScreen(w, logger, info)
		case "u":
			printServerLinks(w, info)
		case "q":
			shutdown()
			return
		}
	}
	if err := scanner.Err(); err != nil {
		logger.Printf("shortcut input error: %v", err)
	}
}

func sortedLinkNames(links map[string]string) []string {
	names := make([]string, 0, len(links))
	for name := range links {
		names = append(names, name)
	}
	sort.Strings(names)
	return names
}

func maxLocalURLWidth(info serverInfo, names []string) int {
	width := 0
	for _, name := range names {
		localURL := fmt.Sprintf("http://localhost:%d/%s", info.port, name)
		if len(localURL) > width {
			width = len(localURL)
		}
	}
	return width
}

func stdinIsTerminal() bool {
	info, err := os.Stdin.Stat()
	if err != nil {
		return false
	}
	return info.Mode()&os.ModeCharDevice != 0
}

func stdoutIsTerminal() bool {
	info, err := os.Stdout.Stat()
	if err != nil {
		return false
	}
	return info.Mode()&os.ModeCharDevice != 0
}

func styleBold(text string, color bool) string {
	return style(text, color, "1")
}

func styleCyan(text string, color bool) string {
	return style(text, color, "36")
}

func styleDim(text string, color bool) string {
	return style(text, color, "2")
}

func style(text string, color bool, code string) string {
	if !color {
		return text
	}
	return "\033[" + code + "m" + text + "\033[0m"
}
