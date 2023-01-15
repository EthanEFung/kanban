package main

import (
	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type boardModel struct {
	focused status
	lists   []list.Model
	deleted list.Model
	tasks   []Task
}

func NewBoard(tasks []Task) *boardModel {
	return &boardModel{
		tasks: tasks,
	}
}

func (m boardModel) Init() tea.Cmd {
	return nil
}

func (m boardModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		columnStyle.Width(msg.Width / divisor)
		focusedStyle.Width(msg.Width / divisor)
		columnStyle.Height(msg.Height - divisor)
		focusedStyle.Height(msg.Height - divisor)
		if len(m.lists) == 0 {
			m.lists = m.newLists(msg.Width, msg.Height)
			m.deleted = m.newDeleted()
		}
	case tea.KeyMsg:
		switch msg.String() {
		case "h":
			m.focused = m.Prev()
		case "l":
			m.focused = m.Next()
		case "o": // create new
			models[board] = m // save the state of the board
			models[form] = NewForm(m.focused, "", "")
			return models[form].Update(nil)
		case "x":
			return m, m.Delete
		case "u":
			return m, m.Undo
		case "L":
			return m, m.NextList
		case "H":
			return m, m.PrevList
		case "J":
			return m, m.MoveDown
		case "K":
			return m, m.MoveUp
		case "i":
			models[board] = m
			title, desc := m.Edit()
			models[form] = NewForm(m.focused, title, desc)
			return models[form].Update(nil)
		}
	case Task:
		task := msg
		return m, m.lists[task.status].InsertItem(len(m.lists[task.status].Items()), task)
	case status:
		m.focused = msg
	}
	var cmd tea.Cmd
	m.lists[m.focused], cmd = m.lists[m.focused].Update(msg)
	return m, cmd
}

func (m boardModel) View() string {
	if len(m.lists) == 0 {
		return "loading..."
	}
	views := []string{}
	for _, status := range statuses {
		if status == m.focused {
			views = append(views, focusedStyle.Render(m.lists[status].View()))
			continue
		}
		views = append(views, columnStyle.Render(m.lists[status].View()))
	}

	return lipgloss.JoinHorizontal(lipgloss.Left, views...)
}

func (m boardModel) Prev() status {
	if m.focused == todo {
		return m.focused
	}
	m.lists[m.focused].Select(-1)
	m.focused--
	if m.lists[m.focused].Index() == -1 {
		m.lists[m.focused].Select(0)
	}
	return m.focused
}

func (m boardModel) Next() status {
	if m.focused == done {
		return m.focused
	}
	m.lists[m.focused].Select(-1)
	m.focused++
	if m.lists[m.focused].Index() == -1 {
		m.lists[m.focused].Select(0)
	}
	return m.focused
}

func (m *boardModel) Delete() tea.Msg {
	if len(m.lists[m.focused].VisibleItems()) > 0 {
		selected := m.lists[m.focused].SelectedItem().(Task)
		m.lists[selected.status].RemoveItem(m.lists[m.focused].Index())
		m.deleted.InsertItem(len(m.deleted.Items())-1, list.Item(selected))
	}
	return nil
}

func (m *boardModel) Undo() tea.Msg { // we'll work on this
	return nil
}

func (m *boardModel) NextList() tea.Msg {
	selected := m.lists[m.focused].SelectedItem()
	if selected == nil {
		return nil
	}
	task := selected.(Task)
	prev := task.status
	m.lists[task.status].RemoveItem(m.lists[m.focused].Index())
	task.Next()
	if prev != task.status {
		m.lists[prev].Select(-1)
	}
	m.lists[task.status].InsertItem(len(m.lists[task.status].Items()), list.Item(task))
	m.lists[task.status].Select(len(m.lists[task.status].Items()))
	return task.status
}

func (m *boardModel) PrevList() tea.Msg {
	selected := m.lists[m.focused].SelectedItem()
	if selected == nil {
		return nil
	}
	task := selected.(Task)
	prev := task.status
	if prev != task.status {
		m.lists[prev].Select(-1)
	}
	m.lists[task.status].RemoveItem(m.lists[m.focused].Index())
	task.Prev()
	m.lists[task.status].InsertItem(len(m.lists[task.status].Items()), list.Item(task))
	m.lists[task.status].Select(len(m.lists[task.status].Items()))
	return task.status
}

func (m *boardModel) MoveDown() tea.Msg {
	selected := m.lists[m.focused].SelectedItem()
	if selected == nil {
		return nil
	}
	task := selected.(Task)
	index := m.lists[m.focused].Index()
	if index == len(m.lists[m.focused].Items())-1 {
		return nil
	}
	index += 2
	m.lists[task.status].InsertItem(index, list.Item(task))
	m.lists[task.status].Select(index - 1)
	m.lists[task.status].RemoveItem(index - 2)
	return task.status
}

func (m *boardModel) MoveUp() tea.Msg {
	selected := m.lists[m.focused].SelectedItem()
	if selected == nil {
		return nil
	}
	task := selected.(Task)
	index := m.lists[m.focused].Index()
	if index == 0 {
		return nil
	}
	m.lists[task.status].InsertItem(index-1, list.Item(task))
	m.lists[task.status].Select(index - 1)
	m.lists[task.status].RemoveItem(index + 1)
	return task.status
}

func (m *boardModel) Edit() (string, string) {
	selected := m.lists[m.focused].SelectedItem()
	if selected == nil {
		return "", ""
	}
	task := selected.(Task)
	m.lists[task.status].RemoveItem(m.lists[task.status].Index())
	return task.Title(), task.Description() 
}

func (m boardModel) newLists(width, height int) []list.Model {
	l := list.New([]list.Item{}, list.NewDefaultDelegate(), width/divisor, height/2)
	l.SetShowHelp(false)
	lists := []list.Model{l, l, l}
	for _, status := range statuses {
		lists[status].Title = titleCaser.String(status.String())
		lists[status].SetItems(m.filterTasks(status))
	}
	return lists
}

func (m boardModel) filterTasks(status status) []list.Item {
	result := []list.Item{}
	for _, task := range m.tasks {
		if task.status != status {
			continue
		}
		result = append(result, task)
	}
	return result
}

func (m boardModel) newDeleted() list.Model {
	return list.New([]list.Item{}, list.NewDefaultDelegate(), 0, 0)
}