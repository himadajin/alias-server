package main

import (
	"fmt"
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
