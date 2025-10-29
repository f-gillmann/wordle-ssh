package ui

import (
	tea "github.com/charmbracelet/bubbletea"
)

type AppState int

const (
	AppStateMenu AppState = iota
	AppStateGame
)

type AppModel struct {
	menu       MenuModel
	game       GameModel
	state      AppState
	targetWord string
}

func NewAppModel(targetWord string) AppModel {
	return AppModel{
		menu:       NewMenuModel(),
		state:      AppStateMenu,
		targetWord: targetWord,
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
			m.game = NewGameModel(m.targetWord)
			m.state = AppStateGame
			return m, m.game.Init()
		} else if m.menu.GetState() == MenuStateExit {
			return m, tea.Quit
		}

		return m, cmd

	case AppStateGame:
		var cmd tea.Cmd
		gameModel, cmd := m.game.Update(msg)
		m.game = gameModel.(GameModel)

		// Check if we should return to menu or quit
		if m.game.GetState() == GameStateMenu {
			m.menu = NewMenuModel()
			m.state = AppStateMenu
			return m, m.menu.Init()
		} else if m.game.GetState() == GameStateQuit {
			return m, tea.Quit
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
	default:
		return ""
	}
}
