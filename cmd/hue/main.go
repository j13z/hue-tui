package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/angristan/hue-tui/internal/config"
	"github.com/angristan/hue-tui/internal/tui"
	tea "github.com/charmbracelet/bubbletea"
)

func main() {
	demoMode := flag.Bool("demo", os.Getenv("HUE_DEMO") != "", "Use demo bridge")
	group := flag.String("group", "", "Comma-separated Hue room/zone IDs (v1 numeric IDs or v2 UUIDs)")
	flag.Parse()

	if *demoMode {
		fmt.Fprintln(os.Stderr, "[hue] Demo mode enabled")
	}

	// Load configuration
	cfg, err := config.Load()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error loading config: %v\n", err)
		os.Exit(1)
	}

	// Create and run the application
	model := tui.NewModel(cfg, *demoMode)
	model.SetGroupScope(*group)
	p := tea.NewProgram(
		model,
		tea.WithAltScreen(),
		tea.WithMouseCellMotion(),
	)

	if _, err := p.Run(); err != nil {
		fmt.Fprintf(os.Stderr, "Error running app: %v\n", err)
		os.Exit(1)
	}
}
