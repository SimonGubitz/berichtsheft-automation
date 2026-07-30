package main

import (
	"context"
	"fmt"
	"os"

	"berichtsheft/internal/config"
	"berichtsheft/internal/keychain"
	"berichtsheft/internal/ui/shared"

	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
)

func main() {

	ctx := context.Background()

	var confClient config.EnvConfig
	conf, err := confClient.GetConfig()
	if err != nil {
		fmt.Println(ConsoleRed(fmt.Sprintf("failed to get config. %+v", err)))
		return
	}

	password, err := keychain.GetPassword(conf.Username)
	if err != nil {
		fmt.Println(ConsoleRed(fmt.Sprintf("failed to get password from the os keychain. %+v", err)))
		return
	}

	m, err := shared.NewProcessModel(ctx, conf, password)
	if err != nil {
		fmt.Println(ConsoleRed(fmt.Sprintf("failed to initialize process model. %+v", err)))
		return
	}

	p := tea.NewProgram(m)
	if _, err := p.Run(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func ConsoleRed(str string) string {
	style := lipgloss.NewStyle().Foreground(lipgloss.Red)
	return style.Render(str)
}
