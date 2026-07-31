package tui

import (
	"fmt"

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
	str := fmt.Sprintf("There was an error.\n%s\n", errorStyle.Render(e.errorMsg))
	return tea.NewView(str)
}
