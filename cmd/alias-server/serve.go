package main

import (
	"context"
	"errors"
	"fmt"
	"log"
	"net"
	"net/http"
	"os"
)

func serve(options runOptions) error {
	addr := fmt.Sprintf(":%d", options.port)
	listener, err := net.Listen("tcp", addr)
	if err != nil {
		return err
	}

	logger := log.New(os.Stdout, "", log.LstdFlags)
	server := &http.Server{
		Handler: withRequestLogging(newHandler(options.links, options.indexLink), logger),
	}
	serverInfo := serverInfo{
		port:      options.port,
		indexLink: options.indexLink,
		links:     options.links,
		color:     stdoutIsTerminal(),
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
