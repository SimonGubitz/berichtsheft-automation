package cmd

import (
	"context"
	"fmt"
	"os"

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

			envConfigFlag, err := cmd.Flags().GetBool("env-config")
			if err != nil {
				return err
			}

			fileConfigFlag, err := cmd.Flags().GetBool("file-config")
			if err != nil {
				return err
			}

			outputFlag, err := cmd.Flags().GetString("output")
			if err != nil {
				return err
			}

			if fileConfigFlag && envConfigFlag {
				return fmt.Errorf("Flags `%s` and `%s` are mutually exclusive.", "env-config", "file-config")
			}

			if outputFlag != "" || outputFlag != "-" {
				// check for a valid path
			}

			return nil
		},
		RunE: func(cmd *cobra.Command, args []string) error {

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

			outputFlag, err := cmd.Flags().GetString("output")
			if err != nil {
				return err
			}

			return runHeadless(cmd.Context(), fileConfigFlag, envConfigFlag, usernameFlag, jiraHostFlag, outputFlag)
		},
	}
)

func init() {

	headlessCmd.Flags().BoolP("env-config", "", false, "Uses the JIRA_URL, JIRA_USERNAME environment variables to configure the run.")
	headlessCmd.Flags().BoolP("file-config", "", false, "Uses the config file that has been created with the TUI run.")
	headlessCmd.Flags().StringP("username", "u", "", "Sets the Jira Username. Will be ignored if either '--file-config' or '--env-config' is set")
	headlessCmd.Flags().StringP("jira-host", "", "", "Sets the Jira-Url. Will be ignored if either '--file-config' or '--env-config' is set")
	headlessCmd.Flags().StringP("output", "o", "-", "Write the generated report to this file instead of standard output.")

	rootCmd.AddCommand(headlessCmd)
}

func runHeadless(ctx context.Context, useFileConfig, useEnvConfig bool, configUsername, configJiraHost, output string) error {

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
		return fmt.Errorf(fmt.Sprintf("failed to get configuration. %+v", err))
	}

	password, err := keychain.GetPassword(conf.Username)
	if err != nil {
		return fmt.Errorf(fmt.Sprintf("failed to get password from the os keychain. %+v", err))
	}

	report, err := shared.RunProcess(ctx, conf, password)
	if err != nil {
		return fmt.Errorf(fmt.Sprintf("failed to generate report. %+v", err))
	}

	if output == "-" {
		_, _ = fmt.Fprintln(os.Stdout, report)
		return nil
	}

	if err := os.WriteFile(output, []byte(report+"\n"), 0o600); err != nil {
		return fmt.Errorf(fmt.Sprintf("failed to write report to %s. %+v", output, err))
	}

	return nil
}
