package ui

import (
	tea "charm.land/bubbletea/v2"
)

type NavigateRoute int

const (
	HomeRoute NavigateRoute = iota
	DateInputRoute
	ConfigRoute
	ProcessRoute
	ErrorRoute
)

type NavigateMsg struct {
	Route NavigateRoute
}

func NavigateToMsg(route NavigateRoute) tea.Cmd {
	return func() tea.Msg {
		return NavigateMsg{
			Route: route,
		}
	}
}
