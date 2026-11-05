package tui

import (
	"fmt"
	"strings"

	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
)

var (
	errorStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("#EB4268"))
)

type ErrorPage struct {
	errorMsg string
}

func NewErrorPage(err error) ErrorPage {
	return ErrorPage{
		errorMsg: err.Error(),
	}
}

func (e ErrorPage) Init() tea.Cmd {
	return nil
}

func (e ErrorPage) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	return e, tea.Quit
}

func (e ErrorPage) View() tea.View {
	lPad := strings.Repeat(" ", 5)

	var sb strings.Builder
	fmt.Fprintf(&sb, "\n\n")
	fmt.Fprintf(&sb, "%s %s\n", lPad, errorStyle.Render("Unfortunately there has been an error."))
	fmt.Fprintf(&sb, "%s %s\n", lPad, e.errorMsg)
	fmt.Fprintf(&sb, "\n\n")

	return tea.NewView(sb.String())
}
