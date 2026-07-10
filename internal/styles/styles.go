package styles

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"
)

var (
	HeaderStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("#FFFFFF")).
			Background(lipgloss.Color("#3498DB")).
			Padding(0, 1)

	SuccessStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("#2ECC71"))

	ErrorStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("#E74C3C"))

	WarnStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#F39C12"))

	InfoStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#3498DB"))

	FaintStyle = lipgloss.NewStyle().
			Faint(true)

	FileStatStyle = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color("#7F8C8D")).
			Padding(0, 1)

	Separator = FaintStyle.Render(strings.Repeat("─", 60))

)

func FileHeader(current, total int, input, output string) string {
	s := fmt.Sprintf(" [%d/%d] %s > %s ",
		current, total, input, output)
	return HeaderStyle.Render(s)
}

func StatBox(title string, items ...string) string {
	var b strings.Builder
	b.WriteString(lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("#3498DB")).Render(title))
	b.WriteString("\n")
	for _, item := range items {
		b.WriteString("  " + item + "\n")
	}
	return FileStatStyle.Render(b.String())
}
