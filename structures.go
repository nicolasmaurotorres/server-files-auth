package main

import "time"

const (
	REQUEST_LOGIN_DOCTOR  = 1
	REQUEST_LOGIN_PLADEMA = 2
	REQUEST_LOGIN_ADMIN   = 3
)

type jwtToken struct {
	token string `json:"token"`
}

type pair struct {
	date  time.Time
	token jwtToken
}

type directory struct {
	path  string   `json:"path"`
	files []string `json:"files"`
}

type directorys []directory

type userLoginRequest struct {
	email    string `json:"email"`
	password string `json:"password"`
}

type exception struct {
	message string `json:"message"`
}

var timeoutLogin = 600

var openedFiles = make(map[string][]string)
var logedUsers = make(map[string]pair)
var signingKeyAdmin = []byte("Sup3rS3cr374dm1n")
var signingKeyPladema = []byte("S3cr37Sup3rPl4d3m4")
var signingKeyDoctor = []byte("Sup4S1cr1tD0ct0r")
var adminLoguedIn = false
