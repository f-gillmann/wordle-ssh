package main

import (
	"os"

	"github.com/charmbracelet/log"
	"github.com/f-gillmann/wordle-ssh/internal/server"
)

func main() {
	// Load configuration
	config := server.LoadConfigFromEnv()

	// Create and start the server
	srv, err := server.New(config)
	if err != nil {
		log.Fatal(err)
		os.Exit(1)
	}

	if err := srv.Start(); err != nil {
		log.Fatal(err)
		os.Exit(1)
	}
}
