package main

type Task struct {
	Status    status
	Name      string
	Desc      string
	createdAt string
	updatedAt string
	filterBy  filterBy
}

func NewTask(status status, title, description string) Task {
	return Task{
		Status: status,
		Name:   title,
		Desc:   description,
	}
}

func (t Task) FilterValue() string {
	switch t.filterBy {
	case title:
		return t.Name
	case description:
		return t.Desc
	}
	return t.Name
}

func (t Task) Title() string {
	return t.Name
}

func (t Task) Description() string {
	return t.Desc
}

func (t *Task) Next() {
	if t.Status == done {
		return
	}
	t.Status++
}

func (t *Task) Prev() {
	if t.Status == todo {
		return
	}
	t.Status--
}
