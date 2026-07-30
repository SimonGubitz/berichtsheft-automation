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
	switch msg := msg.(type) {
	case tea.KeyPressMsg:
		switch keypress := msg.String(); keypress {
		case "q", "ctrl+c":
			return e, tea.Quit
		}
	}

	return e, nil
}

func (e ErrorPage) View() tea.View {
	return tea.NewView("errorrrrr\n\n" + e.errorMsg)
}
