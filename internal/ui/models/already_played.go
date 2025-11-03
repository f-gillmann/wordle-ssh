package models

import (
	"encoding/json"
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/f-gillmann/wordle-ssh/internal/ui/styles"
)

type AlreadyPlayedModel struct {
	gameResult string
	won        bool
	guesses    int
}

func NewAlreadyPlayedModel(gameResultJSON string) AlreadyPlayedModel {
	won := false
	guesses := 0

	if gameResultJSON != "" {
		var result struct {
			W bool     `json:"w"`
			G []string `json:"g"`
		}

		if err := json.Unmarshal([]byte(gameResultJSON), &result); err == nil {
			won = result.W
			guesses = len(result.G)
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
			W bool     `json:"w"`
			G []string `json:"g"`
		}

		if err := json.Unmarshal([]byte(m.gameResult), &result); err == nil && len(result.G) > 0 {
			won := result.W
			guesses := result.G

			// Render the squares from compact format
			for _, guess := range guesses {
				var tiles []string
				// Parse compact format: "L1S1L2S2L3S3L4S4L5S5"
				for i := 0; i < len(guess); i += 2 {
					if i+1 >= len(guess) {
						break
					}

					state := guess[i+1]

					var style lipgloss.Style
					switch state {
					case 'c':
						style = styles.TileStyleCorrect
					case 'p':
						style = styles.TileStylePresent
					case 'a':
						style = styles.TileStyleAbsent
					default:
						style = styles.TileStyleEmpty
					}

					tiles = append(tiles, style.Render("*"))
				}

				if len(tiles) > 0 {
					s.WriteString(lipgloss.JoinHorizontal(lipgloss.Top, tiles...))
					s.WriteString("\n")
				}
			}

			// Render empty rows for remaining guesses
			remainingGuesses := MaxGuesses - len(guesses)
			for i := 0; i < remainingGuesses; i++ {
				var emptyTiles []string
				for j := 0; j < WordLength; j++ {
					emptyTiles = append(emptyTiles, styles.TileStyleEmpty.Render(" "))
				}
				s.WriteString(lipgloss.JoinHorizontal(lipgloss.Top, emptyTiles...))
				s.WriteString("\n")
			}

			s.WriteString("\n")

			if won {
				s.WriteString(styles.SuccessStyle.Render(fmt.Sprintf("You won in %d guesses!", len(guesses))))
			} else {
				s.WriteString(styles.ErrorStyle.Render("You didn't get it this time."))
			}
		}
	} else {
		s.WriteString("Come back tomorrow to play again!")
	}

	s.WriteString("\n\n")
	s.WriteString(styles.HelpStyle.Render("Any key to return | Q/Ctrl+C to quit"))

	return s.String()
}

func (m AlreadyPlayedModel) GetShouldReturnToMenu() bool {
	return false
}
