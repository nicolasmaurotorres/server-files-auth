package main

import (
	"log"
	"net/http"
	"time"
)

const (
	REQUEST_DOCTOR  = 1
	REQUEST_PLADEMA = 2
	REQUEST_ADMIN   = 3

	VALID_DATA_ENTRY = "valid_data_entry"

	LOGOUT_SUCCESS = "logout success"

	ERROR_USER_ALREADY_LOGUED = "user already logged"
	ERROR_BAD_FORMED_EMAIL    = "error bad formed email"
	ERROR_BAD_FORMED_PASSWORD = "error bad formed password"
	ERROR_LOGIN_CREDENTIALS   = "email or password does not match"
	ERROR_NOT_JSON_NEEDED     = "not valid json needeed"
	ERROR_NOT_EXISTING_USER   = "the user does not exists"
	ERROR_NOT_VALID_TOKEN     = "the token provided is not valid"
	ERROR_SERVER              = "error server"
	NOT_LOGUED_USER           = "the user is not logued on"
)

type JwtToken struct {
	Token string `json:"token"`
}

type Pair struct {
	Date  time.Time
	Token JwtToken
}

type Directory struct {
	Path  string   `json:"path"`
	Files []string `json:"files"`
}

type directorys []Directory

type UserLoginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type UserLogoutRequest struct {
	Email string `json:"email"`
	Token string `json:"token"`
}

type Response struct {
	Message string `json:"message"`
	Status  int    `json:"status"`
}

var timeoutLogin = 600

var OpenedFiles = make(map[string][]string)
var LogedUsers = make(map[string]time.Time)
var SigningKeyAdmin = []byte("Sup3rS3cr374dm1n")
var SigningKeyPladema = []byte("S3cr37Sup3rPl4d3m4")
var SigningKeyDoctor = []byte("Sup4S1cr1tD0ct0r")
var AdminLoguedIn = false

func main() {
	// go build && ./rest-api
	router := NewRouter()
	log.Fatal(http.ListenAndServe(":8001", router))
}
