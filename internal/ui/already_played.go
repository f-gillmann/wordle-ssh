package ui

import (
	"encoding/json"
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type AlreadyPlayedModel struct {
	gameResult string
	won        bool
	guesses    int
}

func NewAlreadyPlayedModel(gameResultJSON string) AlreadyPlayedModel {
	var result struct {
		Won     bool       `json:"won"`
		Guesses [][]string `json:"guesses"`
	}

	won := false
	guesses := 0

	if gameResultJSON != "" {
		if err := json.Unmarshal([]byte(gameResultJSON), &result); err == nil {
			won = result.Won
			guesses = len(result.Guesses)
		}
	}

	return AlreadyPlayedModel{
		gameResult: gameResultJSON,
		won:        won,
		guesses:    guesses,
	}
}

func (m AlreadyPlayedModel) Init() tea.Cmd {
	return nil
}

func (m AlreadyPlayedModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "q":
			return m, tea.Quit
		case "esc", "enter":
			// Return to menu
			return m, nil
		}
	}
	return m, nil
}

func (m AlreadyPlayedModel) View() string {
	var s strings.Builder

	titleStyle := lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("214")).
		Padding(1, 0)

	s.WriteString(titleStyle.Render("You've already played today!"))
	s.WriteString("\n")

	// Parse and display the game result
	if m.gameResult != "" {
		var result struct {
			Won     bool       `json:"won"`
			Guesses [][]string `json:"guesses"` // Each guess is array of [letter, state]
		}

		if err := json.Unmarshal([]byte(m.gameResult), &result); err == nil {
			// Render the squares
			for _, guess := range result.Guesses {
				var tiles []string
				for i := 0; i < len(guess); i += 2 {
					if i+1 >= len(guess) {
						break
					}

					state := guess[i+1]

					var style lipgloss.Style
					switch state {
					case "correct":
						style = TileStyleCorrect
					case "present":
						style = TileStylePresent
					case "absent":
						style = TileStyleAbsent
					default:
						style = TileStyleEmpty
					}

					tiles = append(tiles, style.Render("*"))
				}

				if len(tiles) > 0 {
					s.WriteString(lipgloss.JoinHorizontal(lipgloss.Top, tiles...))
					s.WriteString("\n")
				}
			}

			// Render empty rows for remaining guesses
			remainingGuesses := MaxGuesses - len(result.Guesses)
			for i := 0; i < remainingGuesses; i++ {
				var emptyTiles []string
				for j := 0; j < WordLength; j++ {
					emptyTiles = append(emptyTiles, TileStyleEmpty.Render(" "))
				}
				s.WriteString(lipgloss.JoinHorizontal(lipgloss.Top, emptyTiles...))
				s.WriteString("\n")
			}

			s.WriteString("\n")

			if result.Won {
				s.WriteString(SuccessStyle.Render(fmt.Sprintf("You won in %d guesses!", len(result.Guesses))))
			} else {
				s.WriteString(ErrorStyle.Render("You didn't get it this time."))
			}
		}
	} else {
		s.WriteString("Come back tomorrow to play again!")
	}

	s.WriteString("\n\n")
	s.WriteString(HelpStyle.Render("Any key to return | Q/Ctrl+C to quit"))

	return s.String()
}

func (m AlreadyPlayedModel) GetShouldReturnToMenu() bool {
	return false
}
