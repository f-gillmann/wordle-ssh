package ui

import "github.com/charmbracelet/lipgloss"

var (
	// Color palette
	colorGreen       = lipgloss.Color("#6aaa64")
	colorYellow      = lipgloss.Color("#c9b458")
	colorGray        = lipgloss.Color("#787c7e")
	colorWhite       = lipgloss.Color("#ffffff")
	colorLightGray   = lipgloss.Color("#FAFAFA")
	colorPurple      = lipgloss.Color("#7D56F4")
	colorDarkGray    = lipgloss.Color("#626262")
	colorRed         = lipgloss.Color("#FF0000")
	colorBrightGreen = lipgloss.Color("#00FF00")

	TileStyleCorrect = lipgloss.NewStyle().
				Foreground(colorWhite).
				Padding(0, 1).
				Bold(true).
				Border(lipgloss.NormalBorder()).
				BorderForeground(colorGreen).
				Width(1).
				Align(lipgloss.Center)

	TileStylePresent = lipgloss.NewStyle().
				Foreground(colorWhite).
				Padding(0, 1).
				Bold(true).
				Border(lipgloss.NormalBorder()).
				BorderForeground(colorYellow).
				Width(1).
				Align(lipgloss.Center)

	TileStyleAbsent = lipgloss.NewStyle().
			Foreground(colorWhite).
			Padding(0, 1).
			Bold(true).
			Border(lipgloss.NormalBorder()).
			BorderForeground(colorGray).
			Width(1).
			Align(lipgloss.Center)

	TileStyleEmpty = lipgloss.NewStyle().
			Foreground(colorWhite).
			Padding(0, 1).
			Border(lipgloss.NormalBorder()).
			BorderForeground(colorDarkGray).
			Width(1).
			Align(lipgloss.Center)

	TileStyleInvalid = lipgloss.NewStyle().
				Foreground(colorWhite).
				Padding(0, 1).
				Bold(true).
				Border(lipgloss.NormalBorder()).
				BorderForeground(colorRed).
				Width(1).
				Align(lipgloss.Center)

	MenuTitleStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(colorLightGray).
			Border(lipgloss.DoubleBorder()).
			BorderForeground(colorLightGray).
			Padding(0, 4)

	MenuItemStyle = lipgloss.NewStyle().
			PaddingLeft(2).
			Foreground(colorLightGray)

	SelectedMenuItemStyle = lipgloss.NewStyle().
				PaddingLeft(1).
				Foreground(colorPurple).
				Bold(true)

	HelpStyle = lipgloss.NewStyle().
			Foreground(colorDarkGray)

	ErrorStyle = lipgloss.NewStyle().
			Foreground(colorRed).
			Bold(true)

	SuccessStyle = lipgloss.NewStyle().
			Foreground(colorBrightGreen).
			Bold(true)
)
