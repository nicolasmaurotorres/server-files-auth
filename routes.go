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
		"loginDoctor", // ok
		"POST",
		"/login/doctor",
		LoginDoctor,
	},
	Route{
		"loginPladema", // ok
		"POST",
		"/login/pladema",
		LoginPladema,
	},
	Route{
		"loginAdmin", // ok
		"POST",
		"/login/admin",
		LoginAdmin,
	},
	Route{
		"addUser", // ok
		"POST",
		"/admin/add",
		AddUser,
	},
	Route{
		"delUser", //ok
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
		"addFolder", // ok
		"POST",
		"/add/folder",
		AddFolder,
	},
	Route{
		"delFolder", // ok
		"POST",
		"/del/folder",
		DelFolder,
	},
	Route{
		"renameFolder",
		"POST",
		"/rename/folder",
		RenameFolder,
	},
	Route{
		"renameFile",
		"POST",
		"/rename/file",
		RenameFile,
	},
	Route{
		"moveFileToFolder",
		"POST",
		"/move/file/folder",
		MoveFileToFolder,
	},
	Route{
		"addFile", // tanto para doctor como para pladema
		"POST",
		"/add/file",
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
		"openFile",
		"POST",
		"/doctor/open/file",
		OpenFile,
	},
	Route{
		"closeFile",
		"POST",
		"/doctor/close/file",
		CloseFile,
	},
	Route{
		"searchFiles", // operacion para usuario pladema
		"POST",
		"/login/admin",
		SearchFiles,
	},
	Route{ // ok
		"logout",
		"POST",
		"/logout",
		Logout,
	},
}
