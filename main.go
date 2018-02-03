package main

import (
	"log"
	"net/http"
	"time"
)

const (
	REQUEST_LOGIN_DOCTOR  = 1
	REQUEST_LOGIN_PLADEMA = 2
	REQUEST_LOGIN_ADMIN   = 3

	VALID_DATA_ENTRY = "valid_data_entry"

	ERROR_USER_ALREADY_LOGUED = "user already logged"
	ERROR_BAD_FORMED_EMAIL    = "error_bad_formed_email"
	ERROR_BAD_FORMED_PASSWORD = "error_bad_formed_password"
	ERROR_LOGIN_CREDENTIALS   = "email or password does not match"
	ERROR_NOT_JSON_NEEDED     = "not_valid_json"
	ERROR_NOT_EXISTING_USER   = "the user does not exists"
	ERROR_SERVER              = "error server"
)

type jwtToken struct {
	Token string `json:"token"`
}

type pair struct {
	date  time.Time
	token jwtToken
}

type directory struct {
	Path  string   `json:"path"`
	Files []string `json:"files"`
}

type directorys []directory

type userLoginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type exception struct {
	Message string `json:"message"`
	Status  int    `json:"status"`
}

var timeoutLogin = 600

var openedFiles = make(map[string][]string)
var logedUsers = make(map[string]pair)
var signingKeyAdmin = []byte("Sup3rS3cr374dm1n")
var signingKeyPladema = []byte("S3cr37Sup3rPl4d3m4")
var signingKeyDoctor = []byte("Sup4S1cr1tD0ct0r")
var adminLoguedIn = false

func main() {
	// go build && ./rest-api
	router := NewRouter()
	log.Fatal(http.ListenAndServe(":8001", router))
}
