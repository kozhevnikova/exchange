package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"math/rand"
	"net/http"
	"os"
	"strconv"
	"time"
)

type router struct {
	db *sql.DB
}

type UserLogin struct {
	Login    string `json:"login"`
	Password string `json:"password"`
}

type StoreUser struct {
	Userid int
	Login  string
	Role   string
}

type UserRegistrationData struct {
	Login     string `json:"login"`
	Role      string `json:"role"`
	Password  string `json:"password"`
	Firstname string `json:"firstname"`
	Lastname  string `json:"lastname"`
	Position  string `json:"position"`
}

type MainInformation struct {
	Position  string
	Firstname string
	Lastname  string
}

type TaskCreation struct {
	Title       string
	Description string
	Budget      string
}

type TaskInformation struct {
	Taskid      int       `json:"taskid"`
	Title       string    `json:"title"`
	Date        time.Time `json:"date"`
	Description string    `json:"description"`
	Budget      string    `json:"budget"`
	Customer    string    `json:"customer"`
	Customerid  int       `json:"customerid"`
}

type Contract struct {
	Taskid      int
	Bookid      int
	Customerid  int
	Status      string
	Title       string
	Description string
	Budget      string
}

type Proposal struct {
	Taskid      int
	Bookid      int
	Status      string
	Title       string
	Description string
	Budget      string
}
type ChangeStatus struct {
	Taskid int
	Bookid int
	Status string
}

type Account struct {
	AccountNumber string
	Balance       string
}

type Config struct {
	Database struct {
		User     string
		Password string
		Name     string
		Host     string
	}
}

func Logout(w http.ResponseWriter, r *http.Request) {
	cookie := &http.Cookie{
		Name:   cookieName,
		Value:  "",
		Path:   "/",
		MaxAge: -1,
	}
	http.SetCookie(w, cookie)
	http.Redirect(w, r, "/", 302)
}

func checkAuth(userid int, username string) bool {
	if userid == 0 || username == "" {
		return false
	}
	return true
}

func HandleLoginPage(w http.ResponseWriter, r *http.Request) {
	err := loginT.Execute(w, r)
	if err != nil {
		fmt.Fprintln(os.Stderr,
			"FATAL > login template execution > internal server error >", err)
		w.WriteHeader(http.StatusInternalServerError)
	}
}

func (router *router) Login(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	var u UserLogin

	err := json.NewDecoder(r.Body).Decode(&u)
	if err != nil {
		fmt.Fprintln(os.Stderr, "ERROR > login json >", err)
		w.WriteHeader(http.StatusBadRequest)
	}

	defer Close(r.Body)

	var returnuserid int
	var returnlogin string
	var returnpassword string
	var returnrole string

	if u.Login != "" && u.Password != "" {
		err = router.db.QueryRow(
			`SELECT
			userid,
			login,
			password,role FROM users WHERE login = $1 `, u.Login).Scan(
			&returnuserid,
			&returnlogin,
			&returnpassword,
			&returnrole)

		if err == sql.ErrNoRows {
			w.WriteHeader(http.StatusNotFound)
		} else if err == nil && checkPasswordHash(u.Password, returnpassword) {
			user := &StoreUser{
				Userid: returnuserid,
				Login:  returnlogin,
				Role:   returnrole,
			}
			err := user.SetCookieValues(w)
			if err != nil {
				fmt.Fprintln(os.Stderr,
					"ERROR > set cookies in login template > internal server error >", err)
				w.WriteHeader(http.StatusInternalServerError)
			}
			w.WriteHeader(http.StatusAccepted)
		} else if err == sql.ErrConnDone {
			w.WriteHeader(http.StatusInternalServerError)
		} else if err == sql.ErrTxDone {
			w.WriteHeader(http.StatusInternalServerError)
		} else {
			w.WriteHeader(http.StatusNotFound)
		}
	} else {
		w.WriteHeader(http.StatusUnprocessableEntity)
	}
}

func HandleRegistrationPage(w http.ResponseWriter, r *http.Request) {
	if err := registrationT.Execute(w, r); err != nil {
		fmt.Fprintln(os.Stderr,
			"ERROR > registration template execution > internal server error >", err)
		w.WriteHeader(http.StatusInternalServerError)
	}
}

func (router *router) Registration(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	var u UserRegistrationData

	if err := json.NewDecoder(r.Body).Decode(&u); err != nil {
		fmt.Fprintln(os.Stderr,
			"FATAL > decoding json in registration template > ", err)
	}
	defer Close(r.Body)

	if u.Login != "" && u.Firstname != "" && u.Lastname != "" &&
		u.Password != "" && u.Position != "" && u.Role != "" {
		result, err := router.db.Exec(`select login from users where login = $1`, u.Login)
		if err != nil {
			fmt.Fprintln(os.Stderr, "ERROR > checking login exists >", err)
		}

		count, err := result.RowsAffected()
		if err != nil {
			fmt.Fprintln(os.Stderr, "ERROR > getting count of records >", err)
		}

		if count == 0 {
			var userid int
			hashedpasssord, err := hashPassword(u.Password)
			if err != nil {
				fmt.Fprintln(os.Stderr, "ERROR > hashing password >", err)
			}
			_ = router.db.QueryRow(`insert into users(
					login,
					role,
					password,
					firstname,
					lastname,
					position) values($1,$2,$3,$4,$5,$6) returning userid`,
				u.Login,
				u.Role,
				hashedpasssord,
				u.Firstname,
				u.Lastname,
				u.Position).Scan(&userid)
			if err != nil {
				fmt.Fprintln(os.Stderr,
					"ERROR > inserting data in database from registration template >", err)
			}
			router.CreateAccount(u.Login, userid)
		} else {
			w.WriteHeader(http.StatusConflict)
		}
	} else {
		w.WriteHeader(http.StatusUnprocessableEntity)
	}
}

func (router *router) HandleMainPage(w http.ResponseWriter, r *http.Request) {
	inf, err := router.GetMainInformation(r)
	if err != nil {
		fmt.Fprintln(os.Stderr,
			"ERROR > cannot get main information for main template > internal server error >", err)
		w.WriteHeader(http.StatusInternalServerError)
	}

	if err = mainpageT.Execute(w, inf); err != nil {
		fmt.Fprintln(os.Stderr,
			"FATAL > main page template execution > internal server error >", err)
		w.WriteHeader(http.StatusInternalServerError)
	}
}

func (router *router) GetMainInformation(r *http.Request) (MainInformation, error) {
	var u MainInformation

	userid, login, _, err := readCookies(r)
	if err != nil {
		return u, err
	}

	err = router.db.QueryRow(`select 
			firstname, lastname, position 
			from users where userid=$1 and login=$2`,
		userid, login).Scan(&u.Firstname, &u.Lastname, &u.Position)
	if err != nil {
		return u, err
	}

	return u, nil
}

func HandleCreateTaskPage(w http.ResponseWriter, r *http.Request) {
	_, _, role, err := readCookies(r)
	if err != nil {
		fmt.Fprintln(os.Stderr,
			"WARNING > cannot read cookies from tasks page > returning zero values >", err)
	}

	if role == "freelancer" {
		w.WriteHeader(http.StatusForbidden)
	} else if role == "customer" {
		err := createTaskT.Execute(w, r)
		if err != nil {
			fmt.Fprintln(os.Stderr,
				"FATAL > create task template execution > internal server error >", err)
			w.WriteHeader(http.StatusInternalServerError)
		}
	} else {
		http.Redirect(w, r, "/", http.StatusForbidden)
	}
}

func (router *router) CreateTask(w http.ResponseWriter, r *http.Request) {
	var c TaskCreation

	err := json.NewDecoder(r.Body).Decode(&c)
	if err != nil {
		fmt.Fprintln(os.Stderr,
			"FATAL > create task function reading json >", err)
		w.WriteHeader(http.StatusUnprocessableEntity)
	}

	userid, _, _, err := readCookies(r)
	if err != nil {
		fmt.Fprintln(os.Stderr,
			"ERROR > readig cookies for create tasks function >", err)
	}

	_, err = router.db.Exec(`insert into 
		tasks(title,date,description,budget,customerid) values($1,$2,$3,$4,$5)`,
		c.Title, time.Now(), c.Description, c.Budget, userid)
	if err != nil {
		fmt.Fprintln(os.Stderr,
			`ERROR > insertion data in database from create task template >
			unprocessable entity >`, err)
		w.WriteHeader(http.StatusUnprocessableEntity)
	}
}

func (router *router) HandleTasksPage(w http.ResponseWriter, r *http.Request) {
	_, _, role, err := readCookies(r)
	if err != nil {
		fmt.Fprintln(os.Stderr,
			`WARNING > cannot read cookies from tasks page > returning zero values >`, err)
	}

	if role == "customer" {
		tasks, err := router.GetAllTasksForCustomer(r)
		if err != nil {
			fmt.Fprintln(os.Stderr, "ERROR > getting tasks for customer > internal server error >", err)
			w.WriteHeader(http.StatusInternalServerError)
		}

		err = customerTasksT.Execute(w, tasks)
		if err != nil {
			fmt.Fprintln(os.Stderr,
				"FATAL > customer tasks template execution > internal server error >", err)
			w.WriteHeader(http.StatusInternalServerError)
		}

	} else if role == "freelancer" {
		tasks, err := router.GetAllTasksForFreelancer()

		if err != nil {
			fmt.Fprintln(os.Stderr,
				"ERROR > getting tasks for freelancer > internal server error >", err)
			w.WriteHeader(http.StatusInternalServerError)
		}

		if err = freelancerTasksT.Execute(w, tasks); err != nil {
			fmt.Fprintln(os.Stderr,
				"FATAL > freelancer tasks template execution > internal server error >", err)
			w.WriteHeader(http.StatusInternalServerError)
		}

	} else {
		http.Redirect(w, r, "/", http.StatusForbidden)
	}
}

func (router *router) GetAllTasksForFreelancer() ([]TaskInformation, error) {
	var t []TaskInformation

	rows, err := router.db.Query(`select 
		T.taskid, T.title, T.date,T.description,T.budget,T.customerid, U.login from tasks T 
		join users U on T.customerid=U.userid`)
	if err != nil {
		return t, err
	}

	for rows.Next() {
		var tt TaskInformation

		err := rows.Scan(
			&tt.Taskid,
			&tt.Title,
			&tt.Date,
			&tt.Description,
			&tt.Budget,
			&tt.Customerid,
			&tt.Customer)
		if err != nil {
			return t, err
		}

		t = append(t, tt)
	}

	if err := rows.Close(); err != nil {
		fmt.Fprintln(os.Stderr, "ERROR > closing rows >", err)
	}

	return t, nil
}

func (router *router) GetAllTasksForCustomer(r *http.Request) ([]TaskInformation, error) {
	var t []TaskInformation

	userid, _, _, err := readCookies(r)
	if err != nil {
		return t, err
	}

	rows, err := router.db.Query(`select
		T.taskid,T.title, T.date,T.description,T.budget,U.login from tasks T 
		join users U on T.customerid=U.userid where T.customerid = $1`, userid)
	if err != nil {
		return t, err
	}

	for rows.Next() {
		var tt TaskInformation

		if err := rows.Scan(&tt.Taskid, &tt.Title, &tt.Date, &tt.Description, &tt.Budget, &tt.Customer); err != nil {
			return t, err
		}

		t = append(t, tt)
	}

	if err := rows.Close(); err != nil {
		fmt.Fprintln(os.Stderr, "ERROR > closing rows >", err)
	}

	return t, nil
}

func (router *router) CreateProposal(w http.ResponseWriter, r *http.Request) {
	freelancerid, _, role, err := readCookies(r)
	if err != nil {
		fmt.Fprintln(os.Stderr,
			"WARNING > cannot read cookies from tasks page > returning zero values >", err)
	}

	if role == "customer" {
		http.Redirect(w, r, "/", http.StatusForbidden)
	} else if role == "freelancer" {
		taskid, err := strconv.Atoi(r.FormValue("task"))
		if err != nil {
			fmt.Fprintln(os.Stderr, "ERROR > getting task id >", err)
		}

		customerid, err := strconv.Atoi(r.FormValue("customer"))
		if err != nil {
			fmt.Fprintln(os.Stderr, "ERROR > getting customer id >", err)
		}

		if err := r.ParseForm(); err != nil {
			fmt.Fprintln(os.Stderr, "ERROR > parsing form for creating proposal >", err)
			w.WriteHeader(http.StatusInternalServerError)
		}

		result, err := router.db.Exec(`select 
			status,taskid,customerid, freelancerid from books where 
			status=$1 and taskid=$2 and customerid=$3 and freelancerid=$4`,
			"reviewing", taskid, customerid, freelancerid)
		if err != nil {
			fmt.Fprintln(os.Stderr, "ERROR > checking existing proposals > internal server error >", err)
			w.WriteHeader(http.StatusInternalServerError)
		}

		count, _ := result.RowsAffected()
		if count == 0 {
			_, err = router.db.Exec(`insert into 
				books(status,taskid,customerid,freelancerid) 
				values($1,$2,$3,$4)`,
				"reviewing", taskid, customerid, freelancerid)
			if err != nil {
				fmt.Fprintln(os.Stderr, "ERROR > inserting data in book table > ", err)
				w.WriteHeader(http.StatusInternalServerError)
			} else {
				http.Redirect(w, r, "/tasks", http.StatusFound)
			}
		} else if count != 0 {
			w.WriteHeader(http.StatusConflict)
		}
	} else {
		http.Redirect(w, r, "/", http.StatusForbidden)
	}
}

func (router *router) HandleMyContractsPage(w http.ResponseWriter, r *http.Request) {
	_, _, role, err := readCookies(r)
	if err != nil {
		fmt.Fprintln(os.Stderr,
			"WARNING > cannot read cookies from tasks page > returning zero values >", err)
	}

	if role == "customer" {
		http.Redirect(w, r, "/", http.StatusForbidden)
	} else if role == "freelancer" {
		contracts, err := router.GetAllContracts()
		if err != nil {
			fmt.Fprintln(os.Stderr,
				"ERROR > scanning values for GetAllContracts func > internal server error> ", err)
			w.WriteHeader(http.StatusInternalServerError)
		}

		err = myContractsT.Execute(w, contracts)
		if err != nil {
			fmt.Fprintln(os.Stderr, "FATAL > contracts template execution >", err)
			w.WriteHeader(http.StatusInternalServerError)
		}
	} else {
		http.Redirect(w, r, "/", http.StatusForbidden)
	}
}

func (router *router) GetAllContracts() ([]Contract, error) {
	var c []Contract

	rows, err := router.db.Query(`select 
		T.taskid, B.bookid, T.customerid, B.status, T.title, T.description, T.budget 
		from tasks T join books B on T.taskid = B.taskid`)
	if err != nil {
		return c, err
	}

	for rows.Next() {
		var cc Contract

		if err := rows.Scan(&cc.Taskid, &cc.Bookid, &cc.Customerid,
			&cc.Status, &cc.Title, &cc.Description, &cc.Budget); err != nil {
			return c, err
		}

		c = append(c, cc)
	}

	if err := rows.Close(); err != nil {
		fmt.Fprintln(os.Stderr, "ERROR > closing rows >", err)
	}

	return c, nil
}

func (router *router) DeleteTask(w http.ResponseWriter, r *http.Request) {
	taskid, err := strconv.Atoi(r.FormValue("task"))
	if err != nil {
		fmt.Fprintln(os.Stderr, "ERROR > type convertion, delete task function >", err)
		w.WriteHeader(http.StatusInternalServerError)
	}

	customerid, _, _, err := readCookies(r)
	if err != nil {
		fmt.Fprintln(os.Stderr, "ERROR > reading cookies in delete task function >", err)
		w.WriteHeader(http.StatusInternalServerError)
	}

	_, err = router.db.Exec(`delete from 
		tasks where taskid=$1 and customerid=$2`, taskid, customerid)
	if err != nil {
		fmt.Fprintln(os.Stderr, "ERROR > deleting task in table >", err)
		w.WriteHeader(http.StatusInternalServerError)
	}

	http.Redirect(w, r, "/tasks", 302)
}

func (router *router) DeleteProposal(w http.ResponseWriter, r *http.Request) {
	taskid, err := strconv.Atoi(r.FormValue("task"))
	if err != nil {
		fmt.Fprintln(os.Stderr,
			"ERROR > type convertion, delete task function >", err)
		w.WriteHeader(http.StatusInternalServerError)
	}

	customerid, err := strconv.Atoi(r.FormValue("customer"))
	if err != nil {
		fmt.Fprintln(os.Stderr,
			"ERROR > type convertion, delete task function >", err)
		w.WriteHeader(http.StatusInternalServerError)
	}

	freelancerid, _, _, err := readCookies(r)
	if err != nil {
		fmt.Fprintln(os.Stderr,
			"ERROR > reading cookies in delete task function >", err)
		w.WriteHeader(http.StatusInternalServerError)
	}

	_, err = router.db.Exec(`delete from 
		books where taskid=$1 and freelancerid=$2 and customerid=$3`,
		taskid, freelancerid, customerid)
	if err != nil {
		fmt.Fprintln(os.Stderr, "ERROR > deleting task in table >", err)
		w.WriteHeader(http.StatusInternalServerError)
	}

	http.Redirect(w, r, "/tasks", 302)
}

func (router *router) HandleSeeProposals(w http.ResponseWriter, r *http.Request) {
	_, _, role, err := readCookies(r)
	if err != nil {
		fmt.Fprintln(os.Stderr, "ERROR > reading cookies in seeProposals template", err)
	}

	if role == "customer" {
		proposals, err := router.GetAllProposals(r)
		if err != nil {
			fmt.Fprintln(os.Stderr, "ERROR > getting proposals > internal server error >", err)
			w.WriteHeader(http.StatusInternalServerError)
		}

		if err := seeProposalsT.Execute(w, proposals); err != nil {
			fmt.Fprintln(os.Stderr, "ERROR > see proposals execution >", err)
		}

	} else if role == "freelancer" {
		http.Redirect(w, r, "/", http.StatusForbidden)
	} else {
		http.Redirect(w, r, "/", http.StatusForbidden)
	}
}

func (router *router) GetAllProposals(r *http.Request) ([]Proposal, error) {
	var p []Proposal

	customerid, _, _, err := readCookies(r)
	if err != nil {
		return p, err
	}

	rows, err := router.db.Query(`select 
		B.taskid, B.bookid, B.status, T.title,T.description,T.budget from books B 
		join tasks T on T.customerid= B.customerid where T.customerid=$1`,
		customerid)
	if err != nil {
		fmt.Fprintln(os.Stderr, "ERROR > while reading proposals >", err)
	}

	for rows.Next() {
		var pp Proposal

		if err := rows.Scan(&pp.Taskid, &pp.Bookid, &pp.Status,
			&pp.Title, &pp.Description, &pp.Budget); err != nil {
			return p, err
		}

		p = append(p, pp)
	}
	if err := rows.Close(); err != nil {
		return p, err
	}
	return p, nil
}

func (router *router) CreateAccount(login string, userid int) {
	accountnumber := login + "-" + strconv.Itoa(userid) + "-" + strconv.Itoa(rand.Int())

	if _, err := router.db.Exec(`insert into 
		accounts(userid, accountnumber, balance) values($1,$2, $3)`,
		userid, accountnumber, 0); err != nil {
		fmt.Fprintln(os.Stderr, "ERROR > insert into accounts >", err)
	}
}

func (router *router) Transaction(c ChangeStatus, userid int) {
	var freelancerid int

	_ = router.db.QueryRow(`
		select freelancerid from books where taskid=$1 and customerid=$2`,
		c.Taskid, userid).Scan(&freelancerid)

	_, err := router.db.Exec(`
		update accounts set balance=balance-(
			select budget from tasks where taskid=$1 and customerid=$2 for update) 
			where userid=$2`,
		c.Taskid, userid)
	if err != nil {
		fmt.Fprintln(os.Stderr,
			"ERROR > while money transaction > internal server error >", err)
	}

	_, err = router.db.Exec(`
		update accounts set balance=balance+(
			select budget from tasks where taskid=$1 and customerid=$2 for update) 
			where userid=$3`,
		c.Taskid, userid, freelancerid)
	if err != nil {
		fmt.Fprintln(os.Stderr,
			"ERROR > while money transaction > internal server error >", err)
	}
}

func (router *router) ChangeStatus(w http.ResponseWriter, r *http.Request) {
	var c ChangeStatus

	c.Taskid, _ = strconv.Atoi(r.FormValue("task"))
	c.Status = r.FormValue("status")
	c.Bookid, _ = strconv.Atoi(r.FormValue("book"))

	userid, _, _, err := readCookies(r)
	if err != nil {
		fmt.Fprintln(os.Stderr, "ERROR > reading cookies in change status > ", err)
	}

	if c.Status == "finished" {
		if _, err := router.db.Exec(`
		update books set status=$1 where taskid=$2 and bookid=$3`,
			c.Status, c.Taskid, c.Bookid); err != nil {
			fmt.Fprintln(os.Stderr,
				"ERROR > during changing status > internal server error >", err)
			w.WriteHeader(http.StatusInternalServerError)
		}

		router.Transaction(c, userid)
	} else {
		if _, err := router.db.Exec(`
		update books set status=$1 where taskid=$2 and bookid=$3`,
			c.Status, c.Taskid, c.Bookid); err != nil {
			fmt.Fprintln(os.Stderr,
				"ERROR > during changing status > internal server error >", err)
			w.WriteHeader(http.StatusInternalServerError)
		}
	}

	http.Redirect(w, r, "/seeProposals", 302)
}

func (router *router) HandleAccountPage(w http.ResponseWriter, r *http.Request) {
	data := router.GetAccountData(w, r)

	_, _, role, err := readCookies(r)
	if err != nil {
		fmt.Fprintln(os.Stderr,
			"ERROR > reading cookies in accounts template", err)
	}

	if role == "customer" {
		if err := customerAccountT.Execute(w, data); err != nil {
			fmt.Fprintln(os.Stderr,
				"ERROR > account template execution > internal server error >", err)
			w.WriteHeader(http.StatusInternalServerError)
		}
	} else if role == "freelancer" {
		if err := freelancerAccountT.Execute(w, data); err != nil {
			fmt.Fprintln(os.Stderr,
				"ERROR > account template execution > internal server error >", err)
			w.WriteHeader(http.StatusInternalServerError)
		}
	} else {
		http.Redirect(w, r, "/", http.StatusForbidden)
	}
}

func (router *router) GetAccountData(w http.ResponseWriter, r *http.Request) Account {
	userid, _, _, err := readCookies(r)
	if err != nil {
		fmt.Fprintln(os.Stderr,
			"ERROR > reading cookies in accounts template", err)
	}

	var a Account

	_ = router.db.QueryRow(`select 
		accountnumber, balance from accounts where userid=$1`, userid).Scan(
		&a.AccountNumber, &a.Balance)

	return a
}

func (router *router) AddMoney(w http.ResponseWriter, r *http.Request) {
	userid, _, role, err := readCookies(r)
	if err != nil {
		fmt.Fprintln(os.Stderr,
			"ERROR > reading cookies in add money >", err)
	}

	if role == "customer" {
		var a Account

		if err := json.NewDecoder(r.Body).Decode(&a); err != nil {
			fmt.Fprintln(os.Stderr,
				"ERROR > reading json in add money function > unprocessable entity >", err)
			w.WriteHeader(http.StatusUnprocessableEntity)
		}
		if err != nil {
			fmt.Fprintln(os.Stderr,
				"ERROR > reading cookies in add money template", err)
		}

		if _, err := router.db.Exec(`update 
			accounts set balance=balance+$1 where userid=$2`,
			a.Balance, userid); err != nil {
			fmt.Fprintln(os.Stderr,
				"ERROR > while updating balance > internal server error >", err)
			w.WriteHeader(http.StatusInternalServerError)
		}
	} else if role == "freelancer" {
		http.Redirect(w, r, "/account", http.StatusForbidden)
	} else {
		http.Redirect(w, r, "/account", http.StatusForbidden)
	}
}

func CheckCredentials(fn func(w http.ResponseWriter, r *http.Request)) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		userID, _, login, err := readCookies(r)
		if err != nil {
			if err == http.ErrNoCookie {
				http.Redirect(w, r, "/", http.StatusFound)
				return
			}

			fmt.Fprintln(os.Stderr,
				"FATAL > login template execution > internal server error >", err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		if !checkAuth(userID, login) {
			http.Redirect(w, r, "/", http.StatusFound)
			return
		}
		fn(w, r)
	}
}
