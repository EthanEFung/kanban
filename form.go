package main

import (
	"github.com/charmbracelet/bubbles/help"
	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/textarea"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type formKeys struct {
	Next key.Binding
	Prev key.Binding
	Esc  key.Binding
}

var formKeyMap = formKeys{
	Next: key.NewBinding(
		key.WithKeys("enter", "tab"),
		key.WithHelp("enter/tab", "next"),
	),
	Prev: key.NewBinding(
		key.WithKeys("shift+tab"),
		key.WithHelp("shift+tab", "prev"),
	),
	Esc: key.NewBinding(
		key.WithKeys("ctrl+c"),
		key.WithHelp("ctrl+c", "escape"),
	),
}

func (k formKeys) ShortHelp() []key.Binding {
	return []key.Binding{k.Next, k.Prev, k.Esc}
}

func (k formKeys) FullHelp() [][]key.Binding {
	return [][]key.Binding{{k.Next, k.Prev, k.Esc}}
}

type formModel struct {
	keys        formKeys
	help        help.Model
	focused     status
	title       textinput.Model
	description textarea.Model
}

func NewForm(focused status, title, description string) *formModel {
	titleInput := textinput.New()
	titleInput.Placeholder = "Task Name"
	titleInput.CharLimit = 32

	descriptionArea := textarea.New()
	descriptionArea.Placeholder = "Task Description"
	descriptionArea.CharLimit = 32

	form := &formModel{
		keys:        formKeyMap,
		help:        help.New(),
		focused:     focused,
		title:       titleInput,
		description: descriptionArea,
	}

	form.title.SetValue(title)
	form.description.SetValue(description)
	form.title.Focus()
	return form
}

func (m formModel) Init() tea.Cmd {
	return nil
}

func (m formModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c":
			// TODO: instead of sending a quit message, instead lets redirect the user
			// back to the board.
			models[form] = m
			return models[board], m.CreateTask
		case "enter", "tab":
			if m.title.Focused() {
				m.title.Blur()
				m.description.Focus()
				return m, textarea.Blink
			} else {
				models[form] = m
				return models[board], m.CreateTask
			}
		case "shift+tab":
			if m.description.Focused() {
				m.description.Blur()
				m.title.Focus()
				return m, textinput.Blink
			}
		}
	}
	if m.title.Focused() {
		m.title, cmd = m.title.Update(msg)
		return m, cmd
	} else {
		m.description, cmd = m.description.Update(msg)
		return m, cmd
	}
}

func (m formModel) View() string {
	formView := focusedStyle.Render(lipgloss.JoinVertical(lipgloss.Left, m.title.View(), formDescStyle.Render(m.description.View())))
	helpView := helpStyle.Render(m.help.View(m.keys))
	return lipgloss.JoinVertical(lipgloss.Left, formView, helpView)
}

func (m formModel) CreateTask() tea.Msg {
	if len(m.title.Value()) == 0 {
		return nil
	}
	task := NewTask(m.focused, m.title.Value(), m.description.Value())
	return task
}
