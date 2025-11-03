package models

import (
	"fmt"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/f-gillmann/wordle-ssh/internal/ui/styles"
)

type MenuState int

const (
	MenuStateMain MenuState = iota
	MenuStateGame
	MenuStateStats
	MenuStateDeleteData
	MenuStateExit
)

type MenuItem struct {
	Title       string
	Description string
}

type MenuModel struct {
	choices     []MenuItem
	cursor      int
	selected    int
	state       MenuState
	hasUserData bool
}

func NewMenuModel(hasUserData bool) MenuModel {
	choices := []MenuItem{
		{Title: "Play Wordle", Description: "Start a new game"},
		{Title: "View Stats", Description: "View your statistics"},
	}

	// Only add "Delete My Data" option if user has data
	if hasUserData {
		choices = append(choices, MenuItem{Title: "Delete My Data", Description: "Delete all your game data"})
	}

	choices = append(choices, MenuItem{Title: "Exit", Description: "Quit the application"})

	return MenuModel{
		choices:     choices,
		cursor:      0,
		selected:    -1,
		state:       MenuStateMain,
		hasUserData: hasUserData,
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
			selectedTitle := m.choices[m.cursor].Title
			switch selectedTitle {
			case "Play Wordle":
				m.state = MenuStateGame
			case "View Stats":
				m.state = MenuStateStats
			case "Delete My Data":
				m.state = MenuStateDeleteData
			case "Exit":
				m.state = MenuStateExit
				return m, tea.Quit
			}

			return m, nil
		}
	}

	return m, nil
}

func (m MenuModel) View() string {
	s := styles.MenuTitleStyle.Render("github.com/f-gillmann/wordle-ssh")
	s += "\n\n"

	for i, choice := range m.choices {
		cursor := " "
		if m.cursor == i {
			cursor = ">"
			s += styles.SelectedMenuItemStyle.Render(fmt.Sprintf("%s %s", cursor, choice.Title))
		} else {
			s += styles.MenuItemStyle.Render(fmt.Sprintf("%s %s", cursor, choice.Title))
		}

		s += "\n"
	}

	s += "\n"
	s += styles.HelpStyle.Render("↑/↓/j/k to navigate | Enter to select | Q/Ctrl+C to quit")

	return s
}

func (m MenuModel) GetState() MenuState {
	return m.state
}
