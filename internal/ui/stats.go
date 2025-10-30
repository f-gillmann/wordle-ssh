package ui

import (
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/f-gillmann/wordle-ssh/internal/stats"
)

type StatsModel struct {
	stats  *stats.UserStats
	width  int
	height int
}

func NewStatsModel(userStats *stats.UserStats) StatsModel {
	return StatsModel{
		stats: userStats,
	}
}

func (m StatsModel) Init() tea.Cmd {
	return nil
}

func (m StatsModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "q", "esc", "enter":
			return m, tea.Quit
		}
	}
	return m, nil
}

func (m StatsModel) View() string {
	var s strings.Builder

	titleStyle := lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("86")).
		Padding(1, 0)

	statStyle := lipgloss.NewStyle().
		Padding(0, 2)

	labelStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("240"))

	valueStyle := lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("86"))

	s.WriteString(titleStyle.Render("Your Statistics"))
	s.WriteString("\n\n")

	// Main stats
	statsList := []struct {
		label string
		value string
	}{
		{"Games Played", fmt.Sprintf("%d", m.stats.GamesPlayed)},
		{"Games Won", fmt.Sprintf("%d", m.stats.GamesWon)},
		{"Games Lost", fmt.Sprintf("%d", m.stats.GamesLost)},
		{"Win Rate", fmt.Sprintf("%.1f%%", m.stats.GetWinRate())},
		{"Current Streak", fmt.Sprintf("%d", m.stats.CurrentStreak)},
		{"Max Streak", fmt.Sprintf("%d", m.stats.MaxStreak)},
		{"Average Guesses", fmt.Sprintf("%.2f", m.stats.GetAverageGuesses())},
	}

	for _, stat := range statsList {
		line := statStyle.Render(
			labelStyle.Render(stat.label+": ") +
				valueStyle.Render(stat.value),
		)
		s.WriteString(line)
		s.WriteString("\n")
	}

	// Guess distribution
	s.WriteString("\n")
	s.WriteString(titleStyle.Render("Guess Distribution"))
	s.WriteString("\n\n")

	maxCount := 0
	for _, count := range m.stats.GuessDistribution {
		if count > maxCount {
			maxCount = count
		}
	}

	if m.stats.GamesLost > maxCount {
		maxCount = m.stats.GamesLost
	}

	if maxCount == 0 {
		maxCount = 1
	}

	barWidth := 30
	for i, count := range m.stats.GuessDistribution {
		barLength := 0
		if count > 0 {
			barLength = (count * barWidth) / maxCount
			if barLength == 0 {
				barLength = 1
			}
		}

		bar := strings.Repeat("█", barLength)
		s.WriteString(statStyle.Render(
			labelStyle.Render(fmt.Sprintf("%d: ", i+1)) +
				lipgloss.NewStyle().Foreground(lipgloss.Color("86")).Render(bar) +
				valueStyle.Render(fmt.Sprintf(" %d", count)),
		))
		s.WriteString("\n")
	}

	// Add games lost (X)
	lossBarLength := 0
	if m.stats.GamesLost > 0 {
		lossBarLength = (m.stats.GamesLost * barWidth) / maxCount
		if lossBarLength == 0 {
			lossBarLength = 1
		}
	}
	lossBar := strings.Repeat("█", lossBarLength)
	s.WriteString(statStyle.Render(
		labelStyle.Render("X: ") +
			lipgloss.NewStyle().Foreground(lipgloss.Color("196")).Render(lossBar) +
			valueStyle.Render(fmt.Sprintf(" %d", m.stats.GamesLost)),
	))
	s.WriteString("\n")

	s.WriteString("\n")
	s.WriteString(HelpStyle.Render("Press any key to return"))

	return s.String()
}
