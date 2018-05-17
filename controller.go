package main

import (
	"net/http"
)

func (router *router) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	method := r.Method
	url := r.URL.Path

	switch {
	case method == "GET" && url == "/":
		HandleLoginPage(w, r)

	case method == "POST" && url == "/login":
		router.Login(w, r)

	case method == "GET" && url == "/registration":
		HandleRegistrationPage(w, r)

	case method == "POST" && url == "/singup":
		router.Registration(w, r)

	case method == "GET" && url == "/main":
		CheckCredentials(router.HandleMainPage)(w, r)

	case method == "POST" && url == "/logout":
		Logout(w, r)

	case method == "GET" && url == "/tasks":
		CheckCredentials(router.HandleTasksPage)(w, r)

	case method == "GET" && url == "/createTask":
		CheckCredentials(HandleCreateTaskPage)(w, r)

	case method == "POST" && url == "/create":
		CheckCredentials(router.CreateTask)(w, r)

	case method == "POST" && url == "/newProposal":
		CheckCredentials(router.CreateProposal)(w, r)

	case method == "GET" && url == "/myContracts":
		CheckCredentials(router.HandleMyContractsPage)(w, r)

	case method == "POST" && url == "/deleteTask":
		CheckCredentials(router.DeleteTask)(w, r)

	case method == "POST" && url == "/deleteProposal":
		CheckCredentials(router.DeleteProposal)(w, r)

	case method == "GET" && url == "/seeProposals":
		CheckCredentials(router.HandleSeeProposals)(w, r)

	case method == "POST" && url == "/changeStatus":
		CheckCredentials(router.ChangeStatus)(w, r)

	case method == "GET" && url == "/account":
		CheckCredentials(router.HandleAccountPage)(w, r)

	case method == "POST" && url == "/addMoney":
		CheckCredentials(router.AddMoney)(w, r)
	default:
		w.WriteHeader(http.StatusNotFound)
	}
}
