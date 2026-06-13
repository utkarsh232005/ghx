package main

import (
	"fmt"
	"os"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/KDM-cli/ghx/internal/app"
)

func main() {
	application := app.New()

	p := tea.NewProgram(
		application,
		tea.WithAltScreen(),
		tea.WithMouseCellMotion(),
	)

	if _, err := p.Run(); err != nil {
		fmt.Printf("Error running ghx: %v\n", err)
		os.Exit(1)
	}
}
