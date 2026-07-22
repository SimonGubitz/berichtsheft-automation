package ui

import (
	"context"
	"fmt"
	"log"
	"math/rand"
	"os"
	"time"

	"berichtsheft/internal/config"
	"berichtsheft/internal/jira"
	"berichtsheft/internal/keychain"
	"berichtsheft/internal/ollama"

	"charm.land/bubbles/v2/spinner"
	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
)

var (
	helpingStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("241")).Render
	mainStyle    = lipgloss.NewStyle().MarginLeft(1)
)

// func main() {
// 	var (
// 		daemonMode bool
// 		showHelp   bool
// 	)
//
// 	flag.BoolVar(&daemonMode, "d", false, "run as a daemon")
// 	flag.BoolVar(&showHelp, "h", false, "show help")
// 	flag.Parse()
//
// 	if showHelp {
// 		flag.Usage()
// 		os.Exit(0)
// 	}
//
// 	opts := []tea.ProgramOption{}
// 	if daemonMode || !term.IsTerminal(os.Stdout.Fd()) {
// 		// If we're in daemon mode don't render the TUI
// 		opts = append(opts, tea.WithoutRenderer())
// 	} else {
// 		// If we're in TUI mode, discard log output
// 		log.SetOutput(io.Discard)
// 	}
//
// 	p := tea.NewProgram(NewStartPage(), opts...)
// 	if _, err := p.Run(); err != nil {
// 		fmt.Println("Error starting Bubble Tea program:", err)
// 		os.Exit(1)
// 	}
// }

type result struct {
	duration time.Duration
	emoji    string
}

type StartPage struct {
	spinner spinner.Model
	results []result

	conf       *config.Config
	activities *jira.ActivityFeed
	password   string

	quitting bool
}

func NewStartPage(conf *config.Config) StartPage {
	const showLastResults = 2

	sp := spinner.New()
	sp.Style = lipgloss.NewStyle().Foreground(lipgloss.Color("206"))

	return StartPage{
		spinner: sp,
		results: make([]result, showLastResults),

		conf: conf,
	}
}

func (m StartPage) Init() tea.Cmd {
	log.Println("Starting work...")
	return tea.Batch(
		m.spinner.Tick,
		fetchActivitiesCmd(m.conf),
	)
}

func (m StartPage) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyPressMsg:
		m.quitting = true
		return m, tea.Quit

	case spinner.TickMsg:
		var cmd tea.Cmd
		m.spinner, cmd = m.spinner.Update(msg)
		return m, cmd

	case processFinishedMsg:
		d := time.Duration(msg)
		res := result{emoji: randomEmoji(), duration: d}
		log.Printf("%s Job finished in %s", res.emoji, res.duration)
		m.results = append(m.results[1:], res)
		return m, runPretendProcess

	default:
		return m, nil
	}
}

func (m StartPage) View() tea.View {
	s := "\n" +
		m.spinner.View() + " Doing some work...\n\n"

	for _, res := range m.results {
		if res.duration == 0 {
			s += "........................\n"
		} else {
			s += fmt.Sprintf("%s Job finished in %s\n", res.emoji, res.duration)
		}
	}

	s += helpingStyle("\nPress any key to exit\n")

	if m.quitting {
		s += "\n"
	}

	return tea.NewView(mainStyle.Render(s))
}

// processFinishedMsg is sent when a pretend process completes.
type processFinishedMsg time.Duration

// pretendProcess simulates a long-running process.
func runPretendProcess() tea.Msg {
	pause := time.Duration(rand.Int63n(899)+100) * time.Millisecond // nolint:gosec
	time.Sleep(pause)
	return processFinishedMsg(pause)
}

func randomEmoji() string {
	emojis := []rune("🍦🧋🍡🤠👾😭🦊🐯🦆🥨🎏🍔🍒🍥🎮📦🦁🐶🐸🍕🥐🧲🚒🥇🏆🌽")
	return string(emojis[rand.Intn(len(emojis))]) // nolint:gosec
}

type activitiesLoadedMsg struct {
	activities *jira.ActivityFeed
	err        error
}

type aiSummaryLoadedMsg struct {
	summary string
	err     error
}

func fetchActivitiesCmd(conf *config.Config) tea.Cmd {
	return func() tea.Msg {
		ctx := context.Background()
		activities, err := fetchActivities(ctx, conf)
		return activitiesLoadedMsg{
			activities: activities,
			err:        err,
		}
	}
}

func fetchActivities(ctx context.Context, conf *config.Config) (*jira.ActivityFeed, error) {

	password, err := keychain.GetPassword(conf.Username)
	if err != nil {
		return nil, err
	}

	jira, err := jira.NewJiraClient("https://jira.optitime.de", conf.Username, password)
	if err != nil {
		return nil, err
	}

	activities, err := jira.GetActivities(ctx, 150)
	if err != nil {
		return nil, err
	}

	return activities, nil
}

func aiSummaryCmd(activities *jira.ActivityFeed) tea.Cmd {
	if activities == nil {
		return nil
	}

	return func() tea.Msg {
		ctx := context.Background()
		summary, err := getAISummary(ctx)
		return aiSummaryLoadedMsg{
			summary: summary,
			err:     err,
		}
	}
}

func getAISummary(ctx context.Context) (string, error) {

	host := os.Getenv("OLLAMA_HOST")
	o := ollama.NewOllamaClient(host)
	models, err := o.ListModels(ctx)
	if err != nil {
		return "", fmt.Errorf("failed to list models. %w", err)
	}

	if len(models) == 0 {
		err := o.PullModel(ctx)
		if err != nil {
			return "", fmt.Errorf("failed to pull model. %w", err)
		}
	}

	s, err := o.GetResponse(ctx, "Why is the sky blue?")
	if err != nil {
		return "", fmt.Errorf("received error on response. %w", err)
	}

	return s, nil
}
