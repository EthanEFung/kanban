package main

import (
	"encoding/gob"
	"fmt"
	"io"
	"log"
	"os"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"golang.org/x/text/cases"
	"golang.org/x/text/language"
)

type status int

//go:generate stringer -type=status
const (
	todo status = iota
	doing
	done
)

var statuses = []status{todo, doing, done}

type filterBy int

const (
	title filterBy = iota
	description
	createdAt
	updatedAt
)

var models []tea.Model

const (
	board status = iota
	form
)

const divisor int = 4

var titleCaser = cases.Title(language.English)

var (
	columnStyle  = lipgloss.NewStyle().Padding(1, 2).Border(lipgloss.HiddenBorder())
	focusedStyle = columnStyle.Copy().Border(lipgloss.RoundedBorder()).BorderForeground(lipgloss.Color("62"))
)

const filepath = ".kanban/tasks.bin"

func mapTasks(path string) []Task {
	f, err := os.OpenFile(path, os.O_RDONLY, os.ModePerm)
	if err != nil {
		log.Fatal(err)
	}
	gobble := Gobble{}
	tasks, err := gobble.readTasks(f) 
	if err != nil {
		log.Fatalf("Could not read tasks %v", err)
	}
	return tasks
}

func initialTasks() []Task {
	_, err := os.Stat(filepath)
	if err != nil {
		// create files
		f, err := os.Create(filepath)
		if err != nil {
			log.Fatal(err)
		}
		fmt.Printf("created %s\n", f.Name())

		return []Task{}
	}
	return mapTasks(filepath);
}

type Gobble struct {
	path string
}

func (g Gobble) saveTasks(wr io.Writer, tasks []Task) error {
	enc := gob.NewEncoder(wr)
	for _, task := range tasks {
		if err := enc.Encode(task); err != nil {
			return err
		}
	}
	return nil
}

func (g Gobble) readTasks(r io.Reader) ([]Task, error) {
	dec := gob.NewDecoder(r)
	tasks := []Task{}
	var curr Task
	var err error
	for {
		 if err = dec.Decode(&curr); err != nil {
			break
		}
		tasks = append(tasks, curr)
	}
	if err != io.EOF && err != io.ErrUnexpectedEOF {
		return tasks, err
	}
	return tasks, nil
}

func main() {
	tasks := initialTasks()
	b, f := NewBoard(tasks), NewForm(todo, "", "")
	models = []tea.Model{b, f}
	p := tea.NewProgram(models[board])
	if _, err := p.Run(); err != nil {
		fmt.Println("Error running program:", err)
		os.Exit(1)
	}
}
