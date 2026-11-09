package tui

import (
	"fmt"
	"strings"
	"time"

	"berichtsheft/internal/ui"

	"charm.land/bubbles/v2/cursor"
	"charm.land/bubbles/v2/textinput"
	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
)

type DateInputPage struct {
	focusIndex    int
	inputHeadings []string
	inputs        []textinput.Model
	cursorMode    cursor.Mode
	quitting      bool
}

func NewDateInputPage() DateInputPage {
	m := DateInputPage{
		inputs:        make([]textinput.Model, 2),
		inputHeadings: make([]string, 2),
	}

	var t textinput.Model
	for i := range m.inputs {
		t = textinput.New()
		t.SetWidth(10)
		t.CharLimit = 10

		s := t.Styles()
		s.Cursor.Color = lipgloss.Color("205")
		s.Focused.Prompt = focusedStyle
		s.Focused.Text = focusedStyle
		s.Blurred.Prompt = blurredStyle
		s.Focused.Text = focusedStyle
		t.SetStyles(s)

		switch i {
		case 0:
			m.inputHeadings[i] = "From"
			t.Placeholder = "dd.mm.yyyy"
			t.Focus()
		case 1:
			m.inputHeadings[i] = "To"
			t.Placeholder = "dd.mm.yyyy"
		}

		m.inputs[i] = t
	}

	return m
}

func (m DateInputPage) Init() tea.Cmd {
	return textinput.Blink
}

func (m DateInputPage) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	// @gocritic
	switch msg := msg.(type) {
	case tea.KeyPressMsg:
		switch msg.String() {
		case "ctrl+c", "esc":
			m.quitting = true
			return m, tea.Quit

		// Set focus to next input
		case "tab", "shift+tab", "enter", "up", "down":
			s := msg.String()

			// Did the user press enter while the submit button was focused?
			// If so, exit.
			if s == "enter" && m.focusIndex == len(m.inputs) {
				from, err := time.Parse("02.01.2006", m.inputs[0].Value())
				if err != nil {
					return m, ThrowErrorMsg(err)
				}

				to, err := time.Parse("02.01.2006", m.inputs[1].Value())
				if err != nil {
					return m, ThrowErrorMsg(err)
				}

				return m, tea.Sequence(DateMsg(from, to), ui.NavigateToMsg(ui.ProcessRoute))
			}

			// Cycle indexes
			if s == "up" || s == "shift+tab" {
				m.focusIndex--
			} else {
				m.focusIndex++
			}

			if m.focusIndex > len(m.inputs) {
				m.focusIndex = 0
			} else if m.focusIndex < 0 {
				m.focusIndex = len(m.inputs)
			}

			cmds := make([]tea.Cmd, len(m.inputs))
			for i := 0; i <= len(m.inputs)-1; i++ {
				if i == m.focusIndex {
					// Set focused state
					cmds[i] = m.inputs[i].Focus()
					continue
				}
				// Remove focused state
				m.inputs[i].Blur()
			}

			return m, tea.Batch(cmds...)
		}
	}

	// Handle character input and blinking
	cmd := m.updateInputs(msg)

	return m, cmd
}

func (m *DateInputPage) updateInputs(msg tea.Msg) tea.Cmd {
	cmds := make([]tea.Cmd, len(m.inputs))

	// Only text inputs with Focus() set will respond, so it's safe to simply
	// update all of them here without any further logic.
	for i := range m.inputs {
		m.inputs[i], cmds[i] = m.inputs[i].Update(msg)
	}

	return tea.Batch(cmds...)
}

func (m DateInputPage) View() tea.View {
	var sb strings.Builder
	var c *tea.Cursor

	lPad := strings.Repeat(" ", 5)
	fmt.Fprintf(&sb, "\n\n%sEnter the timeframe you want to report.", lPad)

	fmt.Fprintf(&sb, "\n")
	for i, in := range m.inputs {
		fmt.Fprintf(&sb, "\n%s%s", lPad, m.inputHeadings[i])
		fmt.Fprintf(&sb, "\n%s%s", lPad, m.inputs[i].View())
		if i < len(m.inputs)-1 {
			sb.WriteRune('\n')
		}
		if m.cursorMode != cursor.CursorHide && in.Focused() {
			c = in.Cursor()
			if c != nil {
				c.Y += i
			}
		}
	}

	button := &blurredButton
	if m.focusIndex == len(m.inputs) {
		button = &focusedButton
	}
	fmt.Fprintf(&sb, "\n\n%s%s\n\n", lPad, *button)

	if m.quitting {
		sb.WriteRune('\n')
	}

	v := tea.NewView(sb.String())
	v.Cursor = c
	return v
}
