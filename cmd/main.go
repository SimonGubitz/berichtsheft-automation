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

func main() {

	// no config
	noConfig := flag.Bool("no-config", false, "Runs the script in the terminal without configuration. Can only be used after a configuration has already been generated.")

	// Headless
	headless := flag.Bool("headless", false, "Run automatically with JIRA_URL, JIRA_USERNAME, and OLLAMA_HOST. Writes to standard output unless -o is set.")
	output := flag.String("o", "-", "Write the generated report to this file instead of standard output.")

	// parse
	flag.Parse()

	ctx := context.Background()

	switch {
	case *headless:
		runHeadless(ctx, *output)
	case *noConfig:
		runNoConfig(ctx)
	default:
		runTUI(ctx)
	}

}

func ConsoleRed(str string) string {
	style := lipgloss.NewStyle().Foreground(lipgloss.Red)
	return style.Render(str)
}

func runNoConfig(ctx context.Context) {

	var confClient config.FileConfig
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

func runHeadless(ctx context.Context, output string) {
	var confClient config.EnvConfig
	conf, err := confClient.GetConfig()
	if err != nil {
		fmt.Println(ConsoleRed(fmt.Sprintf("failed to get configuration. %+v", err)))
		return
	}

	password, err := keychain.GetPassword(conf.Username)
	if err != nil {
		fmt.Println(ConsoleRed(fmt.Sprintf("failed to get password from the os keychain. %+v", err)))
		return
	}

	report, err := shared.RunProcess(ctx, conf, password)
	if err != nil {
		fmt.Println(ConsoleRed(fmt.Sprintf("failed to generate report. %+v", err)))
		return
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
