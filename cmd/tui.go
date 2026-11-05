package cmd

import (
	"context"
	"errors"
	"fmt"

	"berichtsheft/internal/config"
	"berichtsheft/internal/ui/tui"

	tea "charm.land/bubbletea/v2"
)

func runTui(ctx context.Context, ollamaHost string) error {

	var confClient config.FileConfig
	conf, err := confClient.GetConfig()
	if err != nil && !errors.Is(err, config.ErrNoConfigFile) {
		return fmt.Errorf("error starting program: %w", err)
	}

	m := tui.NewApplication(ctx, conf, ollamaHost)
	p := tea.NewProgram(m, tea.WithContext(ctx))
	if _, err := p.Run(); err != nil {
		return fmt.Errorf("error running program: %w", err)
	}

	return nil
}
