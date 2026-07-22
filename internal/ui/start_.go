package ui

// import (
// 	"context"
// 	"fmt"
// 	"os"
// 	"strings"
//
// 	"berichtsheft/internal/config"
// 	"berichtsheft/internal/jira"
// 	"berichtsheft/internal/keychain"
// 	"berichtsheft/internal/ollama"
//
// 	"charm.land/bubbles/v2/spinner"
// 	tea "charm.land/bubbletea/v2"
// 	"charm.land/lipgloss/v2"
// )
//
// type StartPage struct {
// 	spinner spinner.Model
//
// 	conf       *config.Config
// 	activities *jira.ActivityFeed
// 	password   string
//
// 	loading bool
// 	status  string
//
// 	summary string
//
// 	quitting bool
// }
//
// func NewStartPage(conf *config.Config) StartPage {
//
// 	password, _ := keychain.GetPassword(conf.Username)
// 	// if err != nil {
// 	// 	return NewErrorPage(err), nil
// 	// }
//
// 	s := spinner.New()
// 	s.Spinner = spinner.Dot
// 	s.Style = lipgloss.NewStyle().Foreground(lipgloss.Color("205"))
//
// 	return StartPage{
// 		spinner:  s,
// 		loading:  true,
// 		status:   "loading...",
// 		conf:     conf,
// 		password: password,
// 	}
// }
//
// func (c StartPage) Init() tea.Cmd {
// 	return tea.Batch(
// 		fetchActivitiesCmd(c.conf),
// 		aiSummaryCmd(c.activities),
// 		c.spinner.Tick,
// 	)
// }
//
// func (c StartPage) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
// 	switch msg := msg.(type) {
// 	case tea.KeyPressMsg:
// 		switch keypress := msg.String(); keypress {
// 		case "q", "ctrl+c":
// 			return c, tea.Quit
//
// 		case "-":
// 			return NewHomePage(c.conf), nil
// 		default:
// 			return c, nil
// 		}
//
// 	case activitiesLoadedMsg:
// 		if msg.err != nil {
// 			return NewErrorPage(msg.err), nil
// 		} else {
// 			c.activities = msg.activities
// 			c.status = "Activities loaded"
// 		}
//
// 	case aiSummaryLoadedMsg:
// 		if msg.err != nil {
// 			return NewErrorPage(msg.err), nil
// 		} else {
// 			c.summary = msg.summary
// 			c.status = "Summary ready"
// 		}
//
// 	case spinner.TickMsg:
// 		var cmd tea.Cmd
// 		c.spinner, cmd = c.spinner.Update(msg)
// 		return c, cmd
//
// 	default:
// 		var cmd tea.Cmd
// 		c.spinner, cmd = c.spinner.Update(msg)
// 		return c, cmd
// 	}
//
// 	return c, nil
// }
//
// func (c StartPage) View() tea.View {
//
// 	var b strings.Builder
//
// 	b.WriteString(strings.Repeat(" ", 50) + "|\n")
//
// 	if c.loading {
// 		b.WriteString(lipgloss.JoinHorizontal(
// 			lipgloss.Center,
// 			c.spinner.View(),
// 			c.status,
// 		))
// 	}
//
// 	return tea.NewView(b.String())
// }
