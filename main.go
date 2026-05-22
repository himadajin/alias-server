package main

import (
	"fmt"
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
	if len(args) != 1 {
		return fmt.Errorf("usage: alias-server config.json")
	}

	config, err := loadConfig(args[0])
	if err != nil {
		return err
	}

	addr := fmt.Sprintf(":%d", config.Port)
	return http.ListenAndServe(addr, newHandler(config.Links))
}
