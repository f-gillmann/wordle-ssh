package server

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/log"
	"github.com/charmbracelet/ssh"
	"github.com/charmbracelet/wish"
	"github.com/charmbracelet/wish/bubbletea"
	"github.com/charmbracelet/wish/logging"
	"github.com/f-gillmann/wordle-ssh/internal/ui"
	"github.com/f-gillmann/wordle-ssh/internal/wordle"
)

const (
	defaultHost        = "0.0.0.0"
	defaultPort        = "23234"
	defaultHostKeyPath = ".ssh/id_ed25519"
)

// Config holds the server configuration
type Config struct {
	Host        string
	Port        string
	HostKeyPath string
	Logger      *log.Logger
}

// LoadConfigFromEnv loads configuration from environment variables
func LoadConfigFromEnv() Config {
	host := os.Getenv("WORDLE_SSH_HOST")
	if host == "" {
		host = defaultHost
	}

	port := os.Getenv("WORDLE_SSH_PORT")
	if port == "" {
		port = defaultPort
	}

	hostKeyPath := os.Getenv("WORDLE_SSH_HOST_KEY_PATH")
	if hostKeyPath == "" {
		hostKeyPath = defaultHostKeyPath
	}

	return Config{
		Host:        host,
		Port:        port,
		HostKeyPath: hostKeyPath,
	}
}

// Server represents the SSH server
type Server struct {
	config     Config
	wordleWord string
	wordleDate string
	wishServer *ssh.Server
}

// New creates a new SSH server
func New(config Config) (*Server, error) {
	if config.Host == "" {
		config.Host = defaultHost
	}

	if config.Port == "" {
		config.Port = defaultPort
	}

	if config.HostKeyPath == "" {
		config.HostKeyPath = defaultHostKeyPath
	}

	if config.Logger == nil {
		config.Logger = log.NewWithOptions(os.Stderr, log.Options{
			ReportTimestamp: true,
			TimeFormat:      time.Kitchen,
			Prefix:          "Wordle SSH",
		})
	}

	s := &Server{
		config: config,
	}

	// Fetch today's Wordle word
	if err := s.refreshWordleWord(); err != nil {
		return nil, fmt.Errorf("failed to fetch wordle word: %w", err)
	}

	// Create wish server with bubbletea middleware
	wishServer, err := wish.NewServer(
		wish.WithAddress(fmt.Sprintf("%s:%s", config.Host, config.Port)),
		wish.WithHostKeyPath(config.HostKeyPath),
		wish.WithMiddleware(
			bubbletea.Middleware(s.teaHandler),
			logging.Middleware(),
		),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create server: %w", err)
	}
	s.wishServer = wishServer

	return s, nil
}

// refreshWordleWord fetches the Wordle word only if it's a new day
func (s *Server) refreshWordleWord() error {
	today := time.Now().Format("2006-01-02")

	// Only fetch if we don't have a word yet or if the date has changed
	if s.wordleWord == "" || s.wordleDate != today {
		word, err := wordle.FetchTodayWord()
		if err != nil {
			return fmt.Errorf("failed to fetch wordle word: %w", err)
		}

		s.wordleWord = word
		s.wordleDate = today

		s.config.Logger.Info("Fetched Wordle word", "date", s.wordleDate, "word", s.wordleWord)
	}

	return nil
}

// teaHandler creates a bubbletea program for each SSH session
func (s *Server) teaHandler(ssh.Session) (tea.Model, []tea.ProgramOption) {
	// Refresh Wordle word if it's a new day
	if err := s.refreshWordleWord(); err != nil {
		s.config.Logger.Error("Failed to refresh Wordle word", "error", err)
		return nil, nil
	}

	// Create the app model with the current word
	m := ui.NewAppModel(s.wordleWord)

	return m, []tea.ProgramOption{tea.WithAltScreen()}
}

// Start starts the SSH server
func (s *Server) Start() error {
	done := make(chan os.Signal, 1)
	signal.Notify(done, os.Interrupt, syscall.SIGINT, syscall.SIGTERM)

	s.config.Logger.Info("Starting SSH server", "host", s.config.Host, "port", s.config.Port)

	go func() {
		if err := s.wishServer.ListenAndServe(); err != nil {
			s.config.Logger.Fatal("Server error", "error", err)
		}
	}()

	<-done
	s.config.Logger.Info("Stopping SSH server")

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := s.wishServer.Shutdown(ctx); err != nil {
		return fmt.Errorf("failed to shutdown server: %w", err)
	}

	return nil
}
