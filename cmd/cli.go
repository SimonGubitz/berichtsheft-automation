package cmd

import (
	"context"
	"fmt"
	"time"

	"berichtsheft/internal/config"
	"berichtsheft/internal/keychain"
	"berichtsheft/internal/ui/shared"

	tea "charm.land/bubbletea/v2"
	"github.com/spf13/cobra"
)

var (
	cliCmd = &cobra.Command{
		Use:   "cli",
		Short: "Runs the preconfigured state without setup.",
		PreRunE: func(cmd *cobra.Command, args []string) error {
			return checkDateFlags(cmd)
		},
		RunE: func(cmd *cobra.Command, args []string) error {

			fromFlag, err := cmd.Flags().GetTime("from")
			if err != nil {
				return err
			}

			toFlag, err := cmd.Flags().GetTime("to")
			if err != nil {
				return err
			}

			ollamaFlag, err := cmd.Flags().GetString("ollama")
			if err != nil {
				return err
			}

			return runCli(cmd.Context(), fromFlag, toFlag, ollamaFlag)
		},
	}
)

func init() {

	formats := []string{time.DateOnly, "02.01.2006", "01/02/2006"}
	cliCmd.Flags().TimeP("from", "f", time.Time{}, formats, "Starting date for the activities.")
	cliCmd.Flags().TimeP("to", "t", time.Time{}, formats, "End date for the activities.")
	_ = cliCmd.MarkFlagRequired("from")
	_ = cliCmd.MarkFlagRequired("to")

	cliCmd.Flags().StringP("ollama", "", "http://localhost:11434", "Base URL of the Ollama instance.")

	rootCmd.AddCommand(cliCmd)
}

func runCli(ctx context.Context, from, to time.Time, ollamaHost string) error {

	var confClient config.FileConfig
	conf, err := confClient.GetConfig()
	if err != nil {
		return fmt.Errorf("failed to get config. %w", err)
	}

	password, err := keychain.GetPassword(conf.Username)
	if err != nil {
		return fmt.Errorf("failed to get password from the os keychain. %w", err)
	}

	m, err := shared.NewProcessModel(ctx, conf, from, to, ollamaHost, password)
	if err != nil {
		return fmt.Errorf("failed to initialize process model. %w", err)
	}

	p := tea.NewProgram(m)
	_, err = p.Run()
	return err
}
