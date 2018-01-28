package main

import (
	"encoding/json"
	"fmt"
	"net/http"

	jwt "github.com/dgrijalva/jwt-go"
)

var signingKeyAdmin = []byte("Sup3rS3cr374dm1n")
var signingKeyPladema = []byte("S3cr37Sup3rPl4d3m4")
var signingKeyDoctor = []byte("Sup4S1cr1tD0ct0r")

func generateToken(cat int) jwtToken {
	return nil
}

// returns a auth token as doctor user
func loginDoctor(w http.ResponseWriter, r *http.Request) {}

// returns a auth token as pladema user
func loginPladema(w http.ResponseWriter, r *http.Request) {}

// returns a auth token as admin user
func loginAdmin(w http.ResponseWriter, r *http.Request) {}

// add a new user with the data on a json
func addUser(w http.ResponseWriter, r *http.Request) {}

// delete some valid user with the data on a json
func delUser(w http.ResponseWriter, r *http.Request) {}

// change name or email of a valid user
func editUser(w http.ResponseWriter, r *http.Request) {}

// add a file to visualize
func addFile(w http.ResponseWriter, r *http.Request) {}

// remove a file of the hashtable of opened files
func delFile(w http.ResponseWriter, r *http.Request) {}

// returns all files to visualize
func allFiles(w http.ResponseWriter, req *http.Request) {
	var user User
	_ = json.NewDecoder(req.Body).Decode(&user)
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"username": user.Username,
		"password": user.Password,
	})
	tokenString, error := token.SignedString([]byte("secret"))
	if error != nil {
		fmt.Println(error)
	}
	json.NewEncoder(w).Encode(JwtToken{Token: tokenString})
}

// add to the hashtable the file that is opened
func openedFile(w http.ResponseWriter, r *http.Request) {}

// remove a file of the hashtable of opened files
func closeFile(w http.ResponseWriter, r *http.Request) {}

// search in the files by some filters on json object and return a json object with the result
func searchFiles(w http.ResponseWriter, r *http.Request) {}
