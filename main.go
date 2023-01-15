package main

import (
	"encoding/csv"
	"fmt"
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

func mapTasks(path string) []Task {
	f, err := os.OpenFile(path, os.O_RDONLY, os.ModePerm)
	if err != nil {
		log.Fatal(err)
	}
	r := csv.NewReader(f)
	records, err := r.ReadAll()
	if err != nil {
		log.Fatal(err)
	}

	tasks := []Task{}
	for _, record := range records {
		var status status;
		switch(record[2]) {
		case "todo":
			status = todo;
		case "doing":
			status = doing;
		case "done":
			status = done;
		default:
			log.Fatalf("Unknown status of %s was declared, but unsupported", record[2])
		}
		tasks = append(tasks, NewTask(status, record[0], record[1]))
	}
	return tasks
}

func initialTasks() []Task {
	_, err := os.Stat(".kanban/")

	if err != nil {

		err = os.MkdirAll(".kanban/", os.ModePerm)
		if err != nil {
			log.Fatal(err)
		}

		// create files
		f, err := os.Create(".kanban/items.csv")
		if err != nil {
			log.Fatal(err)
		}
		fmt.Printf("created %s\n", f.Name())

		return []Task{}
	}

	return mapTasks(".kanban/items.csv");
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
