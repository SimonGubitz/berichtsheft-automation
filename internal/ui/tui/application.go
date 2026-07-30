package tui

import (
	"context"
	"fmt"

	tea "charm.land/bubbletea/v2"

	"berichtsheft/internal/config"
	"berichtsheft/internal/keychain"
	"berichtsheft/internal/ui/shared"
)

type NavigateRoute int

const (
	HomeRoute NavigateRoute = iota
	ConfigRoute
	ProcessRoute
	ErrorRoute
)

type NavigateMsg struct {
	Route NavigateRoute
}

func NavigateTo(route NavigateRoute) tea.Cmd {
	return func() tea.Msg {
		return NavigateMsg{
			Route: route,
		}
	}
}

type ErrorMsg struct {
	Error error
}

func ThrowError(err error) tea.Cmd {
	return func() tea.Msg {
		return ErrorMsg{Error: err}
	}
}

type Application struct {
	ctx context.Context

	route  NavigateRoute
	routes []tea.Model

	conf     *config.Config
	password *string
}

func NewApplication(ctx context.Context, conf *config.Config) Application {
	// A missing config file is a valid first-run state. Keep one mutable config
	// object for every page so the config form can populate it before processing.
	if conf == nil {
		conf = &config.Config{}
	}

	routes := []tea.Model{
		NewHomePage(ctx, conf),
		NewConfigPage(ctx, conf),
		nil, // Created when Start is pressed, after reading the current password.
		nil,
	}

	return Application{
		ctx: ctx,

		route:  HomeRoute,
		routes: routes,

		conf:     conf,
		password: nil,
	}
}

func (a Application) Init() tea.Cmd {
	return nil
}

func (a Application) Update(msg tea.Msg) (tea.Model, tea.Cmd) {

	switch msg := msg.(type) {
	case ErrorMsg:
		a.routes[ErrorRoute] = NewErrorPage(msg.Error)
		a.route = ErrorRoute
		return a, nil

	case tea.KeyPressMsg:
		switch keypress := msg.String(); keypress {
		case "ctrl+c", "q":
			return a, tea.Quit
		}

	case NavigateMsg:
		switch msg.Route {
		case ProcessRoute:
			password, err := keychain.GetPassword(a.conf.Username)
			if err != nil {
				a.routes[ErrorRoute] = NewErrorPage(fmt.Errorf("could not load the password for %q: %w", a.conf.Username, err))
				a.route = ErrorRoute
				return a, nil
			}

			a.password = &password
			a.routes[ProcessRoute], err = shared.NewProcessModel(a.ctx, a.conf, password)
			if err != nil {
				a.routes[ErrorRoute] = NewErrorPage(fmt.Errorf("unknown route: %d", msg.Route))
				a.route = ErrorRoute
				return a, nil
			}

			a.route = ProcessRoute
			return a, a.routes[ProcessRoute].Init()
		default:
			if msg.Route < HomeRoute || msg.Route > ErrorRoute {
				a.routes[ErrorRoute] = NewErrorPage(fmt.Errorf("unknown route: %d", msg.Route))
				a.route = ErrorRoute
				return a, nil
			}
			a.route = msg.Route
		}
	}

	updated, cmd := a.routes[a.route].Update(msg)
	a.routes[a.route] = updated
	return a, cmd
}

func (a Application) View() tea.View {

	if a.route < HomeRoute || int(a.route) >= len(a.routes) {
		return tea.NewView("Invalid route")
	}

	if a.routes[a.route] == nil {
		return tea.NewView("This page is not available")
	}

	return a.routes[a.route].View()
}
