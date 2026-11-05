package tui

import (
	"context"
	"fmt"
	"time"

	"berichtsheft/internal/config"
	"berichtsheft/internal/keychain"
	"berichtsheft/internal/ui"
	"berichtsheft/internal/ui/shared"

	tea "charm.land/bubbletea/v2"
)

type DateInputMsg struct {
	From time.Time
	To   time.Time
}

func SendTimeMsg(from, to time.Time) tea.Cmd {
	return func() tea.Msg {
		return DateInputMsg{From: from, To: to}
	}
}

type ErrorMsg struct {
	Error error
}

func ThrowErrorMsg(err error) tea.Cmd {
	return func() tea.Msg {
		return ErrorMsg{Error: err}
	}
}

type Application struct {
	ctx context.Context

	route  ui.NavigateRoute
	routes map[ui.NavigateRoute]tea.Model

	from time.Time
	to   time.Time

	conf       *config.Config
	password   *string
	ollamaHost string
}

func NewApplication(ctx context.Context, conf *config.Config, ollamaHost string) Application {
	// A missing config file is a valid first-run state. Keep one mutable config
	// object for every page so the config form can populate it before processing.
	if conf == nil {
		conf = &config.Config{}
	}

	routes := make(map[ui.NavigateRoute]tea.Model)
	routes[ui.HomeRoute] = NewHomePage(ctx, conf)
	routes[ui.DateInputRoute] = NewDateInputPage()
	routes[ui.ConfigRoute] = NewConfigPage(ctx, conf)

	return Application{
		ctx: ctx,

		route:  ui.HomeRoute,
		routes: routes,

		conf:       conf,
		password:   nil,
		ollamaHost: ollamaHost,
	}
}

func (a Application) Init() tea.Cmd {
	return nil
}

func (a Application) Update(msg tea.Msg) (tea.Model, tea.Cmd) {

	switch msg := msg.(type) {
	case ErrorMsg:
		a.routes[ui.ErrorRoute] = NewErrorPage(msg.Error)
		a.route = ui.ErrorRoute
		return a, nil

	case tea.KeyPressMsg:
		switch keypress := msg.String(); keypress {
		case "ctrl+c", "q":
			return a, tea.Quit
		}

	case ui.NavigateMsg:
		switch msg.Route {
		case ui.ProcessRoute:
			password, err := keychain.GetPassword(a.conf.Username)
			if err != nil {
				a.routes[ui.ErrorRoute] = NewErrorPage(fmt.Errorf("could not load the password for %q: %w", a.conf.Username, err))
				a.route = ui.ErrorRoute
				return a, nil
			}

			a.password = &password
			a.routes[ui.ProcessRoute], err = shared.NewProcessModel(a.ctx, a.conf, a.from, a.to, a.ollamaHost, password)
			if err != nil {
				a.routes[ui.ErrorRoute] = NewErrorPage(fmt.Errorf("%w", err))
				a.route = ui.ErrorRoute
				return a, nil
			}

			a.route = ui.ProcessRoute
			return a, a.routes[ui.ProcessRoute].Init()
		default:
			if msg.Route < ui.HomeRoute || msg.Route > ui.ErrorRoute {
				a.routes[ui.ErrorRoute] = NewErrorPage(fmt.Errorf("unknown route: %d", msg.Route))
				a.route = ui.ErrorRoute
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

	if a.routes[a.route] == nil {
		return tea.NewView("This page is not available")
	}

	return a.routes[a.route].View()
}
