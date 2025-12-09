package render

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"
)

var (
	errorBoxStyle = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color("#EF4444")).
			Padding(0, 1).
			MarginTop(1)

	errorTitleStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("#EF4444"))

	errorMsgStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#FCA5A5"))
)

// RenderError renders an error with nice formatting
func RenderError(err error) string {
	var sb strings.Builder
	sb.WriteString(errorTitleStyle.Render("Error"))
	sb.WriteString("\n")
	sb.WriteString(errorMsgStyle.Render(err.Error()))
	return errorBoxStyle.Render(sb.String())
}

// RenderSuccess renders a success message
func RenderSuccess(msg string) string {
	return SuccessStyle.Render("✓ " + msg)
}

// RenderWarning renders a warning message
func RenderWarning(msg string) string {
	return WarningStyle.Render("⚠ " + msg)
}

// RenderInfo renders an info message
func RenderInfo(msg string) string {
	return SubtitleStyle.Render("ℹ " + msg)
}

// RenderHTTPStatus renders an HTTP status code
func RenderHTTPStatus(code int) string {
	var style lipgloss.Style
	switch {
	case code >= 200 && code < 300:
		style = SuccessStyle
	case code >= 400 && code < 500:
		style = WarningStyle
	case code >= 500:
		style = ErrorStyle
	default:
		style = SubtitleStyle
	}
	return style.Render(fmt.Sprintf("HTTP %d", code))
}
