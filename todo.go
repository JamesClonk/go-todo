package main

import "fmt"
import "log"
import "flag"
import "regexp"
import "strconv"
import "strings"
import "time"
import "errors"
import "net/http"
import "encoding/json"

var isLogging = true
var validPath = regexp.MustCompile("^/(task|account)/([a-zA-Z0-9_]+)$")

type MethodHandler map[string]func(w http.ResponseWriter, r *http.Request, accountId int)

var fileFlag = flag.String("database", "./data/tasks.db", "database file")
var databaseFlag = flag.Bool("createDatabase", false, "will setup a new empty database")
var adminFlag = flag.Bool("createAdmin", false, "will create a new admin account in the database")
var taskFlag = flag.Bool("createTasks", false, "will create some sample tasks in the database")

func main() {
	parseCommandline()

	http.Handle("/client/", http.StripPrefix("/client/", http.FileServer(http.Dir("client/"))))

	http.HandleFunc("/auth/", getAuth)
	http.HandleFunc("/tasks/", authHandler(MethodHandler{
		"GET": getTasks,
	}))
	http.HandleFunc("/task/", authHandler(MethodHandler{
		"GET":    getTask,
		"POST":   addTask,
		"PUT":    editTask,
		"DELETE": deleteTask,
	}))

	http.HandleFunc("/accounts/", authHandler(MethodHandler{
		"GET": getAccounts,
	}))
	http.HandleFunc("/account/", authHandler(MethodHandler{
		"GET":    getAccount,
		"POST":   addAccount,
		"PUT":    editAccount,
		"DELETE": deleteAccount,
	}))

	http.ListenAndServe(":8008", nil)
}

func parseCommandline() {
	flag.Parse()

	SetDatabase(*fileFlag)

	if *databaseFlag {
		SetupDatabase()
	}

	if *adminFlag {
		account, password := SetupAdmin()
		log.Printf("Admin account created: [%v], with password: [%v]", account, password)
	}

	if *taskFlag {
		SetupSampleTasks()
	}
}

func authHandler(mh MethodHandler) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if isLogging {
			log.Printf("%v, %v, %v", r.RemoteAddr, r.Method, r.RequestURI)
		}

		accountId, err := Authenticate(r)
		if err != nil || accountId == nil {
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}

		if isLogging {
			log.Printf("User is authenticated[%v]", *accountId)
		}

		mh[r.Method](w, r, *accountId)
	}
}

func getLogin(r *http.Request) string {
	query := r.URL.Query()
	return query.Get("login")
}

func getId(w http.ResponseWriter, r *http.Request) (int, error) {
	// validate URL Path
	v := validPath.FindStringSubmatch(r.URL.Path)
	if v == nil {
		http.NotFound(w, r)
		return -1, errors.New("Invalid URL")
	}
	return strconv.Atoi(v[2])
}

func getAuth(w http.ResponseWriter, r *http.Request) {
	if isLogging {
		log.Println("get Auth")
	}

	email := getLogin(r)
	if isLogging {
		log.Printf("Auth Email: [%v]", email)
	}

	account, err := GetAccountByEmail(email)
	if err != nil || account == nil {
		http.NotFound(w, r)
		return
	}
	if account.Role == "None" || account.Role == "" {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	// update last auth timestamp
	account.LastAuth = int(time.Now().Unix())
	if err := account.Save(); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	id := strconv.Itoa(account.Id)
	timestamp := fmt.Sprintf("%d", time.Now().Unix())
	auth := `{"AccountId": ` + id + `, "Salt": "` + account.Salt + `", "Timestamp": ` + timestamp + `}`

	w.Header().Set("Content-Type", "application/json")
	w.Write([]byte(auth))
}

func getTasks(w http.ResponseWriter, r *http.Request, accountId int) {
	if isLogging {
		log.Println("get Tasks")
	}

	tasks, err := GetTasksByAccountId(accountId)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	js, err := json.Marshal(tasks)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.Write(js)
}

func getTask(w http.ResponseWriter, r *http.Request, accountId int) {
	id, err := getId(w, r)
	if err != nil {
		return
	}

	if isLogging {
		log.Printf("get Task[%v]", id)
	}

	task, err := GetTaskById(id)
	if err != nil {
		if strings.Trim(err.Error(), "\n") == "sql: no rows in result set" {
			http.NotFound(w, r)
			return
		} else {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	}

	// check if task belongs to account id, or if account has role "Admin"
	if task.AccountId != accountId {
		account, err := GetAccountById(accountId)
		if err != nil {
			if strings.Trim(err.Error(), "\n") == "sql: no rows in result set" {
				http.Error(w, "Unauthorized", http.StatusUnauthorized)
				return
			} else {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
		}
		if account.Role != "Admin" {
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}
	}

	js, err := json.Marshal(task)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.Write(js)
}

func addTask(w http.ResponseWriter, r *http.Request, accountId int) {
	if isLogging {
		log.Println("add Task")
	}

	data := r.Form
	id := -1 // POST ignores taskId and always uses -1 to create a new task entry
	accId, err := strconv.Atoi(data.Get("AccountId"))
	if err != nil {
		http.Error(w, "Invalid data", http.StatusBadRequest)
		return
	}

	// timestamps upon task creating are enforced by server and cannot be overwritten by client
	created := int(time.Now().Unix())
	lastUpdated := int(time.Now().Unix())

	priority, err := strconv.Atoi(data.Get("Priority"))
	if err != nil {
		http.Error(w, "Invalid data", http.StatusBadRequest)
		return
	}

	task := Task{
		id,
		accId,
		created,
		lastUpdated,
		priority,
		data.Get("Task"),
	}

	// check if task belongs to account id, or if account has role "Admin"
	if task.AccountId != accountId {
		account, err := GetAccountById(accountId)
		if err != nil {
			if strings.Trim(err.Error(), "\n") == "sql: no rows in result set" {
				http.Error(w, "Unauthorized", http.StatusUnauthorized)
				return
			} else {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
		}
		if account.Role != "Admin" {
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}
	}

	if err := task.Save(); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.Write([]byte("{\"Add\": \"Success\"}"))
}

func editTask(w http.ResponseWriter, r *http.Request, accountId int) {
	id, err := getId(w, r)
	if err != nil {
		return
	}

	if isLogging {
		log.Printf("edit Task[%v]", id)
	}

	data := r.Form
	formAccountId, err := strconv.Atoi(data.Get("AccountId"))
	if err != nil {
		http.Error(w, "Invalid data", http.StatusBadRequest)
		return
	}

	task, err := GetTaskById(id)
	if err != nil {
		if strings.Trim(err.Error(), "\n") != "sql: no rows in result set" {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		} else {
			task = &Task{}
			task.Id = id
			task.AccountId = formAccountId
		}
	}

	// check if task belongs to account id, or if account has role "Admin"
	if task.AccountId != accountId {
		account, err := GetAccountById(accountId)
		if err != nil {
			if strings.Trim(err.Error(), "\n") == "sql: no rows in result set" {
				http.Error(w, "Unauthorized", http.StatusUnauthorized)
				return
			} else {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
		}
		if account.Role != "Admin" {
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}
		// overwrite accountId only possible if user has role "Admin"
		task.AccountId = formAccountId
	}

	formId, err := strconv.Atoi(data.Get("Id"))
	if err != nil {
		http.Error(w, "Invalid data", http.StatusBadRequest)
		return
	}
	if id != formId {
		http.Error(w, "URL Id and Form Id do not match", http.StatusConflict)
		return
	}

	lastUpdated, err := strconv.Atoi(data.Get("LastUpdated"))
	if err != nil || lastUpdated < 1 {
		// server takes care of lastUpdated timestamp in this case
		lastUpdated = int(time.Now().Unix())
	}

	priority, err := strconv.Atoi(data.Get("Priority"))
	if err != nil {
		http.Error(w, "Invalid data", http.StatusBadRequest)
		return
	}

	//task.Created = created // created cannot be overwritten by client
	if task.Created < 1 {
		task.Created = int(time.Now().Unix())
	}

	task.LastUpdated = lastUpdated
	task.Priority = priority
	task.Task = data.Get("Task")

	if err := task.Save(); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.Write([]byte("{\"Edit\": \"Success\"}"))
}

func deleteTask(w http.ResponseWriter, r *http.Request, accountId int) {
	id, err := getId(w, r)
	if err != nil {
		return
	}

	if isLogging {
		log.Printf("delete Task[%v]", id)
	}

	task, err := GetTaskById(id)
	if err != nil {
		if strings.Trim(err.Error(), "\n") == "sql: no rows in result set" {
			http.NotFound(w, r)
			return
		} else {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	}

	// check if task belongs to account id, or if account has role "Admin"
	if task.AccountId != accountId {
		account, err := GetAccountById(accountId)
		if err != nil {
			if strings.Trim(err.Error(), "\n") == "sql: no rows in result set" {
				http.Error(w, "Unauthorized", http.StatusUnauthorized)
				return
			} else {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
		}
		if account.Role != "Admin" {
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}
	}

	if err := task.Delete(); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.Write([]byte("{\"Delete\": \"Success\"}"))
}

func getAccounts(w http.ResponseWriter, r *http.Request, accountId int) {
	if isLogging {
		log.Println("get Accounts")
	}

	// check if account has role "Admin"
	acc, err := GetAccountById(accountId)
	if err != nil {
		if strings.Trim(err.Error(), "\n") == "sql: no rows in result set" {
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		} else {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	}
	if acc.Role != "Admin" {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	accounts, err := GetAllAccounts()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	js, err := json.Marshal(accounts)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.Write(js)
}

func getAccount(w http.ResponseWriter, r *http.Request, accountId int) {
	id, err := getId(w, r)
	if err != nil {
		return
	}

	if isLogging {
		log.Printf("get Account[%v]", id)
	}

	// check if account belongs to account id, or if account has role "Admin"
	if id != accountId {
		acc, err := GetAccountById(accountId)
		if err != nil {
			if strings.Trim(err.Error(), "\n") == "sql: no rows in result set" {
				http.Error(w, "Unauthorized", http.StatusUnauthorized)
				return
			} else {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
		}
		if acc.Role != "Admin" {
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}
	}

	account, err := GetAccountById(id)
	if err != nil {
		if strings.Trim(err.Error(), "\n") == "sql: no rows in result set" {
			http.NotFound(w, r)
			return
		} else {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	}

	// do not return Password, duh!
	account.Password = ""

	js, err := json.Marshal(account)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.Write(js)
}

func addAccount(w http.ResponseWriter, r *http.Request, accountId int) {
	if isLogging {
		log.Println("add Account")
	}

	data := r.Form

	account := Account{}
	account.Id = -1 // POST ignores accountId and always uses -1 to create a new account entry

	account.Name = data.Get("Name")
	if account.Name == "" {
		http.Error(w, "Invalid data", http.StatusBadRequest)
		return
	}
	account.Email = data.Get("Email")
	if account.Email == "" {
		http.Error(w, "Invalid data", http.StatusBadRequest)
		return
	}
	account.Password = data.Get("Password")
	if account.Password == "" {
		http.Error(w, "Invalid data", http.StatusBadRequest)
		return
	}
	account.Salt = data.Get("Salt")
	if account.Salt == "" {
		http.Error(w, "Invalid data", http.StatusBadRequest)
		return
	}
	account.Role = data.Get("Role")
	if account.Role == "" {
		http.Error(w, "Invalid data", http.StatusBadRequest)
		return
	}

	// check if account has role "Admin"
	acc, err := GetAccountById(accountId)
	if err != nil {
		if strings.Trim(err.Error(), "\n") == "sql: no rows in result set" {
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		} else {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	}
	if acc.Role != "Admin" {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	if err := account.Save(); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.Write([]byte("{\"Add\": \"Success\"}"))
}

func editAccount(w http.ResponseWriter, r *http.Request, accountId int) {
	id, err := getId(w, r)
	if err != nil {
		return
	}

	if isLogging {
		log.Printf("edit Account[%v]", id)
	}

	data := r.Form

	account, err := GetAccountById(id)
	if err != nil {
		if strings.Trim(err.Error(), "\n") != "sql: no rows in result set" {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		} else {
			account = &Account{}
			account.Id = id
		}
	}

	acc, err := GetAccountById(accountId)
	if err != nil {
		if strings.Trim(err.Error(), "\n") == "sql: no rows in result set" {
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		} else {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	}
	// check if account belongs to account id, or if account has role "Admin"
	if account.Id != accountId && acc.Role != "Admin" {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	formId, err := strconv.Atoi(data.Get("Id"))
	if err != nil {
		http.Error(w, "Invalid data", http.StatusBadRequest)
		return
	}
	if id != formId {
		http.Error(w, "URL Id and Form Id do not match", http.StatusConflict)
		return
	}

	account.Name = data.Get("Name")
	account.Email = data.Get("Email")
	account.Password = data.Get("Password")
	account.Salt = data.Get("Salt")
	if acc.Role == "Admin" { // only Admins can change roles
		account.Role = data.Get("Role")
	}

	if err := account.Save(); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.Write([]byte("{\"Edit\": \"Success\"}"))
}

func deleteAccount(w http.ResponseWriter, r *http.Request, accountId int) {
	id, err := getId(w, r)
	if err != nil {
		return
	}

	if isLogging {
		log.Printf("delete Account[%v]", id)
	}

	// check if account belongs to account id, or if account has role "Admin"
	if id != accountId {
		acc, err := GetAccountById(accountId)
		if err != nil {
			if strings.Trim(err.Error(), "\n") == "sql: no rows in result set" {
				http.Error(w, "Unauthorized", http.StatusUnauthorized)
				return
			} else {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
		}
		if acc.Role != "Admin" {
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}
	}

	account, err := GetAccountById(id)
	if err != nil {
		if strings.Trim(err.Error(), "\n") == "sql: no rows in result set" {
			http.NotFound(w, r)
			return
		} else {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	}

	if err := account.Delete(); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.Write([]byte("{\"Delete\": \"Success\"}"))
}
