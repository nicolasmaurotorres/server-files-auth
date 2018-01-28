package main

import (
	"log"
	"net/http"
	"time"
)

const (
	categoryDoctor  = 0
	categoryPladema = 1
	categoryAdmin   = 2
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
type user struct {
	category   int
	name       string
	password   string
	email      string
	directorys []directory
}

type directorys []directory

type exception struct {
	message string `json:"message"`
}

var timeoutLogin = 600

var openedFiles = make(map[string][]string)
var logedUsers = make(map[string]pair)

func isValidToken(token jwtToken, cat int) bool {
	return false
}

func main() {
	router := NewRouter()

	log.Fatal(http.ListenAndServe(":8001", router))
}
