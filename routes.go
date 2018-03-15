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
	Route{
		"AdminEditUser", //ok
		"POST",
		"/admin/edit/user",
		AdminEditUser,
	},
	//Doctor and pladema common operations
	Route{
		"AddFolder",
		"POST",
		"/add/folder",
		AddFolder,
	},
	Route{
		"DeleteFolder",
		"POST",
		"/delete/folder",
		DeleteFolder,
	},
	Route{
		"RenameFolder",
		"POST",
		"/rename/folder",
		RenameFolder,
	},
	Route{
		"RenameFile",
		"POST",
		"/rename/file",
		RenameFile,
	},
	Route{
		"CopyFile", // ok
		"POST",
		"/copy/file/to/location",
		CopyFile,
	},
	Route{
		"CopyFolder", //falta test doctor
		"POST",
		"/copy/folder/to/location",
		CopyFolder,
	},
	Route{
		"DoctorAddFile",
		"POST",
		"/add/file",
		AddFile,
	},
	Route{
		"DoctorDeleteFile",
		"POST",
		"/delete/file",
		DeleteFile,
	},
	//Doctor specific operations
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
	//Pladema specific operations
	Route{
		"plademaSearchFiles", // ok
		"POST",
		"/pladema/search/files",
		PlademaSearchFiles,
	},
	Route{
		"plademaGetAllEmails", // ok
		"POST",
		"/pladema/get/emails",
		PlademaGetEmails,
	},
	Route{
		"plademaGetFile",
		"GET",
		"/pladema/get/file",
		PlademaGetFile,
	},
	// Common Operations
	Route{
		"logout", // ok
		"POST",
		"/logout",
		Logout,
	},
}
