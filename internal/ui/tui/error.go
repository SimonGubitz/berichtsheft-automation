package tui

import (
	tea "charm.land/bubbletea/v2"
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
	if keypress, ok := msg.(tea.KeyPressMsg); ok {
		switch keypress.String() {
		case "q", "ctrl+c":
			return e, tea.Quit
		}
	}

	return e, nil
}

func (e ErrorPage) View() tea.View {
	return tea.NewView("errorrrrr\n\n" + e.errorMsg)
}
