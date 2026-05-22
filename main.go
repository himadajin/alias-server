package main

import (
	"bufio"
	"context"
	"errors"
	"flag"
	"fmt"
	"io"
	"log"
	"net"
	"net/http"
	"os"
	"sort"
	"strings"
)

func main() {
	if err := run(os.Args[1:]); err != nil {
		fmt.Fprintf(os.Stderr, "alias-server: %v\n", err)
		os.Exit(1)
	}
}

func run(args []string) error {
	options, err := parseCLIArgs(args)
	if err != nil {
		return err
	}

	return serve(options)
}

func serve(options runOptions) error {
	addr := fmt.Sprintf(":%d", options.port)
	listener, err := net.Listen("tcp", addr)
	if err != nil {
		return err
	}

	logger := log.New(os.Stdout, "", log.LstdFlags)
	server := &http.Server{
		Handler: withRequestLogging(newHandler(options.links), logger),
	}
	serverInfo := serverInfo{
		port:  options.port,
		links: options.links,
		color: stdoutIsTerminal(),
	}

	printServerStartup(os.Stdout, logger, serverInfo)
	if stdinIsTerminal() {
		go handleShortcuts(os.Stdin, os.Stdout, logger, serverInfo, func() {
			if err := server.Shutdown(context.Background()); err != nil {
				logger.Printf("shutdown error: %v", err)
			}
		})
	}

	if err := server.Serve(listener); err != nil && !errors.Is(err, http.ErrServerClosed) {
		return err
	}
	return nil
}

type runOptions struct {
	port  int
	links map[string]string
}

func parseCLIArgs(args []string) (runOptions, error) {
	flags := flag.NewFlagSet("alias-server", flag.ContinueOnError)
	flags.SetOutput(io.Discard)

	port := flags.Int("port", 0, "port to listen on")
	if err := flags.Parse(args); err != nil {
		return runOptions{}, err
	}
	if flags.NArg() != 1 {
		return runOptions{}, fmt.Errorf("usage: alias-server [-port PORT] config.json")
	}

	config, err := loadConfig(flags.Arg(0))
	if err != nil {
		return runOptions{}, err
	}

	resolvedPort, err := resolvePort(*port, flagWasSet(flags, "port"), config.DefaultPort)
	if err != nil {
		return runOptions{}, err
	}

	return runOptions{
		port:  resolvedPort,
		links: config.Links,
	}, nil
}

func resolvePort(cliPort int, cliPortSet bool, defaultPort *int) (int, error) {
	if cliPortSet {
		if !isValidPort(cliPort) {
			return 0, fmt.Errorf("port must be between %d and %d", minPort, maxPort)
		}
		return cliPort, nil
	}

	if defaultPort == nil {
		return 0, fmt.Errorf("port is required: specify -port or defaultPort")
	}
	if !isValidPort(*defaultPort) {
		return 0, fmt.Errorf("defaultPort must be between %d and %d", minPort, maxPort)
	}

	return *defaultPort, nil
}

func flagWasSet(flags *flag.FlagSet, name string) bool {
	wasSet := false
	flags.Visit(func(flag *flag.Flag) {
		if flag.Name == name {
			wasSet = true
		}
	})
	return wasSet
}

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
