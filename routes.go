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
		"renameFolder", // ok
		"POST",
		"/doctor/rename/folder",
		RenameFolder,
	},
	Route{
		"renameFile", // ok
		"POST",
		"/doctor/rename/file",
		RenameFileDoctor,
	},
	Route{
		"moveFileToFolder",
		"POST",
		"/move/file/folder",
		MoveFileToFolder,
	},
	Route{
		"addFileDoctor", // ok
		"POST",
		"/doctor/add/file",
		AddFileDoctor,
	},
	Route{
		"addFilePladema", // tanto para doctor como para pladema
		"POST",
		"/doctor/add/file",
		AddFilePladema,
	},
	Route{
		"delFile", // ok
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
	Route{
		"logout", // ok
		"POST",
		"/logout",
		Logout,
	},
}
