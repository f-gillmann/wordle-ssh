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
	"github.com/charmbracelet/wish/activeterm"
	"github.com/charmbracelet/wish/bubbletea"
	"github.com/charmbracelet/wish/logging"
	"github.com/f-gillmann/wordle-ssh/internal/stats"
	"github.com/f-gillmann/wordle-ssh/internal/ui"
	"github.com/f-gillmann/wordle-ssh/internal/wordle"
	"github.com/muesli/termenv"
)

const (
	defaultHost        = "0.0.0.0"
	defaultPort        = "23234"
	defaultHostKeyPath = ".ssh/id_ed25519"
	defaultDBPath      = "./wordle-stats.db"
)

// Config holds the server configuration
type Config struct {
	Host        string
	Port        string
	HostKeyPath string
	DBPath      string
	Logger      *log.Logger
	LogLevel    log.Level
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

	dbPath := os.Getenv("WORDLE_SSH_DB_PATH")
	if dbPath == "" {
		dbPath = defaultDBPath
	}

	logLevel := os.Getenv("WORDLE_SSH_LOG_LEVEL")
	var level log.Level

	switch logLevel {
	case "debug":
		level = log.DebugLevel
	case "info":
		level = log.InfoLevel
	case "warn":
		level = log.WarnLevel
	case "error":
		level = log.ErrorLevel
	default:
		level = log.InfoLevel
	}

	return Config{
		Host:        host,
		Port:        port,
		HostKeyPath: hostKeyPath,
		DBPath:      dbPath,
		LogLevel:    level,
	}
}

// Server represents the SSH server
type Server struct {
	config     Config
	wordleWord string
	wordleDate string
	wishServer *ssh.Server
	statsStore *stats.Store
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

	if config.DBPath == "" {
		config.DBPath = defaultDBPath
	}

	if config.Logger == nil {
		return nil, fmt.Errorf("logger must be provided in config")
	}

	s := &Server{
		config: config,
	}

	// Initialize stats store
	statsStore, err := stats.NewStore(config.DBPath, config.Logger)
	if err != nil {
		return nil, fmt.Errorf("failed to initialize stats store: %w", err)
	}
	s.statsStore = statsStore

	// Fetch today's Wordle word
	if err := s.refreshWordleWord(); err != nil {
		return nil, fmt.Errorf("failed to fetch wordle word: %w", err)
	}

	// Create wish server with bubbletea middleware
	wishServer, err := wish.NewServer(
		wish.WithAddress(fmt.Sprintf("%s:%s", config.Host, config.Port)),
		wish.WithHostKeyPath(config.HostKeyPath),
		wish.WithMiddleware(
			bubbletea.MiddlewareWithColorProfile(s.teaHandler, termenv.ANSI256),
			activeterm.Middleware(),
			logging.StructuredMiddlewareWithLogger(config.Logger, config.LogLevel),
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
func (s *Server) teaHandler(sshSession ssh.Session) (tea.Model, []tea.ProgramOption) {
	// Refresh Wordle word if it's a new day
	if err := s.refreshWordleWord(); err != nil {
		s.config.Logger.Error("Failed to refresh Wordle word", "error", err)
		return nil, nil
	}

	// Get username from SSH session
	username := sshSession.User()
	if username == "" {
		username = "anonymous"
	}

	// Check if username is blacklisted
	isBlacklisted := stats.IsBlacklisted(username)

	// Check if user has already played today (but not for blacklisted users)
	hasPlayed := false
	if !isBlacklisted {
		var err error
		hasPlayed, err = s.statsStore.HasPlayedToday(username, s.wordleDate)

		if err != nil {
			s.config.Logger.Error("Failed to check if user played today", "error", err, "username", username)
		}
	}

	// Create the app model with the current word, stats store, and logger
	m := ui.NewAppModel(s.wordleWord, s.wordleDate, username, s.statsStore, hasPlayed, isBlacklisted, s.config.Logger)

	opts := []tea.ProgramOption{tea.WithAltScreen()}
	opts = append(opts, bubbletea.MakeOptions(sshSession)...)

	return m, opts
}

// Start starts the SSH server
func (s *Server) Start() error {
	done := make(chan os.Signal, 1)
	signal.Notify(done, os.Interrupt, syscall.SIGINT, syscall.SIGTERM)

	s.config.Logger.Info("Starting SSH server", "host", s.config.Host, "port", s.config.Port, "db", s.config.DBPath)

	go func() {
		if err := s.wishServer.ListenAndServe(); err != nil {
			s.config.Logger.Fatal("Server error", "error", err)
		}
	}()

	<-done
	s.config.Logger.Info("Stopping SSH server")

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// Close stats store
	if err := s.statsStore.Close(); err != nil {
		s.config.Logger.Error("Failed to close stats store", "error", err)
	}

	if err := s.wishServer.Shutdown(ctx); err != nil {
		return fmt.Errorf("failed to shutdown server: %w", err)
	}

	return nil
}
