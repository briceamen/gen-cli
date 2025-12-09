package render

import "github.com/charmbracelet/lipgloss"

var (
	// Colors
	primaryColor   = lipgloss.Color("#7C3AED") // Purple
	secondaryColor = lipgloss.Color("#10B981") // Green
	errorColor     = lipgloss.Color("#EF4444") // Red
	mutedColor     = lipgloss.Color("#6B7280") // Gray
	borderColor    = lipgloss.Color("#374151") // Dark gray

	// Base styles
	TitleStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(primaryColor).
			MarginBottom(1)

	SubtitleStyle = lipgloss.NewStyle().
			Foreground(mutedColor).
			Italic(true)

	// Table styles
	TableHeaderStyle = lipgloss.NewStyle().
				Bold(true).
				Foreground(primaryColor).
				BorderStyle(lipgloss.NormalBorder()).
				BorderBottom(true).
				BorderForeground(borderColor).
				Padding(0, 1)

	TableCellStyle = lipgloss.NewStyle().
			Padding(0, 1)

	TableSelectedStyle = lipgloss.NewStyle().
				Background(lipgloss.Color("#1F2937")).
				Padding(0, 1)

	// Detail view styles
	KeyStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(secondaryColor).
			Width(20)

	ValueStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#F3F4F6"))

	// Status styles
	SuccessStyle = lipgloss.NewStyle().
			Foreground(secondaryColor).
			Bold(true)

	ErrorStyle = lipgloss.NewStyle().
			Foreground(errorColor).
			Bold(true)

	WarningStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#F59E0B")).
			Bold(true)

	// Box style for detail views
	BoxStyle = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(borderColor).
			Padding(1, 2)
)
