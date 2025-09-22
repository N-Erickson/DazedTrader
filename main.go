package main

import (
	"dazedtrader/models"
	"fmt"
	"log"
	"os"

	tea "github.com/charmbracelet/bubbletea"
	"golang.org/x/term"
)

func main() {
	// Set up logging to file
	logFile, err := os.OpenFile("/tmp/dazedtrader_main.log", os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
	if err != nil {
		fmt.Printf("Failed to open log file: %v", err)
	} else {
		log.SetOutput(logFile)
		log.Println("DazedTrader starting...")
	}

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
		log.Printf("Error running program: %v", err)
		fmt.Printf("Error running program: %v", err)
		os.Exit(1)
	}

	log.Println("DazedTrader exiting...")
	if logFile != nil {
		logFile.Close()
	}
}