package main

import "testing"
import "os"

func _storage_setup(t *testing.T) {
	SetDatabase("./data/tasks_test.db")
	SetupDatabase()

	a1 := Account{-1, "JamesClonk", "JamesClonk@developer", "abcd", "123", "Admin", 1234567890}
	if err := a1.Save(); err != nil {
		t.Fatal(err)
	}
	if a1.Id != 1 {
		t.Fatalf("Account ID after calling Save() is not correct. Got [%v], expected [%v]", a1.Id, 1)
	}
	a2 := Account{-1, "Clude", "clude@CLUDE", "abcd", "456", "User", 1234567891}
	if err := a2.Save(); err != nil {
		t.Fatal(err)
	}
	if a2.Id != 2 {
		t.Fatalf("Account ID after calling Save() is not correct. Got [%v], expected [%v]", a2.Id, 2)
	}
	a3 := Account{-1, "ozzie", "ozzie@abrakadabra", "abcd", "789", "User", 1234567892}
	if err := a3.Save(); err != nil {
		t.Fatal(err)
	}
	if a3.Id != 3 {
		t.Fatalf("Account ID after calling Save() is not correct. Got [%v], expected [%v]", a3.Id, 3)
	}
	a4 := Account{21, "Sonny", "sonny@sunny", "abcd", "999", "None", 1234567895}
	if err := a4.Save(); err != nil {
		t.Fatal(err)
	}
	if a4.Id != 21 {
		t.Fatalf("Account ID after calling Save() is not correct. Got [%v], expected [%v]", a4.Id, 21)
	}
	if err := a4.Delete(); err != nil {
		t.Fatal(err)
	}
	if a4.Id != -1 {
		t.Fatalf("Account ID after calling Delete() is not correct. Got [%v], expected [%v]", a4.Id, -1)
	}
	if err := a4.Save(); err != nil {
		t.Fatal(err)
	}
	if a4.Id != 4 {
		t.Fatalf("Account ID after calling Save() is not correct. Got [%v], expected [%v]", a4.Id, 4)
	}

	ts := Tasks{
		{-1, 1, 1234567890, 1234567895, 3, "Buy food!"},
		{-1, 2, 1234567891, 1234567895, 1, "Get some sleep..."},
		{-1, 2, 1234567892, 1234567895, 4, "Buy xmas presents!"},
		{-1, 1, 1234567893, 1234567895, 3, "Buy water!"},
		{-1, 3, 1234567890, 1234567895, 5, "ALARM!"},
	}
	if err := ts.Save(); err != nil {
		t.Fatal(err)
	}
	if ts[2].Id != 3 {
		t.Fatalf("#3 Task ID after calling Save() is not correct. Got [%v], expected [%v]", ts[2].Id, 3)
	}

	task := Task{-1, 1, 1234567890, 1234567895, 2, "Watch TV.."}
	if err := task.Save(); err != nil {
		t.Fatal(err)
	}
	if task.Id != 6 {
		t.Fatalf("Task ID after calling Save() is not correct. Got [%v], expected [%v]", task.Id, 6)
	}
}

func _storage_cleanup() {
	SetDatabase("./data/tasks_test.db")
	os.Remove(database)
}

func Test_storage_setup(t *testing.T) {
	_storage_setup(t)
}

func Test_storage_GetAllTasks(t *testing.T) {
	tasks, err := GetAllTasks()
	if err != nil {
		t.Error(err)
	}

	// GetAllTasks sorts by Priority by default
	expectedTasks := Tasks{
		{5, 3, 1234567890, 1234567895, 5, "ALARM!"},
		{3, 2, 1234567892, 1234567895, 4, "Buy xmas presents!"},
		{1, 1, 1234567890, 1234567895, 3, "Buy food!"},
		{4, 1, 1234567893, 1234567895, 3, "Buy water!"},
		{6, 1, 1234567890, 1234567895, 2, "Watch TV.."},
		{2, 2, 1234567891, 1234567895, 1, "Get some sleep..."},
	}
	for i, tk := range *tasks {
		if tk != expectedTasks[i] {
			t.Errorf("Tasks are not as expected: [%v], instead of [%v]", tasks, expectedTasks)
			return
		}
	}
}

func Test_storage_GetTaskById(t *testing.T) {
	task, err := GetTaskById(2)
	if err != nil {
		t.Error(err)
	}

	expectedTask := Task{2, 2, 1234567891, 1234567895, 1, "Get some sleep..."}
	if *task != expectedTask {
		t.Errorf("Task is not as expected: [%v], instead of [%v]", task, expectedTask)
		return
	}

	task, err = GetTaskById(17)
	if err == nil {
		t.Error("Expected sql error!")
	}
	if task != nil {
		t.Errorf("Task should be nil, instead of [%v]", task)
	}
}

func Test_storage_GetTasksByAccountId(t *testing.T) {
	tasks, err := GetTasksByAccountId(2)
	if err != nil {
		t.Error(err)
	}

	// GetTasksByAccountId sorts by Priority by default
	expectedTasks := Tasks{
		{3, 2, 1234567892, 1234567895, 4, "Buy xmas presents!"},
		{2, 2, 1234567891, 1234567895, 1, "Get some sleep..."},
	}
	for i, tk := range *tasks {
		if tk != expectedTasks[i] {
			t.Errorf("Tasks are not as expected: [%v], instead of [%v]", tasks, expectedTasks)
			return
		}
	}

	tasks, err = GetTasksByAccountId(7)
	if err != nil {
		t.Error(err)
	}
	if len(*tasks) != 0 {
		t.Errorf("Tasks should be empty, instead of [%v]", tasks)
	}
}

func Test_storage_DeleteTasks(t *testing.T) {
	ts := Tasks{
		{1, 1, 1234567890, 1234567895, 3, "Buy food!"},
		{3, 2, 1234567891, 1234567895, 1, "Get some sleep..."},
		{5, 2, 1234567892, 1234567895, 4, "Buy xmas presents!"},
	}
	if err := ts.Delete(); err != nil {
		t.Error(err)
	}
	if ts[1].Id != -1 {
		t.Errorf("#2 Task ID after calling Delete() is not correct. Got [%v], expected [%v]", ts[1].Id, -1)
	}

	ts2, err := GetAllTasks() // should have deleted 3 tasks
	if err != nil {
		t.Error(err)
	}
	if len(*ts2) != 3 {
		t.Errorf("Amount of Tasks in DB after calling Delete() is not correct. Got [%v], expected [%v]", len(*ts2), 3)
	}

	ts = Tasks{
		{13, 1, 1234567890, 1234567895, 3, "Buy food!"},
		{14, 2, 1234567891, 1234567895, 1, "Get some sleep..."},
	}
	if err := ts.Delete(); err != nil {
		t.Error(err)
	}
	ts2, err = GetAllTasks() // should not have deleted any tasks
	if err != nil {
		t.Error(err)
	}
	if len(*ts2) != 3 {
		t.Errorf("Amount of Tasks in DB after calling Delete() is not correct. Got [%v], expected [%v]", len(*ts2), 3)
	}

	task := Task{6, 1, 1234567890, 1234567895, 2, "Watch TV.."}
	if err := task.Delete(); err != nil {
		t.Error(err)
	}
	if task.Id != -1 {
		t.Errorf("Task ID after calling Delete() is not correct. Got [%v], expected [%v]", task.Id, -1)
	}

	ts2, err = GetAllTasks() // should have deleted 1 more task
	if err != nil {
		t.Error(err)
	}
	if len(*ts2) != 2 {
		t.Errorf("Amount of Tasks in DB after calling Delete() is not correct. Got [%v], expected [%v]", len(*ts2), 2)
	}
}

func Test_storage_GetAllAccounts(t *testing.T) {
	accounts, err := GetAllAccounts()
	if err != nil {
		t.Error(err)
	}

	// GetAllAccounts sorts by Id by default
	expectedAccounts := Accounts{
		Account{1, "JamesClonk", "JamesClonk@developer", "abcd", "123", "Admin", 1234567890},
		Account{2, "Clude", "clude@CLUDE", "abcd", "456", "User", 1234567891},
		Account{3, "ozzie", "ozzie@abrakadabra", "abcd", "789", "User", 1234567892},
		Account{4, "Sonny", "sonny@sunny", "abcd", "999", "None", 1234567895},
	}
	for i, a := range *accounts {
		if a != expectedAccounts[i] {
			t.Errorf("Accounts are not as expected: [%v], instead of [%v]", accounts, expectedAccounts)
			return
		}
	}
}

func Test_storage_GetAccount(t *testing.T) {
	account, err := GetAccountById(2)
	if err != nil {
		t.Error(err)
	}
	expectedAccount := Account{2, "Clude", "clude@CLUDE", "abcd", "456", "User", 1234567891}
	if *account != expectedAccount {
		t.Errorf("Account is not as expected: [%v], instead of [%v]", account, expectedAccount)
		return
	}

	account, err = GetAccountById(17)
	if err == nil {
		t.Error("Expected sql error!")
	}
	if account != nil {
		t.Errorf("Account should be nil, instead of [%v]", account)
	}

	account, err = GetAccountByEmail("JamesClonk@developer")
	if err != nil {
		t.Error(err)
	}
	expectedAccount = Account{1, "JamesClonk", "JamesClonk@developer", "abcd", "123", "Admin", 1234567890}
	if *account != expectedAccount {
		t.Errorf("Account is not as expected: [%v], instead of [%v]", account, expectedAccount)
		return
	}

	account, err = GetAccountByEmail("Whatever")
	if err == nil {
		t.Error("Expected sql error!")
	}
	if account != nil {
		t.Errorf("Account should be nil, instead of [%v]", account)
	}
}

func Test_storage_DeleteAccount(t *testing.T) {
	a := Account{3, "ozzie", "ozzie@abrakadabra", "abcd", "789", "User", 1234567892}
	if err := a.Delete(); err != nil {
		t.Error(err)
	}
	if a.Id != -1 {
		t.Errorf("Account ID after calling Delete() is not correct. Got [%v], expected [%v]", a.Id, -1)
	}

	as, err := GetAllAccounts()
	if err != nil {
		t.Error(err)
	}
	if len(*as) != 3 {
		t.Errorf("Amount of Accounts in DB after calling Delete() is not correct. Got [%v], expected [%v]", len(*as), 3)
	}

	a = Account{5, "Sonny", "Sonny@Sunny", "abcd", "999", "None", 1234567897}
	if err := a.Delete(); err != nil {
		t.Error(err)
	}
	if a.Id != -1 {
		t.Errorf("Account ID after calling Delete() is not correct. Got [%v], expected [%v]", a.Id, -1)
	}
}

func Test_storage_cleanup(t *testing.T) {
	_storage_cleanup()
}
