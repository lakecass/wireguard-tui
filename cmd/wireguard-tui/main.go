package main

import (
	"flag"
	"fmt"
	"os"

	"wireguard-tui/internal/ui"
	"wireguard-tui/internal/wg"

	tea "github.com/charmbracelet/bubbletea"
)

func main() {
	// Parse flags
	useMock := flag.Bool("mock", false, "Use mock data (for development/demo)")
	flag.Parse()

	var client wg.Client
	if *useMock {
		client = wg.NewMockClient()
	} else {
		client = wg.NewLinuxClient()
	}

	m := ui.NewModel(client)
	p := tea.NewProgram(m, tea.WithAltScreen())
	if _, err := p.Run(); err != nil {
		fmt.Fprintf(os.Stderr, "Error starting program: %v\n", err)
		os.Exit(1)
	}
}
