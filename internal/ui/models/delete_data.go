package models

import (
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/f-gillmann/wordle-ssh/internal/ui/styles"
)

type DeleteDataState int

const (
	DeleteDataStateConfirm DeleteDataState = iota
	DeleteDataStateDeleted
	DeleteDataStateCancelled
	DeleteDataStateMenu
)

type DeleteDataModel struct {
	username string
	input    string
	state    DeleteDataState
	err      error
}

func NewDeleteDataModel(username string) DeleteDataModel {
	return DeleteDataModel{
		username: username,
		input:    "",
		state:    DeleteDataStateConfirm,
	}
}

func (m DeleteDataModel) Init() tea.Cmd {
	return nil
}

func (m DeleteDataModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch m.state {
	case DeleteDataStateConfirm:
		switch msg := msg.(type) {
		case tea.KeyMsg:
			switch msg.String() {
			case "ctrl+c", "q", "esc":
				m.state = DeleteDataStateCancelled
				return m, nil

			case "enter":
				// Check if input matches username
				if strings.TrimSpace(m.input) == m.username {
					m.state = DeleteDataStateDeleted
					return m, nil
				} else {
					m.err = fmt.Errorf("username does not match")
					return m, nil
				}

			case "backspace":
				if len(m.input) > 0 {
					m.input = m.input[:len(m.input)-1]
				}

			default:
				// Only accept printable characters
				if len(msg.String()) == 1 {
					m.input += msg.String()
				}
			}
		}

	case DeleteDataStateDeleted, DeleteDataStateCancelled:
		switch msg.(type) {
		case tea.KeyMsg:
			m.state = DeleteDataStateMenu
			return m, nil
		}

	case DeleteDataStateMenu:
		return m, nil
	}

	return m, nil
}

func (m DeleteDataModel) View() string {
	switch m.state {
	case DeleteDataStateConfirm:
		s := styles.MenuTitleWarnStyle.Render("Delete All Data")
		s += "\n\n"
		s += "This action will permanently delete all your game statistics.\n"
		s += "This cannot be undone!\n\n"
		s += fmt.Sprintf("To confirm, please type your username:\n%s\n", styles.MenuTitleStyle.Render(m.username))

		// Show input with cursor
		s += fmt.Sprintf("> %s█\n\n", m.input)

		if m.err != nil {
			s += styles.ErrorStyle.Render(fmt.Sprintf("%s", m.err.Error()))
			s += "\n\n"
		}

		s += styles.HelpStyle.Render("Enter to confirm | Esc/Q to cancel")
		return s

	case DeleteDataStateDeleted:
		s := styles.MenuTitleStyle.Render("✓ Data Deleted")
		s += "\n\n"
		s += styles.SuccessStyle.Render("All your game data has been permanently deleted.")
		s += "\n\n"
		s += styles.HelpStyle.Render("Press any key to return to menu...")
		return s

	case DeleteDataStateCancelled:
		s := styles.MenuTitleStyle.Render("Cancelled")
		s += "\n\n"
		s += "No data was deleted.\n\n"
		s += styles.HelpStyle.Render("Press any key to return to menu...")
		return s

	default:
		return ""
	}
}

func (m DeleteDataModel) GetState() DeleteDataState {
	return m.state
}
