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
		"/login/{user}/pass/{pass}",
		loginDoctor,
	},
	Route{
		"loginPladema",
		"POST",
		"/login/{pladema}/pass/{pass}",
		loginPladema,
	},
	Route{
		"loginAdmin",
		"POST",
		"/login/{admin}/pass/{pass}",
		loginAdmin,
	},
	Route{
		"addUser",
		"POST",
		"/admin/add/{auth}",
		addUser,
	},
	Route{
		"delUser",
		"POST",
		"/admin/del/{auth}",
		delUser,
	},
	Route{
		"editUser",
		"POST",
		"/admin/edit/{auth}",
		editUser,
	},
	Route{
		"addFile",
		"POST",
		"/{doctor}/add/{path}/{filename}/{auth}",
		addFile,
	},
	Route{
		"delFile",
		"POST",
		"/{doctor}/del/{path}/{filename}/{auth}",
		delFile,
	},

	Route{
		"allFiles",
		"POST",
		"/{doctor}/files/{auth}",
		allFiles,
	},
	Route{
		"openedFile",
		"POST",
		"/{doctor}/open/{filename}/{auth}",
		openedFile,
	},
	Route{
		"closeFile",
		"POST",
		"/{doctor}/close/{filename}/{auth}",
		closeFile,
	},
	Route{
		"searchFiles",
		"POST",
		"/login/{admin}/pass/{pass}",
		searchFiles,
	},
}
