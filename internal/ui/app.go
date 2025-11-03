package ui

import (
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/log"
	"github.com/f-gillmann/wordle-ssh/internal/stats"
)

type AppState int

const (
	AppStateMenu AppState = iota
	AppStateGame
	AppStateStats
	AppStateAlreadyPlayed
)

type AppModel struct {
	menu              MenuModel
	game              GameModel
	statsView         StatsModel
	alreadyPlayedView AlreadyPlayedModel
	state             AppState
	targetWord        string
	wordDate          string
	username          string
	sshKeyFingerprint string
	statsStore        *stats.Store
	hasPlayedToday    bool
	logger            *log.Logger
}

func NewAppModel(targetWord string, wordDate string, username string, sshKeyFingerprint string, statsStore *stats.Store, hasPlayedToday bool, logger *log.Logger) AppModel {
	return AppModel{
		menu:              NewMenuModel(),
		state:             AppStateMenu,
		targetWord:        targetWord,
		wordDate:          wordDate,
		username:          username,
		sshKeyFingerprint: sshKeyFingerprint,
		statsStore:        statsStore,
		hasPlayedToday:    hasPlayedToday,
		logger:            logger,
	}
}

func (m AppModel) Init() tea.Cmd {
	return nil
}

func (m AppModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch m.state {
	case AppStateMenu:
		var cmd tea.Cmd
		menuModel, cmd := m.menu.Update(msg)
		m.menu = menuModel.(MenuModel)

		// Check if we should transition to game
		if m.menu.GetState() == MenuStateGame {
			if m.hasPlayedToday {
				// User has already played today, load their result and show it
				userStats, err := m.statsStore.GetUserStats(m.username, m.sshKeyFingerprint)
				if err != nil {
					m.logger.Error("Failed to get user stats for already played", "error", err, "username", m.username)
					userStats = &stats.UserStats{Username: m.username, SSHKeyFingerprint: m.sshKeyFingerprint}
				}

				m.alreadyPlayedView = NewAlreadyPlayedModel(userStats.LastGameResult)
				m.state = AppStateAlreadyPlayed

				return m, m.alreadyPlayedView.Init()
			}
			m.game = NewGameModel(m.targetWord, m.logger)
			m.state = AppStateGame

			return m, m.game.Init()
		} else if m.menu.GetState() == MenuStateStats {
			// Load and show user stats
			userStats, err := m.statsStore.GetUserStats(m.username, m.sshKeyFingerprint)
			if err != nil {
				// Log error but continue with empty stats
				m.logger.Error("Failed to get user stats", "error", err, "username", m.username)
				userStats = &stats.UserStats{Username: m.username, SSHKeyFingerprint: m.sshKeyFingerprint}
			}

			m.statsView = NewStatsModel(userStats)
			m.state = AppStateStats

			return m, m.statsView.Init()
		} else if m.menu.GetState() == MenuStateExit {
			return m, tea.Quit
		}

		return m, cmd

	case AppStateGame:
		var cmd tea.Cmd
		gameModel, cmd := m.game.Update(msg)
		m.game = gameModel.(GameModel)

		// Check if game ended and record stats
		if m.game.GetState() == GameStateWon {
			// Record win with number of guesses and game result
			guesses := len(m.game.guesses)
			gameResultJSON := m.game.GetGameResultJSON()

			if err := m.statsStore.RecordWin(m.username, m.sshKeyFingerprint, guesses, m.wordDate, gameResultJSON); err != nil {
				m.logger.Error("Failed to record win", "error", err, "username", m.username)
			} else {
				// Mark that user has played today
				m.hasPlayedToday = true
			}
		} else if m.game.GetState() == GameStateLost {
			// Record loss with game result
			gameResultJSON := m.game.GetGameResultJSON()

			if err := m.statsStore.RecordLoss(m.username, m.sshKeyFingerprint, m.wordDate, gameResultJSON); err != nil {
				m.logger.Error("Failed to record loss", "error", err, "username", m.username)
			} else {
				// Mark that user has played today
				m.hasPlayedToday = true
			}
		}

		// Check if we should return to menu or quit
		if m.game.GetState() == GameStateMenu {
			m.menu = NewMenuModel()
			m.state = AppStateMenu
			return m, m.menu.Init()
		} else if m.game.GetState() == GameStateQuit {
			return m, tea.Quit
		}

		return m, cmd

	case AppStateStats:
		var cmd tea.Cmd
		statsModel, cmd := m.statsView.Update(msg)
		m.statsView = statsModel.(StatsModel)

		// Check if any key was pressed to return to menu
		if _, ok := msg.(tea.KeyMsg); ok {
			// Any key returns to menu
			m.menu = NewMenuModel()
			m.state = AppStateMenu
			return m, m.menu.Init()
		}

		return m, cmd

	case AppStateAlreadyPlayed:
		var cmd tea.Cmd
		alreadyPlayedModel, cmd := m.alreadyPlayedView.Update(msg)
		m.alreadyPlayedView = alreadyPlayedModel.(AlreadyPlayedModel)

		// If quit command was issued (Ctrl+C), pass it through
		if cmd != nil {
			// Check if it's a quit command by checking the message
			if keyMsg, ok := msg.(tea.KeyMsg); ok {
				if keyMsg.String() == "ctrl+c" || keyMsg.String() == "q" {
					return m, cmd
				}
			}
		}

		// Check if any key was pressed to return to menu
		if _, ok := msg.(tea.KeyMsg); ok {
			// Any other key returns to menu
			m.menu = NewMenuModel()
			m.state = AppStateMenu
			return m, m.menu.Init()
		}

		return m, cmd

	default:
		return m, nil
	}
}

func (m AppModel) View() string {
	switch m.state {
	case AppStateMenu:
		return m.menu.View()
	case AppStateGame:
		return m.game.View()
	case AppStateStats:
		return m.statsView.View()
	case AppStateAlreadyPlayed:
		return m.alreadyPlayedView.View()
	default:
		return ""
	}
}
