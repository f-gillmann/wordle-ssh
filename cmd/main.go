package main

import (
	"os"

	"github.com/charmbracelet/log"
	"github.com/f-gillmann/wordle-ssh/internal/server"
)

func main() {
	// Load configuration
	config := server.LoadConfigFromEnv()

	// Initialize logger
	logger := log.NewWithOptions(os.Stderr, log.Options{
		ReportTimestamp: true,
		ReportCaller:    config.LogLevel == log.DebugLevel,
		TimeFormat:      "2006/01/02 15:04:05",
		Prefix:          "[wordle-ssh]",
		Level:           config.LogLevel,
	})

	// Set the logger in config
	config.Logger = logger

	// Create and start the server
	srv, err := server.New(config)
	if err != nil {
		logger.Fatal("Failed to create server", "error", err)
		os.Exit(1)
	}

	if err := srv.Start(); err != nil {
		logger.Fatal("Failed to start server", "error", err)
		os.Exit(1)
	}
}
