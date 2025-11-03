package models

import (
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/f-gillmann/wordle-ssh/internal/stats"
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
	username  string
	userStats *stats.UserStats
	input     string
	state     DeleteDataState
	err       error
}

func NewDeleteDataModel(username string, userStats *stats.UserStats) DeleteDataModel {
	return DeleteDataModel{
		username:  username,
		userStats: userStats,
		input:     "",
		state:     DeleteDataStateConfirm,
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
		// Display all the data that will be deleted
		s := styles.MenuTitleWarnStyle.Render("Data to be deleted:")
		s += "\n\n"

		if m.userStats != nil && m.userStats.GamesPlayed > 0 {
			s += fmt.Sprintf("  Username:             %s\n", m.userStats.Username)
			s += fmt.Sprintf("  SSH Key Fingerprint:  %s\n", m.userStats.SSHKeyFingerprint)
			s += fmt.Sprintf("  Games Played:         %d\n", m.userStats.GamesPlayed)
			s += fmt.Sprintf("  Games Won:            %d\n", m.userStats.GamesWon)
			s += fmt.Sprintf("  Games Lost:           %d\n", m.userStats.GamesLost)
			s += fmt.Sprintf("  Win Rate:             %.1f%%\n", m.userStats.GetWinRate())
			s += fmt.Sprintf("  Current Streak:       %d\n", m.userStats.CurrentStreak)
			s += fmt.Sprintf("  Max Streak:           %d\n", m.userStats.MaxStreak)
			s += fmt.Sprintf("  Average Guesses:      %.2f\n", m.userStats.GetAverageGuesses())
			s += fmt.Sprintf("  Total Guesses:        %d\n", m.userStats.TotalGuesses)
			s += "\n"

			// Show guess distribution
			s += "  Guess Distribution:\n"
			for i := 0; i < 6; i++ {
				s += fmt.Sprintf("    %d: %d\n", i+1, m.userStats.GuessDistribution[i])
			}
			s += "\n"

			if !m.userStats.LastPlayed.IsZero() {
				s += fmt.Sprintf("  Last Played:          %s\n", m.userStats.LastPlayed.Format("2006-01-02 15:04:05"))
			}
			if m.userStats.LastWordDate != "" {
				s += fmt.Sprintf("  Last Word Date:       %s\n", m.userStats.LastWordDate)
			}
			if m.userStats.LastGameResult != "" {
				s += fmt.Sprintf("  Last Game Result:     %s\n", m.userStats.LastGameResult)
			}
		} else {
			s += "  No data found for this user.\n"
		}

		s += "\n"
		s += styles.ErrorStyle.Render("This action cannot be undone!")
		s += "\n\n"

		// Delete confirmation prompt
		s += fmt.Sprintf("To confirm deletion, type your username:\n")
		s += fmt.Sprintf("> %s█\n\n", m.input)

		if m.err != nil {
			s += styles.ErrorStyle.Render(fmt.Sprintf("✗ %s", m.err.Error()))
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
