package main

import (
	"log"
	"os"
	"time"

	"github.com/charmbracelet/bubbles/help"
	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type boardKeymap struct {
	FocusLeft   key.Binding
	FocusRight  key.Binding
	FocusDown   key.Binding
	FocusUp     key.Binding
	NewTask     key.Binding
	EditTask    key.Binding
	RemoveTask  key.Binding
	UndoRemove  key.Binding
	MoveLeft    key.Binding
	MoveRight   key.Binding
	MoveDown    key.Binding
	MoveUp      key.Binding
	SaveAndQuit key.Binding
}

var boardKeys = boardKeymap{
	FocusUp: key.NewBinding(
		key.WithKeys("k"),
		key.WithHelp("k", "focus up"),
	),
	FocusRight: key.NewBinding(
		key.WithKeys("l"),
		key.WithHelp("l", "focus right"),
	),
	FocusDown: key.NewBinding(
		key.WithKeys("j"),
		key.WithHelp("j", "focus down"),
	),
	FocusLeft: key.NewBinding(
		key.WithKeys("h"),
		key.WithHelp("h", "focus left"),
	),
	MoveUp: key.NewBinding(
		key.WithKeys("K"),
		key.WithHelp("K", "move up"),
	),
	MoveRight: key.NewBinding(
		key.WithKeys("L"),
		key.WithHelp("L", "move right"),
	),
	MoveDown: key.NewBinding(
		key.WithKeys("J"),
		key.WithHelp("J", "move down"),
	),
	MoveLeft: key.NewBinding(
		key.WithKeys("H"),
		key.WithHelp("H", "move left"),
	),
	NewTask: key.NewBinding(
		key.WithKeys("o"),
		key.WithHelp("o", "new task"),
	),
	EditTask: key.NewBinding(
		key.WithKeys("i"),
		key.WithHelp("i", "edit task"),
	),
	RemoveTask: key.NewBinding(
		key.WithKeys("x"),
		key.WithHelp("x", "remove task"),
	),
	UndoRemove: key.NewBinding(
		key.WithKeys("u"),
		key.WithHelp("u", "undo remove"),
	),
	SaveAndQuit: key.NewBinding(
		key.WithKeys("q", "ctrl+c"),
		key.WithHelp("q/ctrl+c", "save and quit"),
	),
}

func (k boardKeymap) ShortHelp() []key.Binding {
	return []key.Binding{
		k.FocusUp, k.FocusRight, k.FocusDown, k.FocusLeft,
		k.MoveUp, k.MoveRight, k.MoveDown, k.MoveLeft,
		k.NewTask, k.EditTask, k.RemoveTask, k.UndoRemove,
		k.SaveAndQuit}
}

func (k boardKeymap) FullHelp() [][]key.Binding {
	return [][]key.Binding{
		{k.FocusUp, k.FocusRight, k.FocusDown, k.FocusLeft},
		{k.MoveUp, k.MoveRight, k.MoveDown, k.MoveLeft},
		{k.NewTask, k.EditTask, k.RemoveTask, k.UndoRemove},
		{k.SaveAndQuit},
	}
}

type TickMsg time.Time

func tick() tea.Cmd {
	return tea.Tick(time.Second*5, func(t time.Time) tea.Msg {
		return TickMsg(t)
	})
}

type boardModel struct {
	keys     boardKeymap
	help     help.Model
	quitting bool
	focused  status
	lists    []list.Model
	deleted  []Task
	tasks    []Task
	last     Task
}

func NewBoard(tasks []Task) *boardModel {
	return &boardModel{
		keys:  boardKeys,
		help:  help.New(),
		tasks: tasks,
	}
}

func (m boardModel) Init() tea.Cmd {
	return tick()
}

func (m boardModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		columnStyle.Width(msg.Width / divisor)
		focusedStyle.Width(msg.Width / divisor)
		columnStyle.Height(msg.Height - 12)
		focusedStyle.Height(msg.Height - 12)
		helpStyle.Width(msg.Width)
		m.lists = m.newLists(msg.Width, msg.Height-12)
		if len(m.lists) == 0 {
			m.deleted = m.newDeleted()
		}
	case tea.KeyMsg:
		switch {
		// focuses
		case key.Matches(msg, m.keys.FocusLeft):
			m.focused = m.Prev()
		case key.Matches(msg, m.keys.FocusRight):
			m.focused = m.Next()

		// just utilizing the the native list keys for focusing down and up
		// a list

		// task movements
		case key.Matches(msg, m.keys.MoveUp):
			return m, m.MoveUp
		case key.Matches(msg, m.keys.MoveRight):
			return m, m.NextList
		case key.Matches(msg, m.keys.MoveDown):
			return m, m.MoveDown
		case key.Matches(msg, m.keys.MoveLeft):
			return m, m.PrevList

		// task crud
		case key.Matches(msg, m.keys.NewTask):
			models[board] = m // save the state of the board
			models[form] = NewForm(m.focused, "", "")
			return models[form].Update(nil)
		case key.Matches(msg, m.keys.EditTask):
			models[board] = m
			title, desc := m.Edit()
			models[form] = NewForm(m.focused, title, desc)
			return models[form].Update(nil)
		case key.Matches(msg, m.keys.RemoveTask):
			item := m.lists[m.focused].SelectedItem()
			if item == nil {
				return m, nil
			}
			selected := item.(Task)                 // first update the board model
			m.deleted = append(m.deleted, selected) // then update the list model
			return m, m.Delete
		case key.Matches(msg, m.keys.UndoRemove):
			l := len(m.deleted)
			if l == 0 {
				return m, nil
			}
			m.last = m.deleted[l-1]
			m.deleted = m.deleted[:l-1]
			m.lists[m.focused].Select(-1)
			m.focused = m.last.Status
			return m, m.Undo

		// misc
		case key.Matches(msg, m.keys.SaveAndQuit):
			m.quitting = true
			return m, m.GracefulShutdown
		}
	case Task:
		task := msg
		m.lists[task.Status].Select(len(m.lists[task.Status].Items()))
		return m, m.lists[task.Status].InsertItem(len(m.lists[task.Status].Items()), task)
	case status:
		m.focused = msg
	case TickMsg:
		// autosave
		m.Save()
		return m, tick()
	}
	var cmd tea.Cmd
	m.lists[m.focused], cmd = m.lists[m.focused].Update(msg)
	return m, cmd
}

func (m boardModel) View() string {
	if len(m.lists) == 0 {
		return "loading..."
	}
	if m.quitting {
		return ""
	}
	views := []string{}
	for _, status := range statuses {
		if status == m.focused {
			views = append(views, focusedStyle.Render(m.lists[status].View()))
			continue
		}
		views = append(views, columnStyle.Render(m.lists[status].View()))
	}
	boardViews := lipgloss.JoinHorizontal(lipgloss.Left, views...)
	helpView := helpStyle.Render(m.help.View(m.keys))

	return lipgloss.JoinVertical(lipgloss.Left, boardViews, helpView)
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
		m.lists[selected.Status].RemoveItem(m.lists[m.focused].Index())
	}
	return nil
}

func (m *boardModel) Undo() tea.Msg { // we'll work on this
	l := len(m.lists[m.focused].Items())
	m.lists[m.focused].Select(l)
	return m.last
}

func (m *boardModel) NextList() tea.Msg {
	selected := m.lists[m.focused].SelectedItem()
	if selected == nil {
		return nil
	}
	task := selected.(Task)
	prev := task.Status
	m.lists[task.Status].RemoveItem(m.lists[m.focused].Index())
	task.Next()
	if prev != task.Status {
		m.lists[prev].Select(-1)
	}
	m.lists[task.Status].InsertItem(len(m.lists[task.Status].Items()), list.Item(task))
	m.lists[task.Status].Select(len(m.lists[task.Status].Items()))
	return task.Status
}

func (m *boardModel) PrevList() tea.Msg {
	selected := m.lists[m.focused].SelectedItem()
	if selected == nil {
		return nil
	}
	task := selected.(Task)
	prev := task.Status
	if prev != task.Status {
		m.lists[prev].Select(-1)
	}
	m.lists[task.Status].RemoveItem(m.lists[m.focused].Index())
	task.Prev()
	m.lists[task.Status].InsertItem(len(m.lists[task.Status].Items()), list.Item(task))
	m.lists[task.Status].Select(len(m.lists[task.Status].Items()))
	return task.Status
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
	m.lists[task.Status].InsertItem(index, list.Item(task))
	m.lists[task.Status].Select(index - 1)
	m.lists[task.Status].RemoveItem(index - 2)
	return task.Status
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
	m.lists[task.Status].InsertItem(index-1, list.Item(task))
	m.lists[task.Status].Select(index - 1)
	m.lists[task.Status].RemoveItem(index + 1)
	return task.Status
}

func (m *boardModel) Edit() (string, string) {
	selected := m.lists[m.focused].SelectedItem()
	if selected == nil {
		return "", ""
	}
	task := selected.(Task)
	m.lists[task.Status].RemoveItem(m.lists[task.Status].Index())
	return task.Title(), task.Description()
}

func (m *boardModel) Save() error {
	gobbler := Gobble{}
	f, err := os.Create(filepath)
	if err != nil {
		return err
	}
	if err := gobbler.saveTasks(f, m.deriveTasks()); err != nil {
		return err
	}
	return nil
}

func (m *boardModel) GracefulShutdown() tea.Msg {
	err := m.Save()
	if err != nil {
		log.Fatal(err)
	}
	return tea.Quit()
}

func (m boardModel) newLists(width, height int) []list.Model {
	l := list.New([]list.Item{}, list.NewDefaultDelegate(), width/divisor, height)
	l.SetShowHelp(false)
	lists := []list.Model{l, l, l}
	for _, status := range statuses {
		lists[status].Title = titleCaser.String(status.String())
		lists[status].SetItems(m.filterTasks(status))
		lists[status].Select(-1)
	}
	lists[todo].Select(0)
	return lists
}

func (m boardModel) filterTasks(status status) []list.Item {
	result := []list.Item{}
	for _, task := range m.tasks {
		if task.Status != status {
			continue
		}
		result = append(result, task)
	}
	return result
}

func (m boardModel) newDeleted() []Task {
	return []Task{}
}

func (m boardModel) deriveTasks() []Task {
	tasks := []Task{}
	for _, list := range m.lists {
		for _, item := range list.Items() {
			if task, ok := item.(Task); ok {
				tasks = append(tasks, task)
			}
		}
	}
	return tasks
}
