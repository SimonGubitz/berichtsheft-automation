package tui

import (
	"context"
	"fmt"
	"strings"

	"berichtsheft/internal/config"
	"berichtsheft/internal/keychain"
	"berichtsheft/internal/ui"

	"charm.land/bubbles/v2/cursor"
	"charm.land/bubbles/v2/textinput"
	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
)

var (
	focusedStyle        = lipgloss.NewStyle().Foreground(lipgloss.Color("205"))
	blurredStyle        = lipgloss.NewStyle().Foreground(lipgloss.Color("240"))
	helpStyle           = blurredStyle
	cursorModeHelpStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("244"))

	focusedButton = focusedStyle.Render("[ Submit ]")
	blurredButton = fmt.Sprintf("[ %s ]", blurredStyle.Render("Submit"))
)

const (
	urlIndex      = 0
	usernameIndex = 1
	passwordIndex = 2
)

type ConfigPage struct {
	ctx context.Context

	conf       *config.Config
	confClient config.ConfigClient

	focusIndex int
	inputs     []textinput.Model
	cursorMode cursor.Mode
	quitting   bool
}

func NewConfigPage(ctx context.Context, conf *config.Config) ConfigPage {
	// The TUI opens this page on first run, before a config file exists.
	if conf == nil {
		conf = &config.Config{}
	}

	var confClient config.FileConfig

	c := ConfigPage{
		ctx:        ctx,
		conf:       conf,
		confClient: &confClient,
		inputs:     make([]textinput.Model, 3),
	}

	var t textinput.Model
	for i := range c.inputs {
		t = textinput.New()
		t.SetWidth(32)
		t.CharLimit = 32

		s := t.Styles()
		s.Cursor.Color = lipgloss.Color("205")
		s.Focused.Prompt = focusedStyle
		s.Focused.Text = focusedStyle
		s.Blurred.Prompt = blurredStyle
		s.Focused.Text = focusedStyle
		t.SetStyles(s)

		switch i {
		case urlIndex:
			t.Placeholder = "Jira URL"
			t.SetValue(conf.JiraHost)
			t.Focus()
		case usernameIndex:
			t.Placeholder = "Username"
			t.CharLimit = 64
			t.SetValue(conf.Username)
		case passwordIndex:
			t.Placeholder = "Password"
			t.EchoMode = textinput.EchoPassword
			t.EchoCharacter = '•'
		}

		c.inputs[i] = t
	}

	return c
}

func (c ConfigPage) Init() tea.Cmd {
	return nil
}

func (c ConfigPage) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	if keypress, ok := msg.(tea.KeyPressMsg); ok {
		switch keypress.String() {
		case "ctrl+c", "q":
			return c, tea.Quit

		// Change cursor mode
		case "ctrl+r":
			c.cursorMode++
			if c.cursorMode > cursor.CursorHide {
				c.cursorMode = cursor.CursorBlink
			}
			cmds := make([]tea.Cmd, len(c.inputs))
			for i := range c.inputs {
				s := c.inputs[i].Styles()
				s.Cursor.Blink = c.cursorMode == cursor.CursorBlink
				c.inputs[i].SetStyles(s)
			}
			return c, tea.Batch(cmds...)

		// Set focus to next input
		case "tab", "shift+tab", "enter", "up", "down":
			s := keypress.String()

			// Did the user press enter while the submit button was focused?
			// If so, exit.
			if s == "enter" && c.focusIndex == len(c.inputs) {
				password := c.inputs[passwordIndex].Value()

				c.conf.Username = c.inputs[usernameIndex].Value()
				c.conf.JiraHost = c.inputs[urlIndex].Value()
				if err := c.confClient.SetConfig(*c.conf); err != nil {
					return c, ThrowErrorMsg(err)
				}

				if err := keychain.SetPassword(c.conf.Username, password); err != nil {
					return c, ThrowErrorMsg(err)
				}

				return c, ui.NavigateToMsg(ui.HomeRoute)
			}

			// Cycle indexes
			if s == "up" || s == "shift+tab" {
				c.focusIndex--
			} else {
				c.focusIndex++
			}

			if c.focusIndex > len(c.inputs) {
				c.focusIndex = 0
			} else if c.focusIndex < 0 {
				c.focusIndex = len(c.inputs)
			}

			cmds := make([]tea.Cmd, len(c.inputs))
			for i := 0; i <= len(c.inputs)-1; i++ {
				if i == c.focusIndex {
					// Set focused state
					cmds[i] = c.inputs[i].Focus()
					continue
				}
				// Remove focused state
				c.inputs[i].Blur()
			}

			return c, tea.Batch(cmds...)
		}
	}

	// Handle character input and blinking
	cmd := c.updateInputs(msg)

	return c, cmd
}

func (c *ConfigPage) updateInputs(msg tea.Msg) tea.Cmd {
	cmds := make([]tea.Cmd, len(c.inputs))

	// Only text inputs with Focus() set will respond, so it's safe to simply
	// update all of them here without any further logic.
	for i := range c.inputs {
		c.inputs[i], cmds[i] = c.inputs[i].Update(msg)
	}

	return tea.Batch(cmds...)
}

func (c ConfigPage) View() tea.View {
	var b strings.Builder
	var cur *tea.Cursor

	b.WriteString("Enter your configuration\n\n")

	for i, in := range c.inputs {
		b.WriteString(c.inputs[i].View())
		if i < len(c.inputs)-1 {
			b.WriteRune('\n')
		}
		if c.cursorMode != cursor.CursorHide && in.Focused() {
			cur = in.Cursor()
			if cur != nil {
				cur.Y += i
			}
		}
	}

	button := &blurredButton
	if c.focusIndex == len(c.inputs) {
		button = &focusedButton
	}
	fmt.Fprintf(&b, "\n\n%s\n\n", *button)

	b.WriteString(helpStyle.Render("cursor mode is "))
	b.WriteString(cursorModeHelpStyle.Render(c.cursorMode.String()))
	b.WriteString(helpStyle.Render(" (ctrl+r to change style)"))

	if c.quitting {
		b.WriteRune('\n')
	}

	v := tea.NewView(b.String())
	v.Cursor = cur
	return v
}
