package ui

import (
	"fmt"
	"io"
	"strings"

	"berichtsheft/internal/config"

	"charm.land/bubbles/v2/list"
	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
)

const listHeight = 14

type styles struct {
	title        lipgloss.Style
	item         lipgloss.Style
	selectedItem lipgloss.Style
	pagination   lipgloss.Style
	help         lipgloss.Style
	quitText     lipgloss.Style
}

func newStyles(darkBG bool) styles {
	var s styles
	s.title = lipgloss.NewStyle().MarginLeft(2)
	s.item = lipgloss.NewStyle().PaddingLeft(4)
	s.selectedItem = lipgloss.NewStyle().PaddingLeft(2).Foreground(lipgloss.Color("170"))
	s.pagination = list.DefaultStyles(darkBG).PaginationStyle.PaddingLeft(4)
	s.help = list.DefaultStyles(darkBG).HelpStyle.PaddingLeft(4).PaddingBottom(1)
	s.quitText = lipgloss.NewStyle().Margin(1, 0, 2, 4)
	return s
}

type item string

func (i item) FilterValue() string { return "" }

type itemDelegate struct {
	styles *styles
}

func (d itemDelegate) Height() int                             { return 1 }
func (d itemDelegate) Spacing() int                            { return 0 }
func (d itemDelegate) Update(_ tea.Msg, _ *list.Model) tea.Cmd { return nil }
func (d itemDelegate) Render(w io.Writer, m list.Model, index int, listItem list.Item) {
	i, ok := listItem.(item)
	if !ok {
		return
	}

	str := fmt.Sprintf("%d. %s", index+1, i)

	fn := d.styles.item.Render
	if index == m.Index() {
		fn = func(s ...string) string {
			return d.styles.selectedItem.Render("> " + strings.Join(s, " "))
		}
	}

	fmt.Fprint(w, fn(str))
}

type HomePage struct {
	conf     *config.Config
	list     list.Model
	choice   string
	styles   styles
	quitting bool
}

func NewHomePage(conf *config.Config) HomePage {
	items := []list.Item{
		item("Start"),
		item("Edit configuration"),
	}

	const defaultWidth = 20

	l := list.New(items, itemDelegate{}, defaultWidth, listHeight)
	l.SetShowStatusBar(false)
	l.SetFilteringEnabled(false)

	if conf == nil {
		l.Title = "Lets do something"
	} else {
		l.Title = "Hello " + conf.Username
	}

	m := HomePage{list: l, conf: conf}
	m.updateStyles(true) // default to dark styles.
	return m
}

func (m *HomePage) updateStyles(isDark bool) {
	m.styles = newStyles(isDark)
	m.list.Styles.Title = m.styles.title
	m.list.Styles.PaginationStyle = m.styles.pagination
	m.list.Styles.HelpStyle = m.styles.help
	m.list.SetDelegate(itemDelegate{styles: &m.styles})
}

func (m HomePage) Init() tea.Cmd {
	return nil
}

func (m HomePage) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.list.SetWidth(msg.Width)
		return m, nil

	case tea.KeyPressMsg:
		switch keypress := msg.String(); keypress {
		case "q", "ctrl+c":
			m.quitting = true
			return m, tea.Quit

		case "enter":
			// TODO: send a message upwards here
			i, ok := m.list.SelectedItem().(item)
			if ok {
				switch string(i) {
				case "Start":
					return NewStartPage(m.conf), nil
				case "Edit configuration":
					return NewConfigPage(m.conf), nil
				default:
					return m, nil
				}
			}
			return m, tea.Quit
		}
	}

	var cmd tea.Cmd
	m.list, cmd = m.list.Update(msg)
	return m, cmd
}

func (m HomePage) View() tea.View {
	if m.quitting {
		return tea.NewView(m.styles.quitText.Render("Quitting..."))
	}
	return tea.NewView("\n" + m.list.View())
}
