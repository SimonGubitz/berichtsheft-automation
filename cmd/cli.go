package cmd

import (
	"context"
	"fmt"
	"os"

	"berichtsheft/internal/config"
	"berichtsheft/internal/keychain"
	"berichtsheft/internal/ui/shared"

	tea "charm.land/bubbletea/v2"
)

var (
	cliCmd = &cobra.Command{
		Use: "cli",
	}
)

func init() {

}

func runCLI(ctx context.Context) {

	var confClient config.FileConfig
	conf, err := confClient.GetConfig()
	if err != nil {
		fmt.Println(ConsoleRed(fmt.Sprintf("failed to get config. %+v", err)))
		os.Exit(1)
	}

	password, err := keychain.GetPassword(conf.Username)
	if err != nil {
		fmt.Println(ConsoleRed(fmt.Sprintf("failed to get password from the os keychain. %+v", err)))
		os.Exit(1)
	}

	m, err := shared.NewProcessModel(ctx, conf, password)
	if err != nil {
		fmt.Println(ConsoleRed(fmt.Sprintf("failed to initialize process model. %+v", err)))
		os.Exit(1)
	}

	p := tea.NewProgram(m)
	if _, err := p.Run(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
