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
		LoginDoctor,
	},
	Route{
		"loginPladema",
		"POST",
		"/login/pladema",
		LoginPladema,
	},
	Route{
		"loginAdmin",
		"POST",
		"/login/admin",
		LoginAdmin,
	},
	Route{
		"addUser",
		"POST",
		"/admin/add",
		AddUser,
	},
	Route{
		"delUser",
		"POST",
		"/admin/del",
		DelUser,
	},
	Route{
		"editUser",
		"POST",
		"/admin/edit",
		EditUser,
	},
	Route{
		"addFile",
		"POST",
		"/doctor/add/file",
		AddFile,
	},
	Route{
		"delFile",
		"POST",
		"/doctor/del/file",
		DelFile,
	},

	Route{
		"allFiles",
		"POST",
		"/doctor/files",
		AllFiles,
	},
	Route{
		"openedFile",
		"POST",
		"/doctor/open/file",
		OpenedFile,
	},
	Route{
		"closeFile",
		"POST",
		"/doctor/close/file",
		CloseFile,
	},
	Route{
		"searchFiles",
		"POST",
		"/login/admin",
		SearchFiles,
	},
	Route{
		"logoutDoctor",
		"POST",
		"/logout/doctor",
		LogoutDoctor,
	},
	Route{
		"logoutPladema",
		"POST",
		"/logout/pladema",
		LogoutPladema,
	},
	Route{
		"logoutAdmin",
		"POST",
		"/logout/admin",
		LogoutAdmin,
	},
}
