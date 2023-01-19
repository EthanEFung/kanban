package main

import (
	"github.com/charmbracelet/bubbles/textarea"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type formModel struct {
	focused     status
	title       textinput.Model
	description textarea.Model
}

func NewForm(focused status, title, description string) *formModel {

	form := &formModel{
		focused:     focused,
		title:       textinput.New(),
		description: textarea.New(),
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
			return m, tea.Quit
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
	return lipgloss.JoinVertical(lipgloss.Left, m.title.View(), m.description.View())
}

func (m formModel) CreateTask() tea.Msg {
	if len(m.title.Value()) == 0 {
		return nil
	}
	task := NewTask(m.focused, m.title.Value(), m.description.Value())
	return task
}
