package ui

import (
	"fmt"
	"github.com/charmbracelet/lipgloss"
)

var (
	// Main styles
	TitleStyle = lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("#7D56F4")).
		Background(lipgloss.Color("#000000")).
		Padding(1, 2).
		Align(lipgloss.Center)

	MenuStyle = lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("#874BFD")).
		Padding(1, 2).
		MarginTop(1)

	SelectedStyle = lipgloss.NewStyle().
		Foreground(lipgloss.Color("#EE6FF8")).
		Bold(true)

	UnselectedStyle = lipgloss.NewStyle().
		Foreground(lipgloss.Color("#FAFAFA"))

	DisabledStyle = lipgloss.NewStyle().
		Foreground(lipgloss.Color("#666666"))

	HeaderStyle = lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("#FAFAFA")).
		Background(lipgloss.Color("#7D56F4")).
		Padding(0, 1)

	InfoStyle = lipgloss.NewStyle().
		BorderStyle(lipgloss.NormalBorder()).
		BorderTop(true).
		BorderForeground(lipgloss.Color("#874BFD"))

	// Data display styles
	ValueStyle = lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("#FAFAFA"))

	PositiveStyle = lipgloss.NewStyle().
		Foreground(lipgloss.Color("#04B575")).
		Bold(true)

	NegativeStyle = lipgloss.NewStyle().
		Foreground(lipgloss.Color("#FF5F87")).
		Bold(true)

	NeutralStyle = lipgloss.NewStyle().
		Foreground(lipgloss.Color("#FAFAFA"))

	// Table styles
	TableHeaderStyle = lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("#7D56F4")).
		Align(lipgloss.Center)

	TableRowStyle = lipgloss.NewStyle().
		Foreground(lipgloss.Color("#FAFAFA"))

	// Loading styles
	LoadingStyle = lipgloss.NewStyle().
		Foreground(lipgloss.Color("#FFA500")).
		Bold(true)

	// Input styles
	InputStyle = lipgloss.NewStyle().
		Foreground(lipgloss.Color("#FAFAFA")).
		Background(lipgloss.Color("#874BFD")).
		Padding(0, 1)
)

func FormatCurrency(value float64) string {
	if value >= 0 {
		return PositiveStyle.Render(fmt.Sprintf("+$%.2f", value))
	}
	return NegativeStyle.Render(fmt.Sprintf("-$%.2f", -value))
}

func FormatValue(value float64) string {
	return ValueStyle.Render(fmt.Sprintf("$%.2f", value))
}

func FormatPercentage(value float64) string {
	if value >= 0 {
		return PositiveStyle.Render(fmt.Sprintf("+%.2f%%", value))
	}
	return NegativeStyle.Render(fmt.Sprintf("%.2f%%", value))
}