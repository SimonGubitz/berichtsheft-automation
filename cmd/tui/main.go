package main

import (
	"context"
	"errors"
	"fmt"
	"os"

	"berichtsheft/internal/config"
	"berichtsheft/internal/logging"
	"berichtsheft/internal/ui"

	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
)

func main() {

	closer, err := logging.SetupDefault(logging.Dev)
	if err != nil {
		fmt.Println(ConsoleRed(err.Error()))
		return
	}
	defer closer.Close()

	ctx := context.Background()

	conf, err := config.GetConfig()
	if err != nil && !errors.Is(err, config.ErrNoConfigFile) {
		fmt.Println(ConsoleRed("Error starting program: " + err.Error()))
		os.Exit(1)
	}

	p := tea.NewProgram(ui.NewApplication(ui.NewRoute("homepage", ui.NewHomePage(conf))), tea.WithContext(ctx))
	if _, err := p.Run(); err != nil {
		fmt.Println(ConsoleRed("Error running program: " + err.Error()))
		os.Exit(1)
	}
}

func ConsoleRed(str string) string {
	style := lipgloss.NewStyle().Foreground(lipgloss.Red)
	return style.Render(str)
}
