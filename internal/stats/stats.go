package stats

import (
	"database/sql"
	"errors"
	"fmt"
	"time"

	"github.com/charmbracelet/log"
	_ "github.com/mattn/go-sqlite3"
)

// UserStats represents a user's game statistics
type UserStats struct {
	Username          string
	SSHKeyFingerprint string
	GamesPlayed       int
	GamesWon          int
	GamesLost         int
	CurrentStreak     int
	MaxStreak         int
	GuessDistribution [6]int // Index 0 = 1 guess, Index 5 = 6 guesses
	TotalGuesses      int
	LastPlayed        time.Time
	LastWordDate      string // To prevent playing same word twice
	LastGameResult    string // JSON-encoded game result for display
}

// Store handles database operations for user statistics
type Store struct {
	db     *sql.DB
	logger *log.Logger
}

// NewStore creates a new statistics store
func NewStore(dbPath string, logger *log.Logger) (*Store, error) {
	db, err := sql.Open("sqlite3", dbPath)
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}

	store := &Store{
		db:     db,
		logger: logger,
	}

	if err := store.initSchema(); err != nil {
		return nil, fmt.Errorf("failed to initialize schema: %w", err)
	}

	return store, nil
}

// initSchema creates the necessary database tables
func (s *Store) initSchema() error {
	schema := `
	CREATE TABLE IF NOT EXISTS user_stats (
		username TEXT NOT NULL,
		ssh_key_fingerprint TEXT NOT NULL,
		games_played INTEGER DEFAULT 0,
		games_won INTEGER DEFAULT 0,
		games_lost INTEGER DEFAULT 0,
		current_streak INTEGER DEFAULT 0,
		max_streak INTEGER DEFAULT 0,
		guess_dist_1 INTEGER DEFAULT 0,
		guess_dist_2 INTEGER DEFAULT 0,
		guess_dist_3 INTEGER DEFAULT 0,
		guess_dist_4 INTEGER DEFAULT 0,
		guess_dist_5 INTEGER DEFAULT 0,
		guess_dist_6 INTEGER DEFAULT 0,
		total_guesses INTEGER DEFAULT 0,
		last_played DATETIME,
		last_word_date TEXT,
		last_game_result TEXT,
		created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
		updated_at DATETIME DEFAULT CURRENT_TIMESTAMP,
		PRIMARY KEY (username, ssh_key_fingerprint)
	);

	CREATE INDEX IF NOT EXISTS idx_last_played ON user_stats(last_played);
	CREATE INDEX IF NOT EXISTS idx_games_won ON user_stats(games_won DESC);
	CREATE INDEX IF NOT EXISTS idx_ssh_key ON user_stats(ssh_key_fingerprint);
	`

	if _, err := s.db.Exec(schema); err != nil {
		return fmt.Errorf("failed to create schema: %w", err)
	}

	s.logger.Debug("Database schema initialized")
	return nil
}

// scanUserStats is a helper method to scan user stats from a row scanner
func (s *Store) scanUserStats(scanner interface {
	Scan(dest ...interface{}) error
}, stats *UserStats) error {
	var lastPlayed sql.NullTime

	err := scanner.Scan(
		&stats.Username,
		&stats.SSHKeyFingerprint,
		&stats.GamesPlayed,
		&stats.GamesWon,
		&stats.GamesLost,
		&stats.CurrentStreak,
		&stats.MaxStreak,
		&stats.GuessDistribution[0],
		&stats.GuessDistribution[1],
		&stats.GuessDistribution[2],
		&stats.GuessDistribution[3],
		&stats.GuessDistribution[4],
		&stats.GuessDistribution[5],
		&stats.TotalGuesses,
		&lastPlayed,
		&stats.LastWordDate,
		&stats.LastGameResult,
	)

	if err != nil {
		return err
	}

	if lastPlayed.Valid {
		stats.LastPlayed = lastPlayed.Time
	}

	return nil
}

// GetUserStats retrieves statistics for a user by username AND ssh_key_fingerprint pair
func (s *Store) GetUserStats(username string, sshKeyFingerprint string) (*UserStats, error) {
	s.logger.Debug("Reading user stats", "username", username, "ssh_key_fingerprint", sshKeyFingerprint)

	query := `
		SELECT username, ssh_key_fingerprint, games_played, games_won, games_lost, current_streak, max_streak,
		       guess_dist_1, guess_dist_2, guess_dist_3, guess_dist_4, guess_dist_5, guess_dist_6,
		       total_guesses, last_played, COALESCE(last_word_date, ''), COALESCE(last_game_result, '')
		FROM user_stats
		WHERE username = ? AND ssh_key_fingerprint = ?
	`

	var stats UserStats

	err := s.scanUserStats(s.db.QueryRow(query, username, sshKeyFingerprint), &stats)
	if errors.Is(err, sql.ErrNoRows) {
		// Return empty stats for new user
		s.logger.Debug("No existing stats found, returning empty stats for new user", "username", username)
		return &UserStats{
			Username:          username,
			SSHKeyFingerprint: sshKeyFingerprint,
		}, nil
	}

	if err != nil {
		return nil, fmt.Errorf("failed to get user stats: %w", err)
	}

	s.logger.Debug("Successfully retrieved user stats",
		"username", username,
		"games_played", stats.GamesPlayed,
		"games_won", stats.GamesWon,
		"current_streak", stats.CurrentStreak,
		"last_word_date", stats.LastWordDate,
	)

	return &stats, nil
}

// HasPlayedToday checks if the user has already played today's word
func (s *Store) HasPlayedToday(username string, sshKeyFingerprint string, wordDate string) (bool, error) {
	s.logger.Debug("Checking if user has played today", "username", username, "word_date", wordDate)

	stats, err := s.GetUserStats(username, sshKeyFingerprint)
	if err != nil {
		return false, err
	}

	hasPlayed := stats.LastWordDate == wordDate
	s.logger.Debug("Played today check result",
		"username", username,
		"has_played", hasPlayed,
		"last_word_date", stats.LastWordDate,
		"current_word_date", wordDate,
	)

	return hasPlayed, nil
}

// RecordWin records a winning game for a user
func (s *Store) RecordWin(username string, sshKeyFingerprint string, guesses int, wordDate string, gameResult string) error {
	if guesses < 1 || guesses > 6 {
		return fmt.Errorf("invalid number of guesses: %d", guesses)
	}

	stats, err := s.GetUserStats(username, sshKeyFingerprint)
	if err != nil {
		return err
	}

	stats.GamesPlayed++
	stats.GamesWon++
	stats.CurrentStreak++
	stats.TotalGuesses += guesses
	stats.GuessDistribution[guesses-1]++
	stats.LastPlayed = time.Now()
	stats.LastWordDate = wordDate
	stats.LastGameResult = gameResult

	if stats.CurrentStreak > stats.MaxStreak {
		stats.MaxStreak = stats.CurrentStreak
	}

	if err := s.saveUserStats(stats); err != nil {
		return err
	}

	s.logger.Info("Recorded win", "username", username, "guesses", guesses, "streak", stats.CurrentStreak)
	return nil
}

// RecordLoss records a losing game for a user
func (s *Store) RecordLoss(username string, sshKeyFingerprint string, wordDate string, gameResult string) error {
	stats, err := s.GetUserStats(username, sshKeyFingerprint)
	if err != nil {
		return err
	}

	stats.GamesPlayed++
	stats.GamesLost++
	stats.CurrentStreak = 0 // Reset streak on loss
	stats.LastPlayed = time.Now()
	stats.LastWordDate = wordDate
	stats.LastGameResult = gameResult

	if err := s.saveUserStats(stats); err != nil {
		return err
	}

	s.logger.Info("Recorded loss", "username", username)
	return nil
}

// saveUserStats saves or updates user statistics
func (s *Store) saveUserStats(stats *UserStats) error {
	// Create a hash of the SSH public key for logging

	s.logger.Debug("Saving user stats",
		"username", stats.Username,
		"ssh_key_fingerprint", stats.SSHKeyFingerprint,
		"games_played", stats.GamesPlayed,
		"games_won", stats.GamesWon,
		"games_lost", stats.GamesLost,
		"current_streak", stats.CurrentStreak,
		"max_streak", stats.MaxStreak,
		"total_guesses", stats.TotalGuesses,
		"last_word_date", stats.LastWordDate,
	)

	query := `
		INSERT INTO user_stats (
			username, ssh_key_fingerprint, games_played, games_won, games_lost, current_streak, max_streak,
			guess_dist_1, guess_dist_2, guess_dist_3, guess_dist_4, guess_dist_5, guess_dist_6,
			total_guesses, last_played, last_word_date, last_game_result, updated_at
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, CURRENT_TIMESTAMP)
		ON CONFLICT(username, ssh_key_fingerprint) DO UPDATE SET
			games_played = excluded.games_played,
			games_won = excluded.games_won,
			games_lost = excluded.games_lost,
			current_streak = excluded.current_streak,
			max_streak = excluded.max_streak,
			guess_dist_1 = excluded.guess_dist_1,
			guess_dist_2 = excluded.guess_dist_2,
			guess_dist_3 = excluded.guess_dist_3,
			guess_dist_4 = excluded.guess_dist_4,
			guess_dist_5 = excluded.guess_dist_5,
			guess_dist_6 = excluded.guess_dist_6,
			total_guesses = excluded.total_guesses,
			last_played = excluded.last_played,
			last_word_date = excluded.last_word_date,
			last_game_result = excluded.last_game_result,
			updated_at = CURRENT_TIMESTAMP
	`

	_, err := s.db.Exec(query,
		stats.Username,
		stats.SSHKeyFingerprint,
		stats.GamesPlayed,
		stats.GamesWon,
		stats.GamesLost,
		stats.CurrentStreak,
		stats.MaxStreak,
		stats.GuessDistribution[0],
		stats.GuessDistribution[1],
		stats.GuessDistribution[2],
		stats.GuessDistribution[3],
		stats.GuessDistribution[4],
		stats.GuessDistribution[5],
		stats.TotalGuesses,
		stats.LastPlayed,
		stats.LastWordDate,
		stats.LastGameResult,
	)

	if err != nil {
		return fmt.Errorf("failed to save user stats: %w", err)
	}

	s.logger.Debug("Successfully saved user stats", "username", stats.Username, "ssh_key_fingerprint", stats.SSHKeyFingerprint)
	return nil
}

// GetAverageGuesses calculates the average number of guesses for a user
func (stats *UserStats) GetAverageGuesses() float64 {
	if stats.GamesWon == 0 {
		return 0
	}
	return float64(stats.TotalGuesses) / float64(stats.GamesWon)
}

// GetWinRate calculates the win rate percentage
func (stats *UserStats) GetWinRate() float64 {
	if stats.GamesPlayed == 0 {
		return 0
	}
	return float64(stats.GamesWon) / float64(stats.GamesPlayed) * 100
}

// Close closes the database connection
func (s *Store) Close() error {
	return s.db.Close()
}
