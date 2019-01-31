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
		"RouteLogin", // ok
		"POST",
		"/login",
		RouteLogin,
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
	Route{
		"AdminViewUsers",
		"POST",
		"/admin/view/users",
		AdminViewUsers,
	},
	//Specialist and technician common operations
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
		"CopyFolder", //falta test Specialist
		"POST",
		"/copy/folder/to/location",
		CopyFolder,
	},
	Route{
		"SpecialistAddFile",
		"POST",
		"/add/file",
		AddFile,
	},
	Route{
		"SpecialistDeleteFile",
		"POST",
		"/delete/file",
		DeleteFile,
	},
	//Specialist specific operations
	Route{
		"SpecialistGetFiles", //ok
		"POST",
		"/specialist/get/files",
		SpecialistGetFiles,
	},
	Route{
		"openFile", // ok
		"POST",
		"/specialist/open/file",
		SpecialistOpenFile,
	},
	Route{
		"closeFile", // ok
		"POST",
		"/specialist/close/file",
		SpecialistCloseFile,
	},
	//Technician specific operations
	Route{
		"technicianSearchFiles", // ok
		"POST",
		"/technician/search/files",
		TechnicianSearchFiles,
	},
	Route{
		"technicianGetAllEmails", // ok
		"POST",
		"/technician/get/emails",
		TechnicianGetEmails,
	},
	Route{
		"TechnicianGetFile",
		"POST",
		"/technician/get/file",
		TechnicianGetFile,
	},
	// Common Operations
	Route{
		"logout", // ok
		"POST",
		"/logout",
		Logout,
	},
}
