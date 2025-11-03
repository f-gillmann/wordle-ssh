package ui

import (
	"fmt"

	tea "github.com/charmbracelet/bubbletea"
)

type MenuState int

const (
	MenuStateMain MenuState = iota
	MenuStateGame
	MenuStateStats
	MenuStateExit
)

type MenuItem struct {
	Title       string
	Description string
}

type MenuModel struct {
	choices  []MenuItem
	cursor   int
	selected int
	state    MenuState
}

func NewMenuModel() MenuModel {
	choices := []MenuItem{
		{Title: "Play Wordle", Description: "Start a new game"},
		{Title: "View Stats", Description: "View your statistics"},
		{Title: "Exit", Description: "Quit the application"},
	}

	return MenuModel{
		choices:  choices,
		cursor:   0,
		selected: -1,
		state:    MenuStateMain,
	}
}

func (m MenuModel) Init() tea.Cmd {
	return nil
}

func (m MenuModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "q":
			m.state = MenuStateExit
			return m, tea.Quit

		case "up", "k":
			if m.cursor > 0 {
				m.cursor--
			}

		case "down", "j":
			if m.cursor < len(m.choices)-1 {
				m.cursor++
			}

		case "enter":
			m.selected = m.cursor

			// Determine action based on menu choice
			switch m.cursor {
			case 0: // Play Wordle
				m.state = MenuStateGame
			case 1: // View Stats
				m.state = MenuStateStats
			case 2: // Exit
				m.state = MenuStateExit
				return m, tea.Quit
			}

			return m, nil
		}
	}

	return m, nil
}

func (m MenuModel) View() string {
	s := MenuTitleStyle.Render("github.com/f-gillmann/wordle-ssh")
	s += "\n\n"

	for i, choice := range m.choices {
		cursor := " "
		if m.cursor == i {
			cursor = ">"
			s += SelectedMenuItemStyle.Render(fmt.Sprintf("%s %s", cursor, choice.Title))
		} else {
			s += MenuItemStyle.Render(fmt.Sprintf("%s %s", cursor, choice.Title))
		}

		s += "\n"
	}

	s += "\n"
	s += HelpStyle.Render("↑/↓/j/k to navigate | Enter to select | Q/Ctrl+C to quit")

	return s
}

func (m MenuModel) GetState() MenuState {
	return m.state
}
