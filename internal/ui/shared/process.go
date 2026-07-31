package shared

import (
	"context"
	"errors"
	"fmt"
	"os"
	"strings"

	"berichtsheft/internal/config"
	"berichtsheft/internal/jira"

	"charm.land/bubbles/v2/spinner"
	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
)

var (
	textStyle     = lipgloss.NewStyle().Foreground(lipgloss.Color("201"))
	successStyle  = lipgloss.NewStyle().Foreground(lipgloss.Color("42")).Bold(true)
	responseStyle = lipgloss.NewStyle() //.Foreground(lipgloss.Color("#999999"))
)

type HashSet map[string]struct{}

type model struct {
	ctx context.Context

	ollamaHost string
	jiraClient *jira.JiraClient
	activities *jira.ActivityFeed

	response       string
	state          State
	jiraActCached  int
	jiraWeekCached int

	spinner  spinner.Model
	quitting bool
	err      error
}

func NewProcessModel(ctx context.Context, conf *config.Config, password string) (model, error) {

	ollamaHost := os.Getenv("OLLAMA_HOST")
	if ollamaHost == "" {
		return model{}, errors.New("environment variable `OLLAMA_HOST` not set")
	}

	j, err := jira.NewJiraClient(conf.JiraHost, conf.Username, password)
	if err != nil {
		return model{}, fmt.Errorf("failed to create jira client. %w", err)
	}

	s := spinner.New()
	s.Spinner = spinner.Dot
	s.Style = lipgloss.NewStyle().Foreground(lipgloss.Color("205"))
	return model{
		ctx:        ctx,
		ollamaHost: ollamaHost,
		jiraClient: j,

		spinner: s,

		state: StateLoadingJiraAct,
	}, nil
}

// RunProcess executes the same process model without rendering a terminal UI.
// It returns the generated report once the model receives a response.
func RunProcess(ctx context.Context, conf *config.Config, password string) (string, error) {
	m, err := NewProcessModel(ctx, conf, password)
	if err != nil {
		return "", err
	}

	p := tea.NewProgram(
		headlessProcessModel{process: m},
		tea.WithContext(ctx),
		tea.WithInput(nil),
		tea.WithoutRenderer(),
	)
	finalModel, err := p.Run()
	if err != nil {
		return "", err
	}

	final, ok := finalModel.(headlessProcessModel)
	if !ok {
		return "", fmt.Errorf("unexpected headless process model: %T", finalModel)
	}
	if final.err != nil {
		return "", final.err
	}

	return final.response, nil
}

// headlessProcessModel delegates work to model while terminating as soon as a
// response or error is received. The interactive model stays open on errors so
// people can read them; headless mode must return control to its caller.
type headlessProcessModel struct {
	process  tea.Model
	response string
	err      error
}

func (m headlessProcessModel) Init() tea.Cmd {
	return m.process.Init()
}

func (m headlessProcessModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case ErrMsg:
		m.err = msg
		return m, tea.Quit
	case OllamaRespMsg:
		m.response = msg.Response
		return m, tea.Quit
	default:
		updated, cmd := m.process.Update(msg)
		m.process = updated
		return m, cmd
	}
}

func (m headlessProcessModel) View() tea.View {
	return tea.NewView("")
}

func (m model) Init() tea.Cmd {
	return tea.Batch(
		m.spinner.Tick,
		GetJiraActivityCmd(m.ctx, m.jiraClient),
	)
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyPressMsg:
		switch msg.String() {
		case "q", "esc", "ctrl+c":
			m.quitting = true
			return m, tea.Quit
		default:
			return m, nil
		}

	case ErrMsg:
		m.err = msg
		return m, nil

	case JiraActivityMsg:
		m.state = StateLoadingJiraWeek
		m.activities = msg.Activities
		m.jiraActCached = msg.Cached
		return m, tea.Batch(
			m.spinner.Tick,
			GetJiraWeekCmd(m.ctx, m.jiraClient, m.activities),
		)

	case JiraWeeksMsg:
		m.state = StateLoadingLLM
		m.jiraWeekCached = msg.Cached
		return m, tea.Batch(
			m.spinner.Tick,
			GetLLMSummaryCmd(m.ctx, m.ollamaHost, msg.Input),
		)

	case OllamaRespMsg:
		m.state = StateFinished
		m.response = msg.Response
		return m, tea.Quit

	default:
		var cmd tea.Cmd
		m.spinner, cmd = m.spinner.Update(msg)
		return m, cmd
	}
}

func (m model) View() tea.View {
	if m.err != nil {
		return tea.NewView(m.err.Error())
	}
	if m.quitting {
		return tea.NewView("\n\n   Quitting...\n")
	}

	var b strings.Builder
	b.WriteString("\n\n")

	check := "✅"

	switch m.state {
	case StateLoadingJiraAct:
		fmt.Fprintf(&b, "   %s %s\n", m.spinner.View(), textStyle.Render("Fetching your Activity..."))
	default:
		fmt.Fprintf(&b, "   %s %s %s\n", check, successStyle.Render("Fetched your Activity"), responseStyle.Render(fmt.Sprintf("(%d%% cached)", m.jiraActCached)))
	}

	switch m.state {
	case StateLoadingJiraWeek:
		fmt.Fprintf(&b, "   %s %s\n", m.spinner.View(), textStyle.Render("Building your weeks..."))
	case StateLoadingLLM, StateFinished:
		fmt.Fprintf(&b, "   %s %s %s\n", check, successStyle.Render("Built your weeks"), responseStyle.Render(fmt.Sprintf("(%d%% cached)", m.jiraWeekCached)))
	}

	switch m.state {
	case StateLoadingLLM:
		fmt.Fprintf(&b, "   %s %s\n", m.spinner.View(), textStyle.Render("Thinking..."))
	case StateFinished:
		fmt.Fprintf(&b, "   %s %s\n", check, successStyle.Render("Finished"))
	}

	if m.state == StateFinished {
		fmt.Fprintf(&b, "\n%s\n\n", responseStyle.Render(m.response))
	} else {
		b.WriteString("\n   press q to quit\n\n")
	}

	return tea.NewView(b.String())
}
