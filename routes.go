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
	//Logins
	Route{
		"DoctorLogin", // ok
		"POST",
		"/doctor/login",
		DoctorLogin,
	},
	Route{
		"PlademaLogin", // ok
		"POST",
		"/pladema/login",
		PlademaLogin,
	},
	Route{
		"AdminLogin", // ok
		"POST",
		"/admin/login",
		AdminLogin,
	},
	//admin operations
	Route{
		"AdminAddUser", // ok
		"POST",
		"/admin/add/user",
		AdminAddUser,
	},
	Route{
		"AdminDeleteUser", //ok
		"POST",
		"/admin/delete/user",
		AdminDeleteUser,
	},
	Route{ //TODO:
		"AdminEditUser",
		"POST",
		"/admin/edit/user",
		AdminEditUser,
	},
	//Doctor operations
	Route{
		"DoctorAddFolder", // ok
		"POST",
		"/doctor/add/folder",
		DoctorAddFolder,
	},
	Route{
		"DoctorDeleteFolder", // ok
		"POST",
		"/doctor/delete/folder",
		DoctorDeleteFolder,
	},
	Route{
		"DoctorRenameFolder", // ok
		"POST",
		"/doctor/rename/folder",
		DoctorRenameFolder,
	},
	Route{
		"DoctorRenameFile", // ok
		"POST",
		"/doctor/rename/file",
		DoctorRenameFile,
	},
	Route{
		"DoctorChangeFileFolder",
		"POST",
		"doctor/change/filefolder",
		DoctorChangeFileFolder,
	},
	Route{
		"DoctorAddFile", // ok
		"POST",
		"/doctor/add/file",
		DoctorAddFile,
	},
	Route{
		"DoctorDeleteFile", // ok
		"POST",
		"/doctor/delete/file",
		DoctorDeleteFile,
	},
	Route{
		"DoctorGetFiles", //ok
		"POST",
		"/doctor/get/files",
		DoctorGetFiles,
	},
	Route{
		"openFile", // ok
		"POST",
		"/doctor/open/file",
		DoctorOpenFile,
	},
	Route{
		"closeFile", // ok
		"POST",
		"/doctor/close/file",
		DoctorCloseFile,
	},
	//Pladema operations
	Route{
		"PlademaAddFile", // tanto para doctor como para pladema
		"POST",
		"/pladema/add/file",
		PlademaAddFile,
	},
	Route{
		"searchFiles", // ok
		"POST",
		"/pladema/search/files",
		PlademaSearchFiles,
	},
	// Common Operations
	Route{
		"logout", // ok
		"POST",
		"/logout",
		Logout,
	},
}
