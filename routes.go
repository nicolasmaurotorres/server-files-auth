package main

import (
	"net/http"
)

// Route contains all the info of the route off the api
type Route struct {
	Name        string
	Method      string
	Pattern     string
	HandlerFunc http.HandlerFunc
}

// Routes type to contain all the routes of the api
type Routes []Route

// variable to contain all the routes of the api
var routes = Routes{
	Route{
		"loginDoctor",
		"POST",
		"/login/doctor",
		loginDoctor,
	},
	Route{
		"loginPladema",
		"POST",
		"/login/pladema",
		loginPladema,
	},
	Route{
		"loginAdmin",
		"POST",
		"/login/admin",
		loginAdmin,
	},
	Route{
		"addUser",
		"POST",
		"/admin/add",
		addUser,
	},
	Route{
		"delUser",
		"POST",
		"/admin/del",
		delUser,
	},
	Route{
		"editUser",
		"POST",
		"/admin/edit",
		editUser,
	},
	Route{
		"addFile",
		"POST",
		"/doctor/add/file",
		addFile,
	},
	Route{
		"delFile",
		"POST",
		"/doctor/del/file",
		delFile,
	},

	Route{
		"allFiles",
		"POST",
		"/doctor/files",
		allFiles,
	},
	Route{
		"openedFile",
		"POST",
		"/doctor/open/file",
		openedFile,
	},
	Route{
		"closeFile",
		"POST",
		"/doctor/close/file",
		closeFile,
	},
	Route{
		"searchFiles",
		"POST",
		"/login/admin",
		searchFiles,
	},
}
