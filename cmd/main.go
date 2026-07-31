package main

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"os"

	"berichtsheft/internal/config"
	"berichtsheft/internal/keychain"
	"berichtsheft/internal/ui/shared"
	"berichtsheft/internal/ui/tui"

	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
)

var (
	version = "development"

	warningStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("#ebae34"))
)

func main() {

	var showVersion bool
	flag.BoolVar(&showVersion, "v", false, "Print the version and exit.")
	flag.BoolVar(&showVersion, "version", false, "Print the version and exit.")

	// no config
	cli := flag.Bool("cli", false, "Run using the saved configuration without interactive input.")

	// Headless
	headless := flag.Bool("headless", false, "Run automatically with OLLAMA_HOST needs an accompanying configuration option. Writes to standard output unless -o is set.")
	useEnvConfig := flag.Bool("env-config", false, "Uses the JIRA_URL, JIRA_USERNAME environment variables to configure the run.")
	useFileConfig := flag.Bool("file-config", false, "Uses the config file that has been created with the TUI run.")
	configUsername := flag.String("config-username", "", "Sets the Jira Username. Will be ignored if either '--file-config' or '--env-config' is set")
	configJiraHost := flag.String("config-jira-url", "", "Sets the Jira-Url. Will be ignored if either '--file-config' or '--env-config' is set")
	output := flag.String("o", "-", "Write the generated report to this file instead of standard output.")

	// parse
	flag.Parse()

	ctx := context.Background()

	switch {
	case showVersion:
		printVersion()
	case *headless:
		runHeadless(ctx, *useFileConfig, *useEnvConfig, *configUsername, *configJiraHost, *output)
	case *cli:
		ShowNonHeadlessWarnings(*useEnvConfig, *useFileConfig, *configUsername, *configJiraHost)
		runCLI(ctx)
	default:
		ShowNonHeadlessWarnings(*useEnvConfig, *useFileConfig, *configUsername, *configJiraHost)
		runTUI(ctx)
	}
}

func printVersion() {
	fmt.Printf("berichtsheft-automation version: %s\n", version)
}

func ConsoleRed(str string) string {
	style := lipgloss.NewStyle().Foreground(lipgloss.Red)
	return style.Render(str)
}

func ShowNonHeadlessWarnings(useEnvConfig, useFileConfig bool, configUsername, configJiraHost string) {
	if useEnvConfig {
		fmt.Println(warningStyle.Render("⚠️ '--env-config' will be ignored in this mode"))
	}
	if useFileConfig {
		fmt.Println(warningStyle.Render("⚠️ '--file-config' will be ignored in this mode"))
	}
	if configUsername != "" {
		fmt.Println(warningStyle.Render("⚠️ '--config-username' will be ignored in this mode"))
	}
	if configJiraHost != "" {
		fmt.Println(warningStyle.Render("⚠️ '--config-jira-url' will be ignored in this mode"))
	}
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

func runHeadless(ctx context.Context, useFileConfig, useEnvConfig bool, configUsername, configJiraHost, output string) {

	if useFileConfig && useEnvConfig {
		fmt.Println(ConsoleRed("cannot use '--file-config' and '--env-config' simultaneously"))
		os.Exit(1)
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
		fmt.Println(ConsoleRed(fmt.Sprintf("failed to get configuration. %+v", err)))
		os.Exit(1)
	}

	password, err := keychain.GetPassword(conf.Username)
	if err != nil {
		fmt.Println(ConsoleRed(fmt.Sprintf("failed to get password from the os keychain. %+v", err)))
		os.Exit(1)
	}

	report, err := shared.RunProcess(ctx, conf, password)
	if err != nil {
		fmt.Println(ConsoleRed(fmt.Sprintf("failed to generate report. %+v", err)))
		os.Exit(1)
	}

	if output == "-" {
		_, _ = fmt.Fprintln(os.Stdout, report)
		return
	}

	if err := os.WriteFile(output, []byte(report+"\n"), 0o600); err != nil {
		fmt.Println(ConsoleRed(fmt.Sprintf("failed to write report to %s. %+v", output, err)))
	}
}

func runTUI(ctx context.Context) {

	var confClient config.FileConfig
	conf, err := confClient.GetConfig()
	if err != nil && !errors.Is(err, config.ErrNoConfigFile) {
		fmt.Println(ConsoleRed("Error starting program: " + err.Error()))
		os.Exit(1)
	}

	m := tui.NewApplication(ctx, conf)
	p := tea.NewProgram(m, tea.WithContext(ctx))
	if _, err := p.Run(); err != nil {
		fmt.Println(ConsoleRed("Error running program: " + err.Error()))
		os.Exit(1)
	}
}
