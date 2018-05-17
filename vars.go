package main

import (
	"html/template"

	"github.com/gorilla/securecookie"
)

const cookieName = ""

var hashKey = []byte("")
var sc = securecookie.New(hashKey, nil)

var loginT *template.Template
var registrationT *template.Template
var mainpageT *template.Template
var customerTasksT *template.Template
var freelancerTasksT *template.Template
var createTaskT *template.Template
var myContractsT *template.Template
var seeProposalsT *template.Template
var customerAccountT *template.Template
var freelancerAccountT *template.Template
