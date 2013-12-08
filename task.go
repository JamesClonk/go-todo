package main

import "sort"
import "strings"

type Task struct {
	Id          int    `db:"ID"`
	AccountId   int    `db:"ACCOUNT_ID"`
	Created     int    `db:"CREATED"`
	LastUpdated int    `db:"LAST_UPDATED"`
	Priority    int    `db:"PRIORITY"`
	Task        string `db:"TASK"`
}

type Tasks []Task

type taskSort struct {
	tasks Tasks
	by    func(t1, t2 *Task) bool
}

func (t *taskSort) Len() int {
	return len(t.tasks)
}

func (t *taskSort) Swap(a, b int) {
	t.tasks[a], t.tasks[b] = t.tasks[b], t.tasks[a]
}

func (t *taskSort) Less(a, b int) bool {
	return t.by(&t.tasks[a], &t.tasks[b])
}

func (t Tasks) sortBy(by func(t1, t2 *Task) bool) *Tasks {
	ts := &taskSort{
		tasks: t,
		by:    by,
	}
	sort.Sort(ts)
	return &t
}

func (t *Tasks) SortByAccountId(order string) *Tasks {
	t.sortBy(func(t1, t2 *Task) bool {
		if order == "DESC" {
			return t1.AccountId > t2.AccountId
		}
		return t1.AccountId < t2.AccountId
	})
	return t
}

func (t *Tasks) SortByCreated(order string) *Tasks {
	t.sortBy(func(t1, t2 *Task) bool {
		if order == "DESC" {
			return t1.Created > t2.Created
		}
		return t1.Created < t2.Created
	})
	return t
}

func (t *Tasks) SortByLastUpdated(order string) *Tasks {
	t.sortBy(func(t1, t2 *Task) bool {
		if order == "DESC" {
			return t1.LastUpdated > t2.LastUpdated
		}
		return t1.LastUpdated < t2.LastUpdated
	})
	return t
}

func (t *Tasks) SortByPriority(order string) *Tasks {
	t.sortBy(func(t1, t2 *Task) bool {
		if order == "DESC" {
			return t1.Priority > t2.Priority
		}
		return t1.Priority < t2.Priority
	})
	return t
}

func (t *Tasks) SortByTask(order string) *Tasks {
	t.sortBy(func(t1, t2 *Task) bool {
		if order == "DESC" {
			return strings.ToLower(t1.Task) > strings.ToLower(t2.Task)
		}
		return strings.ToLower(t1.Task) < strings.ToLower(t2.Task)
	})
	return t
}
