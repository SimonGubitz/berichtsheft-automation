package cmd

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"berichtsheft/internal/config"
	"berichtsheft/internal/keychain"
	"berichtsheft/internal/ui/shared"

	"github.com/spf13/cobra"
)

var (
	headlessCmd = &cobra.Command{
		Use:   "headless",
		Short: "Run without any graphics.",
		Long:  "",
		PreRunE: func(cmd *cobra.Command, args []string) error {

			outputFlag, err := cmd.Flags().GetString("output")
			if err != nil {
				return err
			}

			// check for a valid path
			if outputFlag != "" && outputFlag != "-" {
				dir, _ := filepath.Split(outputFlag)

				_, err := os.Stat(dir)
				if err != nil {
					return err
				}
			}

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

			envConfigFlag, err := cmd.Flags().GetBool("env-config")
			if err != nil {
				return err
			}

			fileConfigFlag, err := cmd.Flags().GetBool("file-config")
			if err != nil {
				return err
			}

			usernameFlag, err := cmd.Flags().GetString("username")
			if err != nil {
				return err
			}

			jiraHostFlag, err := cmd.Flags().GetString("jira-host")
			if err != nil {
				return err
			}

			ollamaHostFlag, err := cmd.Flags().GetString("ollama")
			if err != nil {
				return err
			}

			outputFlag, err := cmd.Flags().GetString("output")
			if err != nil {
				return err
			}

			return runHeadless(cmd.Context(), fromFlag, toFlag, fileConfigFlag, envConfigFlag, usernameFlag, jiraHostFlag, ollamaHostFlag, outputFlag)
		},
	}
)

func init() {

	headlessCmd.Flags().BoolP("env-config", "", false, "Uses the JIRA_URL, JIRA_USERNAME environment variables to configure the run.")
	headlessCmd.Flags().BoolP("file-config", "", false, "Uses the config file that has been created with the TUI run.")
	headlessCmd.MarkFlagsMutuallyExclusive("env-config", "file-config")

	headlessCmd.Flags().StringP("username", "u", "", "Sets the Jira Username. Will be ignored if either '--file-config' or '--env-config' is set")
	headlessCmd.Flags().StringP("jira-host", "", "", "Sets the Jira-Url. Will be ignored if either '--file-config' or '--env-config' is set")
	headlessCmd.Flags().StringP("output", "o", "-", "Write the generated report to this file instead of standard output.")

	formats := []string{time.DateOnly, "02.01.2006", "01/02/2006"}
	headlessCmd.Flags().TimeP("from", "f", time.Time{}, formats, "Starting date for the activities.")
	headlessCmd.Flags().TimeP("to", "t", time.Time{}, formats, "End date for the activities.")
	_ = headlessCmd.MarkFlagRequired("from")
	_ = headlessCmd.MarkFlagRequired("to")

	headlessCmd.Flags().StringP("ollama", "", "http://localhost:11434", "Base URL of the Ollama instance.")

	rootCmd.AddCommand(headlessCmd)
}

func runHeadless(ctx context.Context, from, to time.Time, useFileConfig, useEnvConfig bool, configUsername, configJiraHost, ollamaHost, output string) error {

	if useFileConfig && useEnvConfig {
		return fmt.Errorf("cannot use '--file-config' and '--env-config' simultaneously")
	}

	var confClient config.ConfigClient
	switch {
	case useFileConfig:
		confClient = config.NewFileConfig()
	case useEnvConfig:
		confClient = config.NewEnvConfig()
	default:
		confClient = config.NewVarConfig(configUsername, configJiraHost)
	}

	conf, err := confClient.GetConfig()
	if err != nil {
		return fmt.Errorf("failed to get configuration. %w", err)
	}

	password, err := keychain.GetPassword(conf.Username)
	if err != nil {
		return fmt.Errorf("failed to get password from the os keychain. %w", err)
	}

	report, err := shared.RunProcess(ctx, conf, from, to, ollamaHost, password)
	if err != nil {
		return fmt.Errorf("failed to generate report. %w", err)
	}

	if output == "-" {
		_, _ = fmt.Fprintln(os.Stdout, report)
		return nil
	}

	if err := os.WriteFile(output, []byte(report+"\n"), 0o600); err != nil {
		return fmt.Errorf("failed to write report to %s. %w", output, err)
	}

	return nil
}
