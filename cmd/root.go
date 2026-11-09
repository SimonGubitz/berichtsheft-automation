package cmd

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"time"

	"charm.land/lipgloss/v2"
	"github.com/adrg/xdg"
	"github.com/spf13/cobra"
)

var (
	log     *slog.Logger
	logFile *os.File
	rootCmd = &cobra.Command{
		Use: "berichtsheft",
		PersistentPreRunE: func(cmd *cobra.Command, args []string) error {

			// setup logger
			timeStr := time.Now().Format(time.RFC3339)
			logFilePath, err := xdg.CacheFile(filepath.Join("berichtsheft", "logs", timeStr+".log"))
			if err != nil {
				return err
			}

			logFile, err = os.OpenFile(logFilePath, os.O_CREATE|os.O_RDWR, 0o600)
			if err != nil {
				return err
			}
			defer func() {}()

			h := slog.NewJSONHandler(logFile, &slog.HandlerOptions{Level: slog.LevelDebug})
			log = slog.New(h)

			slog.SetDefault(log)

			return nil
		},
		RunE: func(cmd *cobra.Command, args []string) error {

			ollamaFlag, err := cmd.Flags().GetString("ollama")
			if err != nil {
				return err
			}

			return runTui(cmd.Context(), ollamaFlag)
		},
		PersistentPostRunE: func(cmd *cobra.Command, args []string) error {
			return logFile.Close()
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
