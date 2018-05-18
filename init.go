package main

import "html/template"

func init() {
	loginT = template.Must(template.ParseFiles(
		"./templates/login.html"))

	registrationT = template.Must(template.ParseFiles(
		"./templates/registration.html"))

	mainpageT = template.Must(template.ParseFiles(
		"./templates/mainpage.html",
		"./templates/nav.html"))

	customerTasksT = template.Must(template.ParseFiles(
		"./templates/customerTasks.html",
		"./templates/nav.html"))

	freelancerTasksT = template.Must(template.ParseFiles(
		"./templates/freelancerTasks.html",
		"./templates/nav.html"))

	createTaskT = template.Must(template.ParseFiles(
		"./templates/createTask.html",
		"./templates/nav.html"))

	myContractsT = template.Must(template.ParseFiles(
		"./templates/myContracts.html",
		"./templates/nav.html"))

	seeProposalsT = template.Must(template.ParseFiles(
		"./templates/seeProposals.html",
		"./templates/nav.html"))

	customerAccountT = template.Must(template.ParseFiles(
		"./templates/customerAccount.html",
		"./templates/nav.html"))

	freelancerAccountT = template.Must(template.ParseFiles(
		"./templates/freelancerAccount.html",
		"./templates/nav.html"))
}
