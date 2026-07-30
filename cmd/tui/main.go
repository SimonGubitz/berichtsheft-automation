package main

import (
	"context"
	"errors"
	"fmt"
	"os"

	"berichtsheft/internal/config"
	"berichtsheft/internal/ui/tui"

	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
)

func main() {

	ctx := context.Background()

	var confClient config.FileConfig
	conf, err := confClient.GetConfig()
	if err != nil && !errors.Is(err, config.ErrNoConfigFile) {
		fmt.Println(ConsoleRed("Error starting program: " + err.Error()))
		os.Exit(1)
	}

	m := tui.NewApplication(ctx, conf)
	p := tea.NewProgram(m, tea.WithContext(ctx))
	if _, err := p.Run(); err != nil {
		fmt.Println(ConsoleRed("Error running program: " + err.Error()))
		os.Exit(1)
	}
}

func ConsoleRed(str string) string {
	style := lipgloss.NewStyle().Foreground(lipgloss.Red)
	return style.Render(str)
}
