package main

import "os"
import "log"
import "database/sql"
import _ "github.com/mattn/go-sqlite3"

var sqlAccounts = `
	create table T_ACCOUNTS (
		ID integer not null primary key, 
		NAME text not null, 
		EMAIL text not null,
		PASSWORD text not null,
		SALT text not null,
		ROLE text not null,
		LAST_AUTH integer not null
	);
	`

var sqlAccountIndex = `
	create unique index if not exists IDX_ACCOUNT_EMAIL ON T_ACCOUNTS (EMAIL);
	`

var sqlTasks = `
	create table T_TASKS (
		ID integer not null primary key, 
		ACCOUNT_ID integer not null,   
		CREATED integer not null, 
		LAST_UPDATED integer not null, 
		PRIORITY integer not null, 
		TASK text not null,
		foreign key(ACCOUNT_ID) references T_ACCOUNTS(ID)
	);
	`
var database = "./data/tasks.db"

func connect() (*sql.DB, error) {
	return sql.Open("sqlite3", database)
}

func SetDatabase(db string) {
	database = db
}

func SetupDatabase() {
	os.Remove(database)
	db, err := connect()
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	if _, err := db.Exec(sqlAccounts); err != nil {
		log.Fatal(err)
	}
	if _, err := db.Exec(sqlAccountIndex); err != nil {
		log.Fatal(err)
	}
	if _, err := db.Exec(sqlTasks); err != nil {
		log.Fatal(err)
	}
}

func SetupAdmin() Account {
	salt, err := GenerateRandomString()
	if err != nil {
		log.Fatal(err)
	}
	password, err := GenerateRandomString()
	if err != nil {
		log.Fatal(err)
	}
	a := Account{-1, "Admin", "admin@admin", HashPassword(*salt, *password), *salt, "Admin", 0}
	if err := a.Save(); err != nil {
		log.Fatal(err)
	}
	return a
}

func scanTasks(rows *sql.Rows) (*Tasks, error) {
	ts := Tasks{}
	for rows.Next() {
		var t Task
		if err := rows.Scan(&t.Id, &t.AccountId, &t.Created, &t.LastUpdated, &t.Priority, &t.Task); err != nil {
			return nil, err
		}
		ts = append(ts, t)
	}
	return &ts, nil
}

func scanAccounts(rows *sql.Rows) (*Accounts, error) {
	as := Accounts{}
	for rows.Next() {
		var a Account
		if err := rows.Scan(&a.Id, &a.Name, &a.Email, &a.Password, &a.Salt, &a.Role, &a.LastAuth); err != nil {
			return nil, err
		}
		as = append(as, a)
	}
	return &as, nil
}

func GetAllTasks() (*Tasks, error) {
	db, err := connect()
	if err != nil {
		return nil, err
	}
	defer db.Close()

	rows, err := db.Query("select * from T_TASKS order by PRIORITY desc, LAST_UPDATED asc, CREATED asc")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	ts, err := scanTasks(rows)
	if err != nil {
		return nil, err
	}

	return ts, nil
}

func GetTaskById(id int) (*Task, error) {
	db, err := connect()
	if err != nil {
		return nil, err
	}
	defer db.Close()

	stmt, err := db.Prepare("select * from T_TASKS where ID = ?")
	if err != nil {
		return nil, err
	}
	defer stmt.Close()

	var t Task
	if err := stmt.QueryRow(id).Scan(&t.Id, &t.AccountId, &t.Created, &t.LastUpdated, &t.Priority, &t.Task); err != nil {
		return nil, err
	} else {
		return &t, nil
	}
}

func GetTasksByAccountId(id int) (*Tasks, error) {
	db, err := connect()
	if err != nil {
		return nil, err
	}
	defer db.Close()

	stmt, err := db.Prepare("select * from T_TASKS where ACCOUNT_ID = ? order by PRIORITY desc, LAST_UPDATED asc, CREATED asc")
	if err != nil {
		return nil, err
	}
	defer stmt.Close()

	rows, err := stmt.Query(id)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	ts, err := scanTasks(rows)
	if err != nil {
		return nil, err
	}

	return ts, nil
}

func GetAllAccounts() (*Accounts, error) {
	db, err := connect()
	if err != nil {
		return nil, err
	}
	defer db.Close()

	rows, err := db.Query("select * from T_ACCOUNTS order by ID asc")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	as, err := scanAccounts(rows)
	if err != nil {
		return nil, err
	}

	return as, nil
}

func GetAccountById(id int) (*Account, error) {
	db, err := connect()
	if err != nil {
		return nil, err
	}
	defer db.Close()

	stmt, err := db.Prepare("select * from T_ACCOUNTS where ID = ?")
	if err != nil {
		return nil, err
	}
	defer stmt.Close()

	var a Account
	if err := stmt.QueryRow(id).Scan(&a.Id, &a.Name, &a.Email, &a.Password, &a.Salt, &a.Role, &a.LastAuth); err != nil {
		return nil, err
	} else {
		return &a, nil
	}
}

func GetAccountByEmail(email string) (*Account, error) {
	db, err := connect()
	if err != nil {
		return nil, err
	}
	defer db.Close()

	stmt, err := db.Prepare("select * from T_ACCOUNTS where EMAIL = ?")
	if err != nil {
		return nil, err
	}
	defer stmt.Close()

	var a Account
	if err := stmt.QueryRow(email).Scan(&a.Id, &a.Name, &a.Email, &a.Password, &a.Salt, &a.Role, &a.LastAuth); err != nil {
		return nil, err
	} else {
		return &a, nil
	}
}

func (ts Tasks) Save() error {
	db, err := connect()
	if err != nil {
		return err
	}
	defer db.Close()

	tx, err := db.Begin()
	if err != nil {
		return err
	}

	stmt, err := tx.Prepare("insert or replace into T_TASKS (ID, ACCOUNT_ID, CREATED, LAST_UPDATED, PRIORITY, TASK) values (?,?,?,?,?,?)")
	if err != nil {
		return err
	}
	defer stmt.Close()

	for i, t := range ts {
		var result sql.Result
		if t.Id < 1 {
			result, err = stmt.Exec(nil, t.AccountId, t.Created, t.LastUpdated, t.Priority, t.Task)
		} else {
			result, err = stmt.Exec(t.Id, t.AccountId, t.Created, t.LastUpdated, t.Priority, t.Task)
		}
		if err != nil {
			return err
		}

		id, err := result.LastInsertId()
		if err != nil {
			return err
		}
		t.Id = int(id)

		ts[i] = t
	}

	err = tx.Commit()
	if err != nil {
		return err
	}

	return nil
}

func (t *Task) Save() error {
	tasks := Tasks{*t}
	if err := tasks.Save(); err != nil {
		return err
	}

	t.Id = tasks[0].Id
	return nil
}

func (ts Tasks) Delete() error {
	db, err := connect()
	if err != nil {
		return err
	}
	defer db.Close()

	tx, err := db.Begin()
	if err != nil {
		return err
	}

	stmt, err := tx.Prepare("delete from T_TASKS where ID = ?")
	if err != nil {
		return err
	}
	defer stmt.Close()

	for i, t := range ts {
		if _, err := stmt.Exec(t.Id); err != nil {
			return err
		}
		t.Id = -1
		ts[i] = t
	}

	err = tx.Commit()
	if err != nil {
		return err
	}

	return nil
}

func (t *Task) Delete() error {
	tasks := Tasks{*t}
	if err := tasks.Delete(); err != nil {
		return err
	}

	t.Id = -1
	return nil
}

func (a *Account) Save() error {
	db, err := connect()
	if err != nil {
		return err
	}
	defer db.Close()

	stmt, err := db.Prepare("insert or replace into T_ACCOUNTS (ID, NAME, EMAIL, PASSWORD, SALT, ROLE, LAST_AUTH) values (?,?,?,?,?,?,?)")
	if err != nil {
		return err
	}
	defer stmt.Close()

	var result sql.Result
	if a.Id < 1 {
		result, err = stmt.Exec(nil, a.Name, a.Email, a.Password, a.Salt, a.Role, a.LastAuth)
	} else {
		result, err = stmt.Exec(a.Id, a.Name, a.Email, a.Password, a.Salt, a.Role, a.LastAuth)
	}
	if err != nil {
		return err
	}

	id, err := result.LastInsertId()
	if err != nil {
		return err
	}
	a.Id = int(id)

	return nil
}

func (a *Account) Delete() error {
	db, err := connect()
	if err != nil {
		return err
	}
	defer db.Close()

	stmt, err := db.Prepare("delete from T_ACCOUNTS where ID = ?")
	if err != nil {
		return err
	}
	defer stmt.Close()

	if _, err := stmt.Exec(a.Id); err != nil {
		return err
	}
	a.Id = -1

	return nil
}
