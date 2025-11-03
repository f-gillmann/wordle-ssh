package ui

import (
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/log"
	"github.com/f-gillmann/wordle-ssh/internal/stats"
	"github.com/f-gillmann/wordle-ssh/internal/ui/models"
)

type AppState int

const (
	AppStateMenu AppState = iota
	AppStateGame
	AppStateStats
	AppStateAlreadyPlayed
	AppStateDeleteData
)

type AppModel struct {
	menu              models.MenuModel
	game              models.GameModel
	statsView         models.StatsModel
	alreadyPlayedView models.AlreadyPlayedModel
	deleteDataView    models.DeleteDataModel
	state             AppState
	targetWord        string
	wordDate          string
	username          string
	sshKeyFingerprint string
	statsStore        *stats.Store
	hasPlayedToday    bool
	hasUserData       bool
	motd              string
	logger            *log.Logger
}

func NewAppModel(targetWord string, wordDate string, username string, sshKeyFingerprint string, statsStore *stats.Store, hasPlayedToday bool, motd string, logger *log.Logger) AppModel {
	// Check if user has any data
	hasUserData := false
	if userStats, err := statsStore.GetUserStats(username, sshKeyFingerprint); err == nil && userStats.GamesPlayed > 0 {
		hasUserData = true
	}

	return AppModel{
		menu:              models.NewMenuModel(hasUserData, motd),
		state:             AppStateMenu,
		targetWord:        targetWord,
		wordDate:          wordDate,
		username:          username,
		sshKeyFingerprint: sshKeyFingerprint,
		statsStore:        statsStore,
		hasPlayedToday:    hasPlayedToday,
		hasUserData:       hasUserData,
		motd:              motd,
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
		m.menu = menuModel.(models.MenuModel)

		// Check if we should transition to game
		if m.menu.GetState() == models.MenuStateGame {
			if m.hasPlayedToday {
				// User has already played today, load their result and show it
				userStats, err := m.statsStore.GetUserStats(m.username, m.sshKeyFingerprint)
				if err != nil {
					m.logger.Error("Failed to get user stats for already played", "error", err, "username", m.username)
					userStats = &stats.UserStats{Username: m.username, SSHKeyFingerprint: m.sshKeyFingerprint}
				}

				m.alreadyPlayedView = models.NewAlreadyPlayedModel(userStats.LastGameResult)
				m.state = AppStateAlreadyPlayed

				return m, m.alreadyPlayedView.Init()
			}
			m.game = models.NewGameModel(m.targetWord, m.logger)
			m.state = AppStateGame

			return m, m.game.Init()
		} else if m.menu.GetState() == models.MenuStateStats {
			// Load and show user stats
			userStats, err := m.statsStore.GetUserStats(m.username, m.sshKeyFingerprint)
			if err != nil {
				// Log error but continue with empty stats
				m.logger.Error("Failed to get user stats", "error", err, "username", m.username)
				userStats = &stats.UserStats{Username: m.username, SSHKeyFingerprint: m.sshKeyFingerprint}
			}

			m.statsView = models.NewStatsModel(userStats)
			m.state = AppStateStats

			return m, m.statsView.Init()
		} else if m.menu.GetState() == models.MenuStateDeleteData {
			// Load user stats and show delete data confirmation
			userStats, err := m.statsStore.GetUserStats(m.username, m.sshKeyFingerprint)
			if err != nil {
				m.logger.Error("Failed to get user stats for delete data", "error", err, "username", m.username)
				userStats = &stats.UserStats{Username: m.username, SSHKeyFingerprint: m.sshKeyFingerprint}
			}

			m.deleteDataView = models.NewDeleteDataModel(m.username, userStats)
			m.state = AppStateDeleteData

			return m, m.deleteDataView.Init()
		} else if m.menu.GetState() == models.MenuStateExit {
			return m, tea.Quit
		}

		return m, cmd

	case AppStateGame:
		var cmd tea.Cmd
		gameModel, cmd := m.game.Update(msg)
		m.game = gameModel.(models.GameModel)

		// Check if game ended and record stats
		if m.game.GetState() == models.GameStateWon {
			// Record win with number of guesses and game result
			guesses := m.game.GetGuessCount()
			gameResultJSON := m.game.GetGameResultJSON()

			if err := m.statsStore.RecordWin(m.username, m.sshKeyFingerprint, guesses, m.wordDate, gameResultJSON); err != nil {
				m.logger.Error("Failed to record win", "error", err, "username", m.username)
			} else {
				// Mark that user has played today
				m.hasPlayedToday = true
				m.hasUserData = true
			}
		} else if m.game.GetState() == models.GameStateLost {
			// Record loss with game result
			gameResultJSON := m.game.GetGameResultJSON()

			if err := m.statsStore.RecordLoss(m.username, m.sshKeyFingerprint, m.wordDate, gameResultJSON); err != nil {
				m.logger.Error("Failed to record loss", "error", err, "username", m.username)
			} else {
				// Mark that user has played today
				m.hasPlayedToday = true
				m.hasUserData = true
			}
		}

		// Check if we should return to menu or quit
		if m.game.GetState() == models.GameStateMenu {
			m.menu = models.NewMenuModel(m.hasUserData, m.motd)
			m.state = AppStateMenu
			return m, m.menu.Init()
		} else if m.game.GetState() == models.GameStateQuit {
			return m, tea.Quit
		}

		return m, cmd

	case AppStateStats:
		var cmd tea.Cmd
		statsModel, cmd := m.statsView.Update(msg)
		m.statsView = statsModel.(models.StatsModel)

		// Check if any key was pressed to return to menu
		if _, ok := msg.(tea.KeyMsg); ok {
			// Any key returns to menu
			m.menu = models.NewMenuModel(m.hasUserData, m.motd)
			m.state = AppStateMenu
			return m, m.menu.Init()
		}

		return m, cmd

	case AppStateAlreadyPlayed:
		var cmd tea.Cmd
		alreadyPlayedModel, cmd := m.alreadyPlayedView.Update(msg)
		m.alreadyPlayedView = alreadyPlayedModel.(models.AlreadyPlayedModel)

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
			m.menu = models.NewMenuModel(m.hasUserData, m.motd)
			m.state = AppStateMenu
			return m, m.menu.Init()
		}

		return m, cmd

	case AppStateDeleteData:
		var cmd tea.Cmd
		deleteDataModel, cmd := m.deleteDataView.Update(msg)
		m.deleteDataView = deleteDataModel.(models.DeleteDataModel)

		// Check if user confirmed deletion
		if m.deleteDataView.GetState() == models.DeleteDataStateDeleted {
			// Actually delete the data
			if err := m.statsStore.DeleteUserData(m.username, m.sshKeyFingerprint); err != nil {
				m.logger.Error("Failed to delete user data", "error", err, "username", m.username)
			} else {
				m.logger.Info("Successfully deleted user data", "username", m.username)
				// Reset hasPlayedToday and hasUserData flags since data is deleted
				m.hasPlayedToday = false
				m.hasUserData = false
			}
		}

		// Check if we should return to menu
		if m.deleteDataView.GetState() == models.DeleteDataStateMenu {
			m.menu = models.NewMenuModel(m.hasUserData, m.motd)
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
	case AppStateDeleteData:
		return m.deleteDataView.View()
	default:
		return ""
	}
}
