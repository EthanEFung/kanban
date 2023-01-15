package main

type Task struct {
	status      status
	title       string
	description string
	createdAt   string
	updatedAt   string
	filterBy    filterBy
}

func NewTask(status status, title, description string) Task {
	return Task{
		status:      status,
		title:       title,
		description: description,
	}
}

func (t Task) FilterValue() string {
	switch t.filterBy {
	case title:
		return t.title
	case description:
		return t.description
	}
	return t.title
}

func (t Task) Title() string {
	return t.title
}

func (t Task) Description() string {
	return t.description
}

func (t *Task) Next() {
	if t.status == done {
		return
	}
	t.status++
}

func (t *Task) Prev() {
	if t.status == todo {
		return
	}
	t.status--
}
