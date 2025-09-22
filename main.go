package main

import (
	"dazedtrader/models"
	"fmt"
	"os"

	tea "github.com/charmbracelet/bubbletea"
	"golang.org/x/term"
)

func main() {
	model := models.NewAppModel()

	// Check if we have a proper TTY
	var p *tea.Program
	if term.IsTerminal(int(os.Stdin.Fd())) {
		// We have a proper terminal, use alt screen
		p = tea.NewProgram(model, tea.WithAltScreen())
	} else {
		// No proper terminal, run without alt screen
		p = tea.NewProgram(model)
	}

	if _, err := p.Run(); err != nil {
		fmt.Printf("Error running program: %v", err)
		os.Exit(1)
	}
}