package cmd

import (
	"context"
	"fmt"
	"os"

	"charm.land/lipgloss/v2"
	"github.com/spf13/cobra"
)

var (
	rootCmd = &cobra.Command{
		Use: "berichtsheft",
		RunE: func(cmd *cobra.Command, args []string) error {

			ollamaFlag, err := cmd.Flags().GetString("ollama")
			if err != nil {
				return err
			}

			return runTui(cmd.Context(), ollamaFlag)
		},
	}
)

func Execute(ctx context.Context, version string) {
	rootCmd.SetContext(ctx)
	rootCmd.Version = version

	rootCmd.Flags().StringP("ollama", "", "http://localhost:11434", "Base URL of the Ollama instance.")

	if err := rootCmd.Execute(); err != nil {
		style := lipgloss.NewStyle().Foreground(lipgloss.Red)
		fmt.Fprintf(os.Stderr, style.Render("%+v"), err)
	}
}
