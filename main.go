package main

import (
	"dazedtrader/models"
	"fmt"
	"os"

	tea "github.com/charmbracelet/bubbletea"
)

func main() {
	model := models.NewAppModel()

	p := tea.NewProgram(model, tea.WithAltScreen())
	if _, err := p.Run(); err != nil {
		fmt.Printf("Error running program: %v", err)
		os.Exit(1)
	}
}