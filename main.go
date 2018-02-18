package main

import (
	"log"
	"net/http"
	"time"
)

const (
	REQUEST_DOCTOR  int8 = 0
	REQUEST_PLADEMA int8 = 1
	REQUEST_ADMIN   int8 = 2

	VALID_DATA_ENTRY = "valid_data_entry"

	USER_CREATED_SUCCESS          = "user created success"
	LOGOUT_SUCCESS                = "logout success"
	CREATE_FOLDER_SUCCESS         = "folder created successfully"
	DELETE_FOLDER_SUCCESS         = "folder deleted successfully"
	DELETE_USER_SUCCESS           = "user deleted successfully"
	RENAME_FOLDER_SUCCESS         = "renaming folder successfully"
	ERROR_USER_ALREADY_LOGUED     = "user already logged"
	ERROR_BAD_FORMED_EMAIL        = "error bad formed email"
	ERROR_BAD_FORMED_PASSWORD     = "error bad formed password"
	ERROR_BAD_FORMED_NAME         = "the name cannot be empty"
	ERROR_BAD_FORMED_TOKEN        = "the token cannot be empty"
	ERROR_BAD_FORMED_OLD_NAME     = "the old name cannot be empty"
	ERROR_BAD_FORMED_NEW_NAME     = "the new name cannot be empty"
	ERROR_BAD_FORMED_FOLDER       = "the folder cannot be empty"
	ERROR_BAD_FORMED_FILE_NAME    = "the file name cannot be empty"
	ERROR_BAD_CATEGORY            = "the category is not valid"
	ERROR_LOGIN_CREDENTIALS       = "email or password does not match"
	ERROR_NOT_JSON_NEEDED         = "not valid json needeed"
	ERROR_NOT_EXISTING_USER       = "the user does not exists"
	ERROR_NOT_VALID_TOKEN         = "the token provided is not valid"
	ERROR_EMAIL_ALREADY_EXISTS    = "the email is already registered in the database"
	ERROR_INSERT_NEW_DOCTOR       = "error when trying to insert new doctor in the database"
	ERROR_INSERT_NEW_PLADEMA      = "error when trying to insert new pladema user in the database"
	ERROR_NEW_ADMIN               = "there only can be one admin"
	ERROR_USER_NOT_IN_DB          = "the user data is not in the database"
	ERROR_MISSMATCH_USER_PASSWORD = "the password or the user are not valid"
	ERROR_NOT_LOGUED_USER         = "the user is not logued on"
	ERROR_SERVER                  = "error server"
	ERROR_ADMIN_NOT_LOGUED        = "the admin is not logued"
	ERROR_REQUIRE_LOGIN_AGAIN     = "login timeout"
	ERROR_FOLDER_WITH_OPEN_FILE   = "the folder contains one opened file"
	ERROR_FILE_ALREADY_EXISTS     = "the file already exists"
)

type Pair struct {
	TimeLogIn time.Time
	Email     string
}

var TimeoutLogin = 600
var OpenedFiles = make(map[string][]string)
var LogedUsers = make(map[string]*Pair)
var SigningKeyAdmin = []byte("Sup3rS3cr374dm1n")
var SigningKeyPladema = []byte("S3cr37Sup3rPl4d3m4")
var SigningKeyDoctor = []byte("Sup4S1cr1tD0ct0r")

func main() {
	// cd $GOPATH/src/github/nicolasmaurotorres/rest-api
	// go build && ./rest-api
	router := NewRouter()
	log.Fatal(http.ListenAndServe(":8001", router))
}
