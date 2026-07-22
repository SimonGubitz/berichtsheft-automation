package ui

import (
	tea "charm.land/bubbletea/v2"
)

type Route struct {
	Key   string
	Value tea.Model
}

func NewRoute(key string, value tea.Model) Route {
	return Route{key, value}
}

type Application struct {
	route Route
}

func NewApplication(baseRoute Route) Application {
	return Application{
		route: baseRoute,
	}
}

func (a Application) Init() tea.Cmd {
	return nil
}

func (a Application) Update(_ tea.Msg) (tea.Model, tea.Cmd) {
	return a.route.Value, nil
}

func (a Application) View() tea.View {
	return a.route.Value.View()
}
