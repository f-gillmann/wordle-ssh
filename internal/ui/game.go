package ui

import (
	"encoding/json"
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/charmbracelet/log"
	"github.com/f-gillmann/wordle-ssh/internal/wordle"
)

const (
	MaxGuesses = 6
	WordLength = 5
)

type GameState int

const (
	GameStatePlaying GameState = iota
	GameStateWon
	GameStateLost
	GameStateMenu
	GameStateQuit
)

type GuessResult struct {
	Letter string
	State  LetterState
}

type LetterState int

const (
	LetterStateCorrect LetterState = iota // Green - correct position
	LetterStatePresent                    // Yellow - in word but wrong position
	LetterStateAbsent                     // Gray - not in word
)

type GameModel struct {
	targetWord   string
	guesses      []string
	currentGuess string
	guessResults [][]GuessResult
	state        GameState
	errorMessage string
	letterMap    map[rune]LetterState
	invalidWord  bool
	logger       *log.Logger
}

func NewGameModel(targetWord string, logger *log.Logger) GameModel {
	logger.Debug("Creating new game model", "targetWord", targetWord)

	return GameModel{
		targetWord:   strings.ToLower(targetWord),
		guesses:      []string{},
		currentGuess: "",
		guessResults: [][]GuessResult{},
		state:        GameStatePlaying,
		letterMap:    make(map[rune]LetterState),
		logger:       logger,
	}
}

func (m GameModel) Init() tea.Cmd {
	return nil
}

func (m GameModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c":
			m.logger.Debug("User quit game")
			m.state = GameStateQuit
			return m, tea.Quit

		case "esc":
			m.logger.Debug("User returned to menu")
			m.state = GameStateMenu
			return m, nil

		case "enter":
			if m.state != GameStatePlaying {
				// Return to menu if game is over
				m.logger.Debug("User returned to menu after game ended")
				m.state = GameStateMenu
				return m, nil
			}

			if len([]rune(m.currentGuess)) != WordLength {
				m.logger.Debug("Invalid guess length", "guess", m.currentGuess, "length", len([]rune(m.currentGuess)))
				m.errorMessage = fmt.Sprintf("Word must be %d letters\n", WordLength)
				return m, nil
			}

			// Validate the guess against the wordlist
			if !wordle.IsValidWord(strings.ToLower(m.currentGuess)) {
				m.logger.Debug("Invalid word attempted", "guess", m.currentGuess)
				m.invalidWord = true
				m.errorMessage = "Invalid word\n"
				return m, nil
			}

			// Process the guess
			m.logger.Info("Valid guess submitted", "guess", m.currentGuess, "attempt", len(m.guesses)+1)
			m.errorMessage = ""
			m.invalidWord = false
			m.guesses = append(m.guesses, m.currentGuess)
			result := m.evaluateGuess(m.currentGuess)
			m.guessResults = append(m.guessResults, result)

			// Update letter map
			for _, gr := range result {
				if len(gr.Letter) == 0 {
					continue // Skip empty letters
				}

				letter := rune(strings.ToLower(gr.Letter)[0])

				// Only update if it's better information than we had
				if existing, ok := m.letterMap[letter]; !ok || gr.State < existing {
					m.letterMap[letter] = gr.State
				}
			}

			// Check win condition
			if strings.ToLower(m.currentGuess) == m.targetWord {
				m.logger.Info("Game won", "attempts", len(m.guesses), "targetWord", m.targetWord)
				m.state = GameStateWon
			} else if len(m.guesses) >= MaxGuesses {
				m.logger.Info("Game lost", "attempts", len(m.guesses), "targetWord", m.targetWord)
				m.state = GameStateLost
			}

			m.currentGuess = ""
			return m, nil

		case "backspace":
			if len(m.currentGuess) > 0 {
				m.currentGuess = m.currentGuess[:len(m.currentGuess)-1]
				m.errorMessage = ""
				m.invalidWord = false
			}

		default:
			// Only accept letters
			if len(msg.String()) == 1 && msg.String()[0] >= 'a' && msg.String()[0] <= 'z' {
				if len([]rune(m.currentGuess)) < WordLength {
					m.currentGuess += strings.ToUpper(msg.String())
					m.errorMessage = ""
				}
			} else if len(msg.String()) == 1 && msg.String()[0] >= 'A' && msg.String()[0] <= 'Z' {
				if len([]rune(m.currentGuess)) < WordLength {
					m.currentGuess += msg.String()
					m.errorMessage = ""
				}
			}
		}
	}

	return m, nil
}

func (m GameModel) evaluateGuess(guess string) []GuessResult {
	guess = strings.ToLower(guess)
	result := make([]GuessResult, WordLength)
	targetLetters := []rune(m.targetWord)
	guessLetters := []rune(guess)
	used := make([]bool, WordLength)

	// Safety check: ensure guess is the correct length
	if len(guessLetters) != WordLength || len(targetLetters) != WordLength {
		// Return empty results if lengths don't match
		for i := 0; i < WordLength; i++ {
			result[i] = GuessResult{
				Letter: "",
				State:  LetterStateAbsent,
			}
		}

		return result
	}

	// Initialize all results with letters (will be updated with correct states)
	for i := 0; i < WordLength; i++ {
		result[i] = GuessResult{
			Letter: strings.ToUpper(string(guessLetters[i])),
			State:  LetterStateAbsent, // Default to absent, will be updated if correct or present
		}
	}

	// First pass: mark correct positions
	for i := 0; i < WordLength; i++ {
		if guessLetters[i] == targetLetters[i] {
			result[i].State = LetterStateCorrect
			used[i] = true
		}
	}

	// Second pass: mark present letters
	for i := 0; i < WordLength; i++ {
		if result[i].State == LetterStateCorrect {
			continue
		}

		for j := 0; j < WordLength; j++ {
			if !used[j] && guessLetters[i] == targetLetters[j] {
				result[i].State = LetterStatePresent
				used[j] = true
				break
			}
		}
	}

	return result
}

func (m GameModel) View() string {
	var s strings.Builder

	// Render previous guesses
	for i := 0; i < MaxGuesses; i++ {
		var tiles []string

		if i < len(m.guessResults) {
			// Render completed guess with colored boxes
			for _, result := range m.guessResults[i] {
				var style lipgloss.Style

				switch result.State {
				case LetterStateCorrect:
					style = TileStyleCorrect
				case LetterStatePresent:
					style = TileStylePresent
				case LetterStateAbsent:
					style = TileStyleAbsent
				default:
					style = TileStyleEmpty
				}

				tiles = append(tiles, style.Render(result.Letter))
			}
		} else if i == len(m.guesses) {
			// Render current guess being typed
			for j := 0; j < WordLength; j++ {
				if j < len([]rune(m.currentGuess)) {
					// Use red style if word is invalid
					style := TileStyleEmpty
					if m.invalidWord {
						style = TileStyleInvalid
					}

					tiles = append(tiles, style.Render(string([]rune(m.currentGuess)[j])))
				} else {
					tiles = append(tiles, TileStyleEmpty.Render(" "))
				}
			}
		} else {
			// Render empty row
			for j := 0; j < WordLength; j++ {
				tiles = append(tiles, TileStyleEmpty.Render(" "))
			}
		}

		s.WriteString(lipgloss.JoinHorizontal(lipgloss.Top, tiles...))
		s.WriteString("\n")
	}

	s.WriteString("\n")

	// Show game state messages
	switch m.state {
	case GameStateWon:
		s.WriteString(SuccessStyle.Render(fmt.Sprintf("Congratulations! You won in %d guesses!", len(m.guesses))))
		s.WriteString("\n\n")
		s.WriteString(HelpStyle.Render("Enter/Esc to menu | Ctrl+C to quit"))
	case GameStateLost:
		s.WriteString(ErrorStyle.Render(fmt.Sprintf("Game Over!")))
		s.WriteString("\n\n")
		s.WriteString(HelpStyle.Render("Enter/Esc to menu | Ctrl+C to quit"))
	case GameStatePlaying:
		if m.errorMessage != "" {
			s.WriteString(ErrorStyle.Render(m.errorMessage))
			s.WriteString("\n")
		}

		s.WriteString(HelpStyle.Render(fmt.Sprintf("Guess %d/%d", len(m.guesses)+1, MaxGuesses)))
		s.WriteString("\n\n")
		s.WriteString(HelpStyle.Render("Enter to submit | Backspace to delete | Esc to menu | Ctrl+C to quit"))
	default:
		// Unknown state, show playing instructions
		s.WriteString(HelpStyle.Render("Esc to menu | Ctrl+C to quit"))
	}

	return s.String()
}

func (m GameModel) GetState() GameState {
	return m.state
}

// GetGameResultJSON returns the game result as a JSON string for storage
func (m GameModel) GetGameResultJSON() string {
	type GameResultData struct {
		Won     bool       `json:"won"`
		Guesses [][]string `json:"guesses"` // Each guess is array of [letter, state]
	}

	result := GameResultData{
		Won:     m.state == GameStateWon,
		Guesses: [][]string{},
	}

	// Convert guess results to simplified format
	for _, guessResult := range m.guessResults {
		var guess []string
		for _, gr := range guessResult {
			var state string
			switch gr.State {
			case LetterStateCorrect:
				state = "correct"
			case LetterStatePresent:
				state = "present"
			case LetterStateAbsent:
				state = "absent"
			}

			guess = append(guess, gr.Letter, state)
		}

		result.Guesses = append(result.Guesses, guess)
	}

	jsonBytes, err := json.Marshal(result)
	if err != nil {
		m.logger.Error("Failed to marshal game result", "error", err)
		return ""
	}

	return string(jsonBytes)
}
