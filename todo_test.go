package main

import "testing"
import "strconv"
import "strings"
import "net/url"
import "net/http"
import "net/http/httptest"
import "encoding/json"

func Test_todo_setup(t *testing.T) {
	isLogging = false
	_storage_setup(t)
}

func Test_todo_parseCommandline(t *testing.T) {
	t.Fail()
}

func Test_todo_authHandler(t *testing.T) {
	_, id, timestamp, salt, token, err := _authSetup(t, 1) // generate request token for AccountId 1
	if err != nil {
		t.Error(err)
		return
	}

	testfn := func(w http.ResponseWriter, r *http.Request, accountId int) {
		if strconv.Itoa(accountId) != *id {
			w.Write([]byte("Failure!"))
		} else {
			w.Write([]byte("Success!"))
		}
	}
	mh := MethodHandler{"GET": testfn}

	// ============================================ Valid ============================================
	request, err := http.NewRequest("GET", "http://localhost:8008/does not matter here/?rId="+*id+"&rTimestamp="+*timestamp+"&rSalt="+*salt+"&rToken="+*token, nil)
	if err != nil {
		t.Error(err)
		return
	}
	response := httptest.NewRecorder()
	authHandler(mh)(response, request)
	_checkResponseCode(t, response, 200)
	_checkResponseBody(t, response, "Success!")

	// ============================================ Invalid ============================================
	request, err = http.NewRequest("GET", "http://localhost:8008/does not matter here/?rId=2&rTimestamp="+*timestamp+"&rSalt="+*salt+"&rToken="+*token, nil) // AccountId 2 has different Password
	if err != nil {
		t.Error(err)
		return
	}
	response = httptest.NewRecorder()
	authHandler(mh)(response, request)
	_checkResponseCode(t, response, 200)
	_checkResponseBody(t, response, "Failure!")

	// ============================================ Unauthorized ============================================
	request, err = http.NewRequest("GET", "http://localhost:8008/does not matter here/?rId="+*id+"&rTimestamp="+*timestamp+"&rSalt="+*salt+"&rToken=", nil) // Invalid Token
	if err != nil {
		t.Error(err)
		return
	}
	response = httptest.NewRecorder()
	authHandler(mh)(response, request)
	_checkResponseCode(t, response, 401)
	_checkResponseBody(t, response, "Unauthorized")

	// ============================================ Nonexisting AccountId ============================================
	request, err = http.NewRequest("GET", "http://localhost:8008/does not matter here/?rId=17&rTimestamp="+*timestamp+"&rSalt="+*salt+"&rToken="+*token, nil)
	if err != nil {
		t.Error(err)
		return
	}
	response = httptest.NewRecorder()
	authHandler(mh)(response, request)
	_checkResponseCode(t, response, 401)
	_checkResponseBody(t, response, "Unauthorized")
}

func Test_todo_getId(t *testing.T) {
	// ============================================ Valid ============================================
	request, err := http.NewRequest("GET", "http://localhost:8008/task/7", nil)
	if err != nil {
		t.Error(err)
		return
	}
	response := httptest.NewRecorder()

	getId(response, request)
	_checkResponseCode(t, response, 200)

	body := response.Body.String()
	if body != "" {
		t.Errorf("getId() response body was [%v], but expected it to be empty", body)
	}

	// ============================================ Invalid ============================================
	request, err = http.NewRequest("GET", "http://localhost:8008/tasks/7", nil)
	if err != nil {
		t.Error(err)
		return
	}
	response = httptest.NewRecorder()

	getId(response, request)
	_checkResponseCode(t, response, 404)
	_checkResponseBody(t, response, "404 page not found")
}

func Test_todo_getAuth(t *testing.T) {
	// ============================================ Valid ============================================
	request, err := http.NewRequest("GET", "http://localhost:8008/auth/?login=JamesClonk@developer", nil)
	if err != nil {
		t.Error(err)
		return
	}
	response := httptest.NewRecorder()

	getAuth(response, request)
	_checkResponseCode(t, response, 200)
	_checkResponseBody(t, response, `{"AccountId": 1, "Salt": "123", "Timestamp":`)

	account, err := GetAccountById(1)
	if err != nil {
		t.Error(err)
		return
	}
	beforeLastauthUpdate := Account{1, "JamesClonk", "JamesClonk@developer", "abcd", "123", "Admin", 1234567890}
	if *account == beforeLastauthUpdate {
		t.Errorf("getAuth() Account.LastAuth should not be the same anymore: [%v] vs. [%v]", *account, beforeLastauthUpdate)
	}

	// ============================================ Not Found ============================================
	request, err = http.NewRequest("GET", "http://localhost:8008/auth/?login=Clude@developer", nil) // Email does not exist in DB
	if err != nil {
		t.Error(err)
		return
	}
	response = httptest.NewRecorder()

	getAuth(response, request)
	_checkResponseCode(t, response, 404)
	_checkResponseBody(t, response, "404 page not found")

	// ============================================ Disabled Account ============================================
	request, err = http.NewRequest("GET", "http://localhost:8008/auth/?login=sonny@sunny", nil) // Role is set to "None"
	if err != nil {
		t.Error(err)
		return
	}
	response = httptest.NewRecorder()

	getAuth(response, request)
	_checkResponseCode(t, response, 401)
	_checkResponseBody(t, response, "Unauthorized")
}

func _todo_getTasks(t *testing.T, id int, expectedTasks Tasks) {
	request, err := http.NewRequest("GET", "http://localhost:8008/tasks/", nil)
	if err != nil {
		t.Error(err)
		return
	}
	response := httptest.NewRecorder()

	getTasks(response, request, id)
	_checkResponseCode(t, response, 200)

	body := response.Body.String()
	var tasks Tasks
	if err := json.Unmarshal([]byte(body), &tasks); err != nil {
		t.Error(err)
		return
	}
	for i, tk := range tasks {
		if tk != expectedTasks[i] {
			t.Errorf("getTasks() are not as expected: [%v], instead of [%v]", tasks, expectedTasks)
			return
		}
	}
}

func Test_todo_getTasks(t *testing.T) {
	// should be sorted by Priority by default, and only return users tasks.
	expectedTasks := Tasks{
		{1, 1, 1234567890, 1234567895, 3, "Buy food!"},
		{4, 1, 1234567893, 1234567895, 3, "Buy water!"},
		{6, 1, 1234567890, 1234567895, 2, "Watch TV.."},
	}
	_todo_getTasks(t, 1, expectedTasks)

	expectedTasks = Tasks{
		{3, 2, 1234567892, 1234567895, 4, "Buy xmas presents!"},
		{2, 2, 1234567891, 1234567895, 1, "Get some sleep..."},
	}
	_todo_getTasks(t, 2, expectedTasks)

	expectedTasks = Tasks{}
	_todo_getTasks(t, 4, expectedTasks) // AccountId 4 does not exists in DB, 0 tasks expected to be returned
}

func Test_todo_getTask(t *testing.T) {
	// ============================================ Valid ============================================
	request, err := http.NewRequest("GET", "http://localhost:8008/task/1", nil) // task belongs to AccountId 1
	if err != nil {
		t.Error(err)
		return
	}
	response := httptest.NewRecorder()

	getTask(response, request, 1) // Use AccountId 1
	_checkResponseCode(t, response, 200)

	body := response.Body.String()
	expected := Task{1, 1, 1234567890, 1234567895, 3, "Buy food!"}
	var task Task
	if err := json.Unmarshal([]byte(body), &task); err != nil {
		t.Error(err)
	}
	if task != expected {
		t.Errorf("getTask() response json was [%v], but expected [%v]", task, expected)
	}

	// ============================================ Invalid ============================================
	request, err = http.NewRequest("GET", "http://localhost:8008/task/öäü", nil)
	if err != nil {
		t.Error(err)
		return
	}
	response = httptest.NewRecorder()

	getTask(response, request, 1)
	_checkResponseCode(t, response, 404)
	_checkResponseBody(t, response, "404 page not found")

	// ============================================ Unauthorized ============================================
	request, err = http.NewRequest("GET", "http://localhost:8008/task/1", nil) // task belongs to AccountId 1
	if err != nil {
		t.Error(err)
		return
	}
	response = httptest.NewRecorder()

	getTask(response, request, 2) // Use AccountId 2, which does not have Admin role
	_checkResponseCode(t, response, 401)
	_checkResponseBody(t, response, "Unauthorized")

	// ============================================ Valid Admin ============================================
	request, err = http.NewRequest("GET", "http://localhost:8008/task/2", nil) // task belongs to AccountId 2
	if err != nil {
		t.Error(err)
		return
	}
	response = httptest.NewRecorder()

	getTask(response, request, 1) // Use AccountId 1, which has Admin role
	_checkResponseCode(t, response, 200)

	body = response.Body.String()
	expected = Task{2, 2, 1234567891, 1234567895, 1, "Get some sleep..."}
	if err := json.Unmarshal([]byte(body), &task); err != nil {
		t.Error(err)
	}
	if task != expected {
		t.Errorf("getTask() response json was [%v], but expected [%v]", task, expected)
	}

	// ============================================ Nonexisting Task ============================================
	request, err = http.NewRequest("GET", "http://localhost:8008/task/77", nil)
	if err != nil {
		t.Error(err)
		return
	}
	response = httptest.NewRecorder()

	getTask(response, request, 1)
	_checkResponseCode(t, response, 404)
	_checkResponseBody(t, response, "404 page not found")

	// ============================================ Nonexisting Account ============================================
	request, err = http.NewRequest("GET", "http://localhost:8008/task/1", nil)
	if err != nil {
		t.Error(err)
		return
	}
	response = httptest.NewRecorder()

	getTask(response, request, 77)
	_checkResponseCode(t, response, 401)
	_checkResponseBody(t, response, "Unauthorized")
}

func Test_todo_addTask(t *testing.T) {
	t.Fail() // need to check automatically set Created and LastUpdated timestamps.

	// ============================================ Valid ============================================
	request, err := http.NewRequest("POST", "http://localhost:8008/task/", nil)
	if err != nil {
		t.Error(err)
		return
	}
	newTask := Task{7, 2, 1234567890, 1234567899, 1, "Get some more sleep!"} // task will belong to AccountId 2
	request.Form = url.Values{
		"Id":          {"-1"},
		"AccountId":   {"2"},
		"Created":     {"1234567890"},
		"LastUpdated": {"1234567899"},
		"Priority":    {"1"},
		"Task":        {"Get some more sleep!"},
	}

	response := httptest.NewRecorder()

	addTask(response, request, 2) // Use AccountId 2
	_checkResponseCode(t, response, 200)
	_checkResponseBody(t, response, "{\"Add\": \"Success\"}")

	task, err := GetTaskById(7)
	if err != nil {
		t.Error(err)
		return
	}
	if *task != newTask {
		t.Errorf("GetTaskById() after addTask() returned [%v], but expected task [%v]", task, newTask)
	}

	// ============================================ Invalid ============================================
	request, err = http.NewRequest("POST", "http://localhost:8008/task/öäü", nil)
	if err != nil {
		t.Error(err)
		return
	}
	response = httptest.NewRecorder()

	addTask(response, request, 1)
	_checkResponseCode(t, response, 400)
	_checkResponseBody(t, response, "Invalid data")

	// ============================================ Unauthorized ============================================
	request, err = http.NewRequest("POST", "http://localhost:8008/task/", nil)
	if err != nil {
		t.Error(err)
		return
	}
	request.Form = url.Values{
		"Id":          {"-1"},
		"AccountId":   {"1"}, // task would belong to AccountId 1
		"Created":     {"1234567890"},
		"LastUpdated": {"1234567899"},
		"Priority":    {"1"},
		"Task":        {"Get some more sleep!"},
	}

	response = httptest.NewRecorder()

	addTask(response, request, 2) // Use AccountId 2
	_checkResponseCode(t, response, 401)
	_checkResponseBody(t, response, "Unauthorized")

	task, err = GetTaskById(8)
	if err == nil || task != nil {
		t.Errorf("GetTaskById() after unauthorized addTask() returned [%v], but expected nil", task)
	}

	// ============================================ Valid Admin ============================================
	request, err = http.NewRequest("POST", "http://localhost:8008/task/", nil)
	if err != nil {
		t.Error(err)
		return
	}
	newTask = Task{8, 3, 1234567800, 1234567809, 3, "Get some more sleep!!!"} // task would belong to AccountId 3
	request.Form = url.Values{
		"Id":          {"-1"},
		"AccountId":   {"3"}, // task would belong to AccountId 3
		"Created":     {"1234567800"},
		"LastUpdated": {"1234567809"},
		"Priority":    {"3"},
		"Task":        {"Get some more sleep!!!"},
	}

	response = httptest.NewRecorder()

	addTask(response, request, 1) // Use AccountId 1, which has Admin role
	_checkResponseCode(t, response, 200)
	_checkResponseBody(t, response, "{\"Add\": \"Success\"}")

	task, err = GetTaskById(8)
	if err != nil {
		t.Error(err)
		return
	}
	if *task != newTask {
		t.Errorf("GetTaskById() after addTask() returned [%v], but expected task [%v]", task, newTask)
	}

	// ============================================ Nonexisting Account ============================================
	request, err = http.NewRequest("POST", "http://localhost:8008/task/1", nil)
	if err != nil {
		t.Error(err)
		return
	}
	request.Form = url.Values{
		"Id":          {"-1"},
		"AccountId":   {"3"}, // task would belong to AccountId 3
		"Created":     {"1234567800"},
		"LastUpdated": {"1234567809"},
		"Priority":    {"3"},
		"Task":        {"Get some more sleep!!!"},
	}
	response = httptest.NewRecorder()

	addTask(response, request, 77)
	_checkResponseCode(t, response, 401)
	_checkResponseBody(t, response, "Unauthorized")
}

func Test_todo_editTask(t *testing.T) {
	t.Fail() // need to check against not being allowed to modify Created and LastUpdated timestamps.
	// LastUpdated should be set automatically!

	// ============================================ Valid ============================================
	request, err := http.NewRequest("PUT", "http://localhost:8008/task/6", nil)
	if err != nil {
		t.Error(err)
		return
	}
	editedTask := Task{6, 1, 12345678977, 12345678977, 7, "Watch TV.. !!!!!!"}
	request.Form = url.Values{
		"Id":          {"6"},
		"AccountId":   {"1"},
		"Created":     {"12345678977"},
		"LastUpdated": {"12345678977"},
		"Priority":    {"7"},
		"Task":        {"Watch TV.. !!!!!!"},
	}

	response := httptest.NewRecorder()

	editTask(response, request, 1)
	_checkResponseCode(t, response, 200)
	_checkResponseBody(t, response, "{\"Edit\": \"Success\"}")

	task, err := GetTaskById(6)
	if err != nil {
		t.Error(err)
		return
	}
	if *task != editedTask {
		t.Errorf("GetTaskById() after editTask() returned [%v], but expected task [%v]", task, editedTask)
	}

	// ============================================ Invalid URL ============================================
	request, err = http.NewRequest("PUT", "http://localhost:8008/task/öäü", nil)
	if err != nil {
		t.Error(err)
		return
	}
	response = httptest.NewRecorder()

	editTask(response, request, 1)
	_checkResponseCode(t, response, 404)
	_checkResponseBody(t, response, "404 page not found")

	// ============================================ Invalid Task ============================================
	request, err = http.NewRequest("PUT", "http://localhost:8008/task/77", nil)
	if err != nil {
		t.Error(err)
		return
	}
	response = httptest.NewRecorder()

	editTask(response, request, 1)
	_checkResponseCode(t, response, 400)
	_checkResponseBody(t, response, "Invalid data")

	// ============================================ Unauthorized ============================================
	request, err = http.NewRequest("PUT", "http://localhost:8008/task/1", nil) // task 1 belongs to AccountId 1
	if err != nil {
		t.Error(err)
		return
	}
	request.Form = url.Values{
		"Id":          {"1"},
		"AccountId":   {"1"},
		"Created":     {"1234567897"},
		"LastUpdated": {"1234567897"},
		"Priority":    {"1"},
		"Task":        {"Test!"},
	}

	response = httptest.NewRecorder()

	editTask(response, request, 2) // Use AccountId 2, which does not have Admin role
	_checkResponseCode(t, response, 401)
	_checkResponseBody(t, response, "Unauthorized")

	editedTask = Task{1, 1, 1234567890, 1234567895, 3, "Buy food!"}
	task, err = GetTaskById(1)
	if err != nil {
		t.Error(err)
		return
	}
	if *task != editedTask {
		t.Errorf("GetTaskById() after editTask() returned [%v], but expected task [%v]", task, editedTask)
	}

	// ============================================ Valid Admin ============================================
	request, err = http.NewRequest("PUT", "http://localhost:8008/task/5", nil) // task 5 belongs to AccountId 3
	if err != nil {
		t.Error(err)
		return
	}
	request.Form = url.Values{
		"Id":          {"5"},
		"AccountId":   {"1"}, // Change AccountId / Task Owner
		"Created":     {"1234567897"},
		"LastUpdated": {"1234567897"},
		"Priority":    {"1"},
		"Task":        {"Test!"},
	}

	response = httptest.NewRecorder()

	editTask(response, request, 1) // Use AccountId 1, which has Admin role
	_checkResponseCode(t, response, 200)
	_checkResponseBody(t, response, "{\"Edit\": \"Success\"}")

	editedTask = Task{5, 1, 1234567897, 1234567897, 1, "Test!"}
	task, err = GetTaskById(5)
	if err != nil {
		t.Error(err)
		return
	}
	if *task != editedTask {
		t.Errorf("GetTaskById() after editTask() returned [%v], but expected task [%v]", task, editedTask)
	}

	// ============================================ Nonmatching Id ============================================
	request, err = http.NewRequest("PUT", "http://localhost:8008/task/1", nil) // task 1 belongs to AccountId 1
	if err != nil {
		t.Error(err)
		return
	}
	request.Form = url.Values{
		"Id":          {"2"}, // Id does not match Id used in URL
		"AccountId":   {"1"},
		"Created":     {"1234567897"},
		"LastUpdated": {"1234567897"},
		"Priority":    {"1"},
		"Task":        {"Test!"},
	}

	response = httptest.NewRecorder()

	editTask(response, request, 1) // Use AccountId 1
	_checkResponseCode(t, response, 409)
	_checkResponseBody(t, response, "URL Id and Form Id do not match")

	editedTask = Task{1, 1, 1234567890, 1234567895, 3, "Buy food!"}
	task, err = GetTaskById(1)
	if err != nil {
		t.Error(err)
		return
	}
	if *task != editedTask {
		t.Errorf("GetTaskById() after editTask() returned [%v], but expected task [%v]", task, editedTask)
	}

	editedTask = Task{2, 2, 1234567891, 1234567895, 1, "Get some sleep..."}
	task, err = GetTaskById(2)
	if err != nil {
		t.Error(err)
		return
	}
	if *task != editedTask {
		t.Errorf("GetTaskById() after editTask() returned [%v], but expected task [%v]", task, editedTask)
	}

	// ============================================ Valid Nonexisting Task ============================================
	request, err = http.NewRequest("PUT", "http://localhost:8008/task/10", nil) // task 10 does not yet exist
	if err != nil {
		t.Error(err)
		return
	}
	request.Form = url.Values{
		"Id":          {"10"}, // task 10 does not yet exist
		"AccountId":   {"3"},
		"Created":     {"1234567897"},
		"LastUpdated": {"1234567897"},
		"Priority":    {"1"},
		"Task":        {"Test!"},
	}

	response = httptest.NewRecorder()

	editTask(response, request, 3) // Use AccountId 3
	_checkResponseCode(t, response, 200)
	_checkResponseBody(t, response, "{\"Edit\": \"Success\"}")

	editedTask = Task{10, 3, 1234567897, 1234567897, 1, "Test!"}
	task, err = GetTaskById(10)
	if err != nil {
		t.Error(err)
		return
	}
	if *task != editedTask {
		t.Errorf("GetTaskById() after editTask() returned [%v], but expected task [%v]", task, editedTask)
	}

	// ============================================ Unauthorized Nonexisting Task ============================================
	request, err = http.NewRequest("PUT", "http://localhost:8008/task/11", nil) // task 11 does not yet exist
	if err != nil {
		t.Error(err)
		return
	}
	request.Form = url.Values{
		"Id":          {"11"}, // task 11 does not yet exist
		"AccountId":   {"2"},
		"Created":     {"1234567897"},
		"LastUpdated": {"1234567897"},
		"Priority":    {"1"},
		"Task":        {"Test!!"},
	}

	response = httptest.NewRecorder()

	editTask(response, request, 3) // Use AccountId 3
	_checkResponseCode(t, response, 401)
	_checkResponseBody(t, response, "Unauthorized")

	task, err = GetTaskById(11)
	if err == nil || task != nil {
		t.Errorf("GetTaskById() after unauthorized editTask() returned [%v], but expected nil", task)
	}

	// ============================================ Valid Admin Nonexisting Task ============================================
	request, err = http.NewRequest("PUT", "http://localhost:8008/task/12", nil) // task 12 does not yet exist
	if err != nil {
		t.Error(err)
		return
	}
	request.Form = url.Values{
		"Id":          {"12"}, // task 12 does not yet exist
		"AccountId":   {"2"},
		"Created":     {"1234567897"},
		"LastUpdated": {"1234567897"},
		"Priority":    {"1"},
		"Task":        {"Test!!!"},
	}

	response = httptest.NewRecorder()

	editTask(response, request, 1) // Use AccountId 1
	_checkResponseCode(t, response, 200)
	_checkResponseBody(t, response, "{\"Edit\": \"Success\"}")

	editedTask = Task{12, 2, 1234567897, 1234567897, 1, "Test!!!"}
	task, err = GetTaskById(12)
	if err != nil {
		t.Error(err)
		return
	}
	if *task != editedTask {
		t.Errorf("GetTaskById() after editTask() returned [%v], but expected task [%v]", task, editedTask)
	}

	// ============================================ Nonexisting Account ============================================
	request, err = http.NewRequest("PUT", "http://localhost:8008/task/17", nil)
	if err != nil {
		t.Error(err)
		return
	}
	request.Form = url.Values{
		"Id":        {"17"}, // task 17 does not yet exist
		"AccountId": {"55"},
	}
	response = httptest.NewRecorder()

	editTask(response, request, 88)
	_checkResponseCode(t, response, 401)
	_checkResponseBody(t, response, "Unauthorized")
}

func Test_todo_deleteTask(t *testing.T) {
	// ============================================ Valid ============================================
	request, err := http.NewRequest("DELETE", "http://localhost:8008/task/3", nil) // task belongs to AccountId 2
	if err != nil {
		t.Error(err)
		return
	}
	response := httptest.NewRecorder()

	deleteTask(response, request, 2) // Use AccountId 2
	_checkResponseCode(t, response, 200)
	_checkResponseBody(t, response, "{\"Delete\": \"Success\"}")

	task, err := GetTaskById(3)
	if task != nil {
		t.Errorf("GetTaskById() after deleteTask() still returned a task object, when it should not! Got [%v]", task)
		if err != nil {
			t.Error(err)
		}
	}

	// ============================================ Invalid ============================================
	request, err = http.NewRequest("DELETE", "http://localhost:8008/task/öäü", nil)
	if err != nil {
		t.Error(err)
		return
	}
	response = httptest.NewRecorder()

	deleteTask(response, request, 1)
	_checkResponseCode(t, response, 404)
	_checkResponseBody(t, response, "404 page not found")

	// ============================================ Unauthorized ============================================
	request, err = http.NewRequest("DELETE", "http://localhost:8008/task/4", nil) // task belongs to AccountId 1
	if err != nil {
		t.Error(err)
		return
	}
	response = httptest.NewRecorder()

	deleteTask(response, request, 2) // Use AccountId 2, which does not have Admin role
	_checkResponseCode(t, response, 401)
	_checkResponseBody(t, response, "Unauthorized")

	task, err = GetTaskById(4)
	if err != nil {
		t.Error(err)
	}
	if task == nil {
		t.Error("GetTaskById() after deleteTask() did not return a task object, but it still should!")
	}

	// ============================================ Valid Admin ============================================
	request, err = http.NewRequest("DELETE", "http://localhost:8008/task/2", nil) // task belongs to AccountId 2
	if err != nil {
		t.Error(err)
		return
	}
	response = httptest.NewRecorder()

	deleteTask(response, request, 1) // Use AccountId 1, which has Admin role
	_checkResponseCode(t, response, 200)
	_checkResponseBody(t, response, "{\"Delete\": \"Success\"}")

	task, err = GetTaskById(2)
	if task != nil {
		t.Errorf("GetTaskById() after deleteTask() still returned a task object, when it should not! Got [%v]", task)
		if err != nil {
			t.Error(err)
		}
	}

	// ============================================ Nonexisting Task ============================================
	request, err = http.NewRequest("DELETE", "http://localhost:8008/task/77", nil)
	if err != nil {
		t.Error(err)
		return
	}
	response = httptest.NewRecorder()

	deleteTask(response, request, 1)
	_checkResponseCode(t, response, 404)
	_checkResponseBody(t, response, "404 page not found")

	// ============================================ Nonexisting Account ============================================
	request, err = http.NewRequest("DELETE", "http://localhost:8008/task/1", nil)
	if err != nil {
		t.Error(err)
		return
	}
	response = httptest.NewRecorder()

	deleteTask(response, request, 77)
	_checkResponseCode(t, response, 401)
	_checkResponseBody(t, response, "Unauthorized")
}

func Test_todo_getAccounts(t *testing.T) {
	// ============================================ Valid ============================================
	request, err := http.NewRequest("GET", "http://localhost:8008/accounts/", nil)
	if err != nil {
		t.Error(err)
		return
	}
	response := httptest.NewRecorder()

	expectedAccounts := Accounts{
		{1, "JamesClonk", "JamesClonk@developer", "abcd", "123", "Admin", 1234567890},
		{2, "Clude", "clude@CLUDE", "abcd", "456", "User", 1234567891},
		{3, "ozzie", "ozzie@abrakadabra", "abcd", "789", "User", 1234567892},
		{4, "Sonny", "sonny@sunny", "abcd", "999", "None", 1234567895},
	}

	getAccounts(response, request, 1) // Use AccountId 1, which has Admin role
	_checkResponseCode(t, response, 200)

	body := response.Body.String()
	var accounts Accounts
	if err := json.Unmarshal([]byte(body), &accounts); err != nil {
		t.Error(err)
		return
	}
	for i, a := range accounts {
		if a.Email != expectedAccounts[i].Email || a.Name != expectedAccounts[i].Name {
			t.Errorf("getAccounts() are not as expected: [%v], instead of [%v]", accounts, expectedAccounts)
			return
		}
	}

	// ============================================ Unauthorized ============================================
	request, err = http.NewRequest("GET", "http://localhost:8008/accounts/", nil)
	if err != nil {
		t.Error(err)
		return
	}
	response = httptest.NewRecorder()

	getAccounts(response, request, 2) // Use AccountId 2, which does not have Admin role
	_checkResponseCode(t, response, 401)
	_checkResponseBody(t, response, "Unauthorized")

	// ============================================ Unauthorized ============================================
	request, err = http.NewRequest("GET", "http://localhost:8008/accounts/", nil)
	if err != nil {
		t.Error(err)
		return
	}
	response = httptest.NewRecorder()

	getAccounts(response, request, 99)
	_checkResponseCode(t, response, 401)
	_checkResponseBody(t, response, "Unauthorized")
}

func Test_todo_getAccount(t *testing.T) {
	// ============================================ Valid ============================================
	request, err := http.NewRequest("GET", "http://localhost:8008/account/2", nil)
	if err != nil {
		t.Error(err)
		return
	}
	response := httptest.NewRecorder()

	getAccount(response, request, 2) // Use AccountId 2
	_checkResponseCode(t, response, 200)

	body := response.Body.String()
	expected := Account{2, "Clude", "clude@CLUDE", "", "456", "User", 1234567891}
	var account Account
	if err := json.Unmarshal([]byte(body), &account); err != nil {
		t.Error(err)
	}
	if account != expected {
		t.Errorf("getAccount() response json was [%v], but expected [%v]", account, expected)
	}

	// ============================================ Invalid ============================================
	request, err = http.NewRequest("GET", "http://localhost:8008/account/öäü", nil)
	if err != nil {
		t.Error(err)
		return
	}
	response = httptest.NewRecorder()

	getAccount(response, request, 1)
	_checkResponseCode(t, response, 404)
	_checkResponseBody(t, response, "404 page not found")

	// ============================================ Unauthorized ============================================
	request, err = http.NewRequest("GET", "http://localhost:8008/account/1", nil)
	if err != nil {
		t.Error(err)
		return
	}
	response = httptest.NewRecorder()

	getAccount(response, request, 2) // Use AccountId 2, which does not have Admin role
	_checkResponseCode(t, response, 401)
	_checkResponseBody(t, response, "Unauthorized")

	// ============================================ Valid Admin ============================================
	request, err = http.NewRequest("GET", "http://localhost:8008/account/2", nil)
	if err != nil {
		t.Error(err)
		return
	}
	response = httptest.NewRecorder()

	getAccount(response, request, 1) // Use AccountId 1, which has Admin role
	_checkResponseCode(t, response, 200)

	body = response.Body.String()
	expected = Account{2, "Clude", "clude@CLUDE", "", "456", "User", 1234567891}
	if err := json.Unmarshal([]byte(body), &account); err != nil {
		t.Error(err)
	}
	if account != expected {
		t.Errorf("getAccount() response json was [%v], but expected [%v]", account, expected)
	}

	// ============================================ Nonexisting Account ============================================
	request, err = http.NewRequest("GET", "http://localhost:8008/account/1", nil)
	if err != nil {
		t.Error(err)
		return
	}
	response = httptest.NewRecorder()

	getAccount(response, request, 7)
	_checkResponseCode(t, response, 401)
	_checkResponseBody(t, response, "Unauthorized")

	// ============================================ Nonexisting Account ============================================
	request, err = http.NewRequest("GET", "http://localhost:8008/account/77", nil)
	if err != nil {
		t.Error(err)
		return
	}
	response = httptest.NewRecorder()

	getAccount(response, request, 1)
	_checkResponseCode(t, response, 404)
	_checkResponseBody(t, response, "404 page not found")
}

func Test_todo_addAccount(t *testing.T) {
	// ============================================ Valid ============================================
	request, err := http.NewRequest("POST", "http://localhost:8008/account/", nil)
	if err != nil {
		t.Error(err)
		return
	}
	newAccount := Account{5, "Samurai", "Samurai@Ronin", "abcdef", "123456789", "User", 0}
	request.Form = url.Values{
		"Id":       {"23"}, // ignored, does not matter
		"Name":     {"Samurai"},
		"Email":    {"Samurai@Ronin"},
		"Password": {"abcdef"},
		"Salt":     {"123456789"},
		"Role":     {"User"},
		"LastAuth": {"123456789"}, // cannot be set
	}

	response := httptest.NewRecorder()

	addAccount(response, request, 1) // Use AccountId 1, which has Admin role
	_checkResponseCode(t, response, 200)
	_checkResponseBody(t, response, "{\"Add\": \"Success\"}")

	account, err := GetAccountById(5) // should be the 5th account by now
	if err != nil {
		t.Error(err)
		return
	}
	if *account != newAccount {
		t.Errorf("GetAccountById() after addAccount() returned [%v], but expected account [%v]", account, newAccount)
	}

	// ============================================ Invalid ============================================
	request, err = http.NewRequest("POST", "http://localhost:8008/account/öäü", nil)
	if err != nil {
		t.Error(err)
		return
	}
	response = httptest.NewRecorder()

	addAccount(response, request, 1)
	_checkResponseCode(t, response, 400)
	_checkResponseBody(t, response, "Invalid data")

	// ============================================ Unauthorized ============================================
	request, err = http.NewRequest("POST", "http://localhost:8008/account/", nil)
	if err != nil {
		t.Error(err)
		return
	}
	request.Form = url.Values{
		"Id":       {"42"},
		"Name":     {"Samurai 2"},
		"Email":    {"Samurai@Ronin"},
		"Password": {"abcdefgh"},
		"Salt":     {"1234567890"},
		"Role":     {"User"},
		"LastAuth": {"1234567890"},
	}

	response = httptest.NewRecorder()

	addAccount(response, request, 2) // Use AccountId 2, which does not have Admin role
	_checkResponseCode(t, response, 401)
	_checkResponseBody(t, response, "Unauthorized")

	account, err = GetAccountById(6)
	if err == nil || account != nil {
		t.Errorf("GetAccountById() after unauthorized addAccount() returned [%v], but expected nil", account)
	}
}

func Test_todo_editAccount(t *testing.T) {
	// ============================================ Valid ============================================
	request, err := http.NewRequest("PUT", "http://localhost:8008/account/2", nil)
	if err != nil {
		t.Error(err)
		return
	}
	editedAccount := Account{2, "Cluderzky", "clude@CLUDE", "abcd", "123456", "User", 1234567891}
	request.Form = url.Values{
		"Id":       {"2"},
		"Name":     {"Cluderzky"},
		"Email":    {"clude@CLUDE"},
		"Password": {"abcd"},
		"Salt":     {"123456"},
		"Role":     {"User"},
		"LastAuth": {"111111"}, // cannot be changed
	}

	response := httptest.NewRecorder()

	editAccount(response, request, 2)
	_checkResponseCode(t, response, 200)
	_checkResponseBody(t, response, "{\"Edit\": \"Success\"}")

	account, err := GetAccountById(2)
	if err != nil {
		t.Error(err)
		return
	}
	if *account != editedAccount {
		t.Errorf("GetAccountById() after editAccount() returned [%v], but expected account [%v]", account, editedAccount)
	}

	// ============================================ Role Change NonAdmin ============================================
	request, err = http.NewRequest("PUT", "http://localhost:8008/account/3", nil)
	if err != nil {
		t.Error(err)
		return
	}
	editedAccount = Account{3, "ozzie123", "ozzie@abrakadabra123", "abcd", "789", "User", 1234567892}
	request.Form = url.Values{
		"Id":       {"3"},
		"Name":     {"ozzie123"},
		"Email":    {"ozzie@abrakadabra123"},
		"Password": {"abcd"},
		"Salt":     {"789"},
		"Role":     {"Admin"},
	}

	response = httptest.NewRecorder()

	editAccount(response, request, 3)
	_checkResponseCode(t, response, 200)
	_checkResponseBody(t, response, "{\"Edit\": \"Success\"}")

	account, err = GetAccountById(3)
	if err != nil {
		t.Error(err)
		return
	}
	if *account != editedAccount {
		t.Errorf("GetAccountById() after editAccount() returned [%v], but expected account [%v]", account, editedAccount)
	}

	// ============================================ Role Change Admin ============================================
	request, err = http.NewRequest("PUT", "http://localhost:8008/account/3", nil)
	if err != nil {
		t.Error(err)
		return
	}
	editedAccount = Account{3, "ozzie", "ozzie@abrakadabra", "abcd", "789", "None", 1234567892}
	request.Form = url.Values{
		"Id":       {"3"},
		"Name":     {"ozzie"},
		"Email":    {"ozzie@abrakadabra"},
		"Password": {"abcd"},
		"Salt":     {"789"},
		"Role":     {"None"},
	}

	response = httptest.NewRecorder()

	editAccount(response, request, 1)
	_checkResponseCode(t, response, 200)
	_checkResponseBody(t, response, "{\"Edit\": \"Success\"}")

	account, err = GetAccountById(3)
	if err != nil {
		t.Error(err)
		return
	}
	if *account != editedAccount {
		t.Errorf("GetAccountById() after editAccount() returned [%v], but expected account [%v]", account, editedAccount)
	}

	// ============================================ Invalid URL ============================================
	request, err = http.NewRequest("PUT", "http://localhost:8008/account/öäü", nil)
	if err != nil {
		t.Error(err)
		return
	}
	response = httptest.NewRecorder()

	editAccount(response, request, 1)
	_checkResponseCode(t, response, 404)
	_checkResponseBody(t, response, "404 page not found")

	// ============================================ Invalid Account ============================================
	request, err = http.NewRequest("PUT", "http://localhost:8008/account/77", nil)
	if err != nil {
		t.Error(err)
		return
	}
	response = httptest.NewRecorder()

	editAccount(response, request, 1)
	_checkResponseCode(t, response, 400)
	_checkResponseBody(t, response, "Invalid data")

	// ============================================ Unauthorized ============================================
	request, err = http.NewRequest("PUT", "http://localhost:8008/account/3", nil)
	if err != nil {
		t.Error(err)
		return
	}
	request.Form = url.Values{
		"Id":       {"3"},
		"Name":     {"ozzy"},
		"Email":    {"ozzy@kadabra"},
		"Password": {"abcdef"},
		"Salt":     {"789000"},
		"Role":     {"None"},
	}

	response = httptest.NewRecorder()

	editAccount(response, request, 2) // Use AccountId 2, which does not have Admin role
	_checkResponseCode(t, response, 401)
	_checkResponseBody(t, response, "Unauthorized")

	editedAccount = Account{3, "ozzie", "ozzie@abrakadabra", "abcd", "789", "None", 1234567892}
	account, err = GetAccountById(3)
	if err != nil {
		t.Error(err)
		return
	}
	if *account != editedAccount {
		t.Errorf("GetAccountById() after editAccount() returned [%v], but expected account [%v]", account, editedAccount)
	}

	// ============================================ Valid Admin ============================================
	request, err = http.NewRequest("PUT", "http://localhost:8008/account/3", nil)
	if err != nil {
		t.Error(err)
		return
	}
	request.Form = url.Values{
		"Id":       {"3"},
		"Name":     {"OZZY"},
		"Email":    {"ozzy@nodev"},
		"Password": {"ABCDEF"},
		"Salt":     {"12345"},
		"Role":     {"None"},
	}

	response = httptest.NewRecorder()

	editAccount(response, request, 1) // Use AccountId 1, which has Admin role
	_checkResponseCode(t, response, 200)
	_checkResponseBody(t, response, "{\"Edit\": \"Success\"}")

	editedAccount = Account{3, "OZZY", "ozzy@nodev", "ABCDEF", "12345", "None", 1234567892}
	account, err = GetAccountById(3)
	if err != nil {
		t.Error(err)
		return
	}
	if *account != editedAccount {
		t.Errorf("GetAccountById() after editAccount() returned [%v], but expected account [%v]", account, editedAccount)
	}

	// ============================================ Nonmatching Id ============================================
	request, err = http.NewRequest("PUT", "http://localhost:8008/account/1", nil)
	if err != nil {
		t.Error(err)
		return
	}
	request.Form = url.Values{
		"Id":       {"2"}, // Id does not match Id used in URL
		"Name":     {"OZZY"},
		"Email":    {"ozzy@nodev"},
		"Password": {"ABCDEF"},
		"Salt":     {"12345"},
		"Role":     {"None"},
	}

	response = httptest.NewRecorder()

	editAccount(response, request, 1) // Use AccountId 1
	_checkResponseCode(t, response, 409)
	_checkResponseBody(t, response, "URL Id and Form Id do not match")

	account, err = GetAccountById(1)
	if err != nil {
		t.Error(err)
		return
	}
	if account.Name != "JamesClonk" {
		t.Errorf("GetAccountById() after editAccount() returned [%v], but expected account [%v]", account.Name, "JamesClonk")
	}

	editedAccount = Account{2, "Cluderzky", "clude@CLUDE", "abcd", "123456", "User", 1234567891}
	account, err = GetAccountById(2)
	if err != nil {
		t.Error(err)
		return
	}
	if *account != editedAccount {
		t.Errorf("GetAccountById() after editAccount() returned [%v], but expected account [%v]", account, editedAccount)
	}

	// ============================================ Valid Nonexisting Account, but not Admin ============================================
	request, err = http.NewRequest("PUT", "http://localhost:8008/account/7", nil)
	if err != nil {
		t.Error(err)
		return
	}
	request.Form = url.Values{
		"Id":       {"7"}, // Id does not yet exist
		"Name":     {"Hadron"},
		"Email":    {"hadron@hadron"},
		"Password": {"ABCDEFGH"},
		"Salt":     {"1234567"},
		"Role":     {"User"},
		"LastAuth": {"12345"},
	}

	response = httptest.NewRecorder()

	editAccount(response, request, 2) // Use AccountId 2, which has not "Admin" role
	_checkResponseCode(t, response, 401)
	_checkResponseBody(t, response, "Unauthorized")

	account, err = GetAccountById(7)
	if err == nil || account != nil {
		t.Errorf("GetAccountById() after unauthorized editAccount() returned [%v], but expected nil", account)
	}

	// ============================================ Valid Nonexisting Account ============================================
	request, err = http.NewRequest("PUT", "http://localhost:8008/account/7", nil)
	if err != nil {
		t.Error(err)
		return
	}
	request.Form = url.Values{
		"Id":       {"7"}, // Id does not yet exist
		"Name":     {"Hadron"},
		"Email":    {"hadron@hadron"},
		"Password": {"ABCDEFGH"},
		"Salt":     {"1234567"},
		"Role":     {"User"},
		"LastAuth": {"12345"},
	}

	response = httptest.NewRecorder()

	editAccount(response, request, 1) // Use AccountId 1, which has "Admin" role
	_checkResponseCode(t, response, 200)
	_checkResponseBody(t, response, "{\"Edit\": \"Success\"}")

	editedAccount = Account{7, "Hadron", "hadron@hadron", "ABCDEFGH", "1234567", "User", 0} // LastAuth cannot be set
	account, err = GetAccountById(7)
	if err != nil {
		t.Error(err)
		return
	}
	if *account != editedAccount {
		t.Errorf("GetAccountById() after editAccount() returned [%v], but expected account [%v]", account, editedAccount)
	}

	// ============================================ Unauthorized Nonexisting Account ============================================
	request, err = http.NewRequest("PUT", "http://localhost:8008/account/9", nil)
	if err != nil {
		t.Error(err)
		return
	}
	request.Form = url.Values{
		"Id":       {"9"}, // Id does not yet exist
		"Name":     {"Hadron 2"},
		"Email":    {"hadron@hadron"},
		"Password": {"ABCDEFGH"},
		"Salt":     {"1234567"},
		"Role":     {"User"},
		"LastAuth": {"12345"},
	}

	response = httptest.NewRecorder()

	editAccount(response, request, 2) // Use AccountId 2
	_checkResponseCode(t, response, 401)
	_checkResponseBody(t, response, "Unauthorized")

	account, err = GetAccountById(9)
	if err == nil || account != nil {
		t.Errorf("GetAccountById() after unauthorized editAccount() returned [%v], but expected nil", account)
	}

	// ============================================ Nonexisting Account ============================================
	request, err = http.NewRequest("PUT", "http://localhost:8008/account/17", nil)
	if err != nil {
		t.Error(err)
		return
	}
	request.Form = url.Values{
		"Id":        {"17"}, // account 17 does not yet exist
		"AccountId": {"55"},
	}
	response = httptest.NewRecorder()

	editAccount(response, request, 88)
	_checkResponseCode(t, response, 401)
	_checkResponseBody(t, response, "Unauthorized")
}

func Test_todo_deleteAccount(t *testing.T) {
	// ============================================ Valid ============================================
	request, err := http.NewRequest("DELETE", "http://localhost:8008/account/3", nil)
	if err != nil {
		t.Error(err)
		return
	}
	response := httptest.NewRecorder()

	deleteAccount(response, request, 3) // Use AccountId 3
	_checkResponseCode(t, response, 200)
	_checkResponseBody(t, response, "{\"Delete\": \"Success\"}")

	account, err := GetAccountById(3)
	if account != nil {
		t.Errorf("GetAccountById() after deleteAccount() still returned a account object, when it should not! Got [%v]", account)
		if err != nil {
			t.Error(err)
		}
	}

	// ============================================ Invalid ============================================
	request, err = http.NewRequest("DELETE", "http://localhost:8008/account/öäü", nil)
	if err != nil {
		t.Error(err)
		return
	}
	response = httptest.NewRecorder()

	deleteAccount(response, request, 1)
	_checkResponseCode(t, response, 404)
	_checkResponseBody(t, response, "404 page not found")

	// ============================================ Unauthorized ============================================
	request, err = http.NewRequest("DELETE", "http://localhost:8008/account/1", nil)
	if err != nil {
		t.Error(err)
		return
	}
	response = httptest.NewRecorder()

	deleteAccount(response, request, 2) // Use AccountId 2, which does not have Admin role
	_checkResponseCode(t, response, 401)
	_checkResponseBody(t, response, "Unauthorized")

	account, err = GetAccountById(1)
	if err != nil {
		t.Error(err)
	}
	if account == nil {
		t.Error("GetAccountById() after deleteAccount() did not return a account object, but it still should!")
	}

	// ============================================ Valid Admin ============================================
	request, err = http.NewRequest("DELETE", "http://localhost:8008/account/2", nil)
	if err != nil {
		t.Error(err)
		return
	}
	response = httptest.NewRecorder()

	deleteAccount(response, request, 1) // Use AccountId 1, which has Admin role
	_checkResponseCode(t, response, 200)
	_checkResponseBody(t, response, "{\"Delete\": \"Success\"}")

	account, err = GetAccountById(2)
	if account != nil {
		t.Errorf("GetAccountById() after deleteAccount() still returned a account object, when it should not! Got [%v]", account)
		if err != nil {
			t.Error(err)
		}
	}

	// ============================================ Nonexisting Account ============================================
	request, err = http.NewRequest("DELETE", "http://localhost:8008/account/77", nil)
	if err != nil {
		t.Error(err)
		return
	}
	response = httptest.NewRecorder()

	deleteAccount(response, request, 1)
	_checkResponseCode(t, response, 404)
	_checkResponseBody(t, response, "404 page not found")

	// ============================================ Nonexisting Account ============================================
	request, err = http.NewRequest("DELETE", "http://localhost:8008/account/1", nil)
	if err != nil {
		t.Error(err)
		return
	}
	response = httptest.NewRecorder()

	deleteAccount(response, request, 77)
	_checkResponseCode(t, response, 401)
	_checkResponseBody(t, response, "Unauthorized")
}

func Test_todo_cleanup(t *testing.T) {
	_storage_cleanup()
}

func _checkResponseCode(t *testing.T, response *httptest.ResponseRecorder, expected int) {
	code := response.Code
	if code != expected {
		t.Errorf("Response code was [%v], but expected [%v]", code, expected)
	}
}

func _checkResponseBody(t *testing.T, response *httptest.ResponseRecorder, expected string) {
	body := response.Body.String()
	if !strings.Contains(body, expected) {
		t.Errorf("Response body was [%v], but expected it to contain [%v]", body, expected)
	}
}
