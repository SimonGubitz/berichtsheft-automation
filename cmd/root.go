package cmd

import (
	"context"
	"errors"
	"fmt"

	"berichtsheft/internal/config"
	"berichtsheft/internal/ui/tui"

	tea "charm.land/bubbletea/v2"
	"github.com/spf13/cobra"
)

var (
	rootCmd = &cobra.Command{
		Use: "berichtsheft",
	}
)

func runTUI(ctx context.Context) error {

	var confClient config.FileConfig
	conf, err := confClient.GetConfig()
	if err != nil && !errors.Is(err, config.ErrNoConfigFile) {
		return fmt.Errorf("Error starting program: " + err.Error())
	}

	m := tui.NewApplication(ctx, conf)
	p := tea.NewProgram(m, tea.WithContext(ctx))
	if _, err := p.Run(); err != nil {
		return fmt.Errorf("Error running program: " + err.Error())
	}

	return nil
}
