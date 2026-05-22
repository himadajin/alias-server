package main

import (
	"flag"
	"fmt"
	"io"
	"net/http"
	"os"
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
	return http.ListenAndServe(addr, newHandler(options.links))
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
