package main

import "testing"

func Test_task_SortBy(t *testing.T) {
	var ts1 = Tasks{
		Task{1, 1, 1234567890, 1234567895, 3, "A"},
		Task{2, 2, 1234567891, 1234567895, 1, "B..."},
		Task{3, 2, 1234567892, 1234567895, 4, "C!"},
		Task{4, 1, 1234567893, 1234567895, 3, "D?"},
		Task{5, 3, 1234567890, 1234567895, 5, "E$"},
	}

	var ts2 = Tasks{
		Task{5, 3, 1234567890, 1234567895, 5, "E$"},
		Task{3, 2, 1234567892, 1234567895, 4, "C!"},
		Task{1, 1, 1234567890, 1234567895, 3, "A"},
		Task{2, 2, 1234567891, 1234567895, 1, "B..."},
		Task{4, 1, 1234567893, 1234567895, 3, "D?"},
	}

	ts1.sortBy(func(t1, t2 *Task) bool {
		return t1.Created%2 == 0
	})

	for i, t1 := range ts1 {
		if t1 != ts2[i] {
			t.Errorf("SortBy is not as expected: [%v], instead of [%v]", ts1, ts2)
			return
		}
	}
}

func Test_task_SortByAccountId(t *testing.T) {
	var ts1 = Tasks{
		Task{1, 1, 1234567890, 1234567895, 3, "A"},
		Task{2, 2, 1234567891, 1234567895, 1, "B..."},
		Task{3, 2, 1234567892, 1234567895, 4, "C!"},
		Task{4, 1, 1234567893, 1234567895, 3, "D?"},
		Task{5, 3, 1234567890, 1234567895, 5, "E$"},
	}

	var ts2 = Tasks{
		Task{1, 1, 1234567890, 1234567895, 3, "A"},
		Task{4, 1, 1234567893, 1234567895, 3, "D?"},
		Task{2, 2, 1234567891, 1234567895, 1, "B..."},
		Task{3, 2, 1234567892, 1234567895, 4, "C!"},
		Task{5, 3, 1234567890, 1234567895, 5, "E$"},
	}

	var ts3 = Tasks{
		Task{5, 3, 1234567890, 1234567895, 5, "E$"},
		Task{2, 2, 1234567891, 1234567895, 1, "B..."},
		Task{3, 2, 1234567892, 1234567895, 4, "C!"},
		Task{1, 1, 1234567890, 1234567895, 3, "A"},
		Task{4, 1, 1234567893, 1234567895, 3, "D?"},
	}

	ts1.SortByAccountId("ASC")
	for i, t1 := range ts1 {
		if t1 != ts2[i] {
			t.Errorf("SortByAccountId ASC is not as expected: [%v], instead of [%v]", ts1, ts2)
			return
		}
	}

	ts1.SortByAccountId("DESC")
	for i, t1 := range ts1 {
		if t1 != ts3[i] {
			t.Errorf("SortByAccountId DESC is not as expected: [%v], instead of [%v]", ts1, ts3)
			return
		}
	}
}

func Test_task_SortByCreated(t *testing.T) {
	var ts1 = Tasks{
		Task{1, 1, 1234567890, 1234567895, 3, "A"},
		Task{2, 2, 1234567891, 1234567895, 1, "B..."},
		Task{3, 2, 1234567892, 1234567895, 4, "C!"},
		Task{4, 1, 1234567893, 1234567895, 3, "D?"},
		Task{5, 3, 1234567890, 1234567895, 5, "E$"},
	}

	var ts2 = Tasks{
		Task{1, 1, 1234567890, 1234567895, 3, "A"},
		Task{5, 3, 1234567890, 1234567895, 5, "E$"},
		Task{2, 2, 1234567891, 1234567895, 1, "B..."},
		Task{3, 2, 1234567892, 1234567895, 4, "C!"},
		Task{4, 1, 1234567893, 1234567895, 3, "D?"},
	}

	var ts3 = Tasks{
		Task{4, 1, 1234567893, 1234567895, 3, "D?"},
		Task{3, 2, 1234567892, 1234567895, 4, "C!"},
		Task{2, 2, 1234567891, 1234567895, 1, "B..."},
		Task{1, 1, 1234567890, 1234567895, 3, "A"},
		Task{5, 3, 1234567890, 1234567895, 5, "E$"},
	}

	ts1.SortByCreated("ASC")
	for i, t1 := range ts1 {
		if t1 != ts2[i] {
			t.Errorf("SortByCreated ASC is not as expected: [%v], instead of [%v]", ts1, ts2)
			return
		}
	}

	ts1.SortByCreated("DESC")
	for i, t1 := range ts1 {
		if t1 != ts3[i] {
			t.Errorf("SortByCreated DESC is not as expected: [%v], instead of [%v]", ts1, ts3)
			return
		}
	}
}

func Test_task_SortByLastUpdated(t *testing.T) {
	var ts1 = Tasks{
		Task{1, 1, 1234567890, 1234567891, 3, "A"},
		Task{2, 2, 1234567891, 1234567893, 1, "B..."},
		Task{3, 2, 1234567892, 1234567894, 4, "C!"},
		Task{4, 1, 1234567893, 1234567895, 3, "D?"},
		Task{5, 3, 1234567890, 1234567892, 5, "E$"},
	}

	var ts2 = Tasks{
		Task{1, 1, 1234567890, 1234567891, 3, "A"},
		Task{5, 3, 1234567890, 1234567892, 5, "E$"},
		Task{2, 2, 1234567891, 1234567893, 1, "B..."},
		Task{3, 2, 1234567892, 1234567894, 4, "C!"},
		Task{4, 1, 1234567893, 1234567895, 3, "D?"},
	}

	var ts3 = Tasks{
		Task{4, 1, 1234567893, 1234567895, 3, "D?"},
		Task{3, 2, 1234567892, 1234567894, 4, "C!"},
		Task{2, 2, 1234567891, 1234567893, 1, "B..."},
		Task{5, 3, 1234567890, 1234567892, 5, "E$"},
		Task{1, 1, 1234567890, 1234567891, 3, "A"},
	}

	ts1.SortByLastUpdated("ASC")
	for i, t1 := range ts1 {
		if t1 != ts2[i] {
			t.Errorf("SortByLastUpdated ASC is not as expected: [%v], instead of [%v]", ts1, ts2)
			return
		}
	}

	ts1.SortByLastUpdated("DESC")
	for i, t1 := range ts1 {
		if t1 != ts3[i] {
			t.Errorf("SortByLastUpdated DESC is not as expected: [%v], instead of [%v]", ts1, ts3)
			return
		}
	}
}

func Test_task_SortByPriority(t *testing.T) {
	var ts1 = Tasks{
		Task{1, 1, 1234567890, 1234567895, 3, "A"},
		Task{2, 2, 1234567891, 1234567895, 1, "B..."},
		Task{3, 2, 1234567892, 1234567895, 4, "C!"},
		Task{4, 1, 1234567893, 1234567895, 3, "D?"},
		Task{5, 3, 1234567890, 1234567895, 5, "E$"},
	}

	var ts2 = Tasks{
		Task{2, 2, 1234567891, 1234567895, 1, "B..."},
		Task{1, 1, 1234567890, 1234567895, 3, "A"},
		Task{4, 1, 1234567893, 1234567895, 3, "D?"},
		Task{3, 2, 1234567892, 1234567895, 4, "C!"},
		Task{5, 3, 1234567890, 1234567895, 5, "E$"},
	}

	var ts3 = Tasks{
		Task{5, 3, 1234567890, 1234567895, 5, "E$"},
		Task{3, 2, 1234567892, 1234567895, 4, "C!"},
		Task{1, 1, 1234567890, 1234567895, 3, "A"},
		Task{4, 1, 1234567893, 1234567895, 3, "D?"},
		Task{2, 2, 1234567891, 1234567895, 1, "B..."},
	}

	ts1.SortByPriority("ASC")
	for i, t1 := range ts1 {
		if t1 != ts2[i] {
			t.Errorf("SortByPriority ASC is not as expected: [%v], instead of [%v]", ts1, ts2)
			return
		}
	}

	ts1.SortByPriority("DESC")
	for i, t1 := range ts1 {
		if t1 != ts3[i] {
			t.Errorf("SortByPriority DESC is not as expected: [%v], instead of [%v]", ts1, ts3)
			return
		}
	}
}

func Test_task_SortByTask(t *testing.T) {
	var ts1 = Tasks{
		Task{1, 1, 1234567890, 1234567895, 3, "a"},
		Task{2, 2, 1234567891, 1234567895, 1, "C..."},
		Task{3, 2, 1234567892, 1234567895, 4, "b!"},
		Task{4, 1, 1234567893, 1234567895, 3, "D?"},
		Task{5, 3, 1234567890, 1234567895, 5, "e$"},
	}

	var ts2 = Tasks{
		Task{1, 1, 1234567890, 1234567895, 3, "a"},
		Task{3, 2, 1234567892, 1234567895, 4, "b!"},
		Task{2, 2, 1234567891, 1234567895, 1, "C..."},
		Task{4, 1, 1234567893, 1234567895, 3, "D?"},
		Task{5, 3, 1234567890, 1234567895, 5, "e$"},
	}

	var ts3 = Tasks{
		Task{5, 3, 1234567890, 1234567895, 5, "e$"},
		Task{4, 1, 1234567893, 1234567895, 3, "D?"},
		Task{2, 2, 1234567891, 1234567895, 1, "C..."},
		Task{3, 2, 1234567892, 1234567895, 4, "b!"},
		Task{1, 1, 1234567890, 1234567895, 3, "a"},
	}

	ts1.SortByTask("ASC")
	for i, t1 := range ts1 {
		if t1 != ts2[i] {
			t.Errorf("SortByTask ASC is not as expected: [%v], instead of [%v]", ts1, ts2)
			return
		}
	}

	ts1.SortByTask("DESC")
	for i, t1 := range ts1 {
		if t1 != ts3[i] {
			t.Errorf("SortByTask DESC is not as expected: [%v], instead of [%v]", ts1, ts3)
			return
		}
	}
}
