package main

import (
	"encoding/json"
	"net/http"

	jwt "github.com/dgrijalva/jwt-go"
)

func isValidToken(token jwtToken, cat int) bool {
	return false
}

// Precondicion: el token que se pasa es valido, contiene el user.Email y user.Password y son datos validos
func generateToken(user userLoginRequest, cat int) (jwtToken, bool) {
	var key []byte
	var password string
	switch cat {
	case REQUEST_LOGIN_DOCTOR:
		key = signingKeyDoctor
		password = "passwordDoctor" // TODO: buscar en MongoDB la clave del usuario doctor
		// TODO: controlar que : exista el usuario // coincidan el usuario y contraseña
		// TODO: controlar que no este logueado previamente el usuario, por el tema de que un proceso de adueña de un archivo y no lo libera
		break
	case REQUEST_LOGIN_PLADEMA:
		key = signingKeyPladema
		password = "passwordPladema" // TODO: buscar en MongoDB la clave del usuario pladema
		// TODO: controlar que : exista el usuario // coincidan el usuario y contraseña
		// TODO: controlar que no este logueado previamente el usuario, por el tema de que un proceso de adueña de un archivo y no lo libera
		break
	case REQUEST_LOGIN_ADMIN:
		if !adminLoguedIn {
			key = signingKeyAdmin
			password = "passwordAdmin" // TODO: buscar en MongoDB la clave del usuario administrador
			// TODO: controlar que : exista el usuario // coincidan el usuario y contraseña
			adminLoguedIn = true // 1 solo admin puede haber logueado en el sistema
		} else {
			return jwtToken{Token: ERROR_USER_ALREADY_LOGUED}, false
		}
		break
	default:
		return jwtToken{Token: ERROR_SERVER}, false
	}

	if password == "" {
		password = "pepe"
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"username": user.Email,
		"password": user.Password,
	})
	tokenString, _ := token.SignedString(key)
	return jwtToken{Token: tokenString}, true
}

func loginPerson(cat int, r *http.Request) (jwtToken, bool) {
	var user userLoginRequest
	tokenReponse := validateLoginRequest(r)
	if tokenReponse.Token == VALID_DATA_ENTRY {
		// no hay error ni en el token requerido, ni en los datos proporcionados, ni en el logueo
		return generateToken(user, cat)
	}
	return tokenReponse, false
}

func getLoginError(token jwtToken) ([]byte, int) {
	var e exception
	switch token.Token {
	case ERROR_BAD_FORMED_EMAIL:
		e.Status = http.StatusBadRequest
		break
	case ERROR_BAD_FORMED_PASSWORD:
		e.Status = http.StatusBadRequest
		break
	case ERROR_LOGIN_CREDENTIALS:
		e.Status = http.StatusForbidden
		break
	case ERROR_NOT_JSON_NEEDED:
		e.Status = http.StatusBadRequest
		break
	case ERROR_USER_ALREADY_LOGUED:
		e.Status = http.StatusForbidden
		break
	case ERROR_SERVER:
		e.Status = http.StatusInternalServerError
		break
	}
	e.Message = token.Token
	exceptionJSON, _ := json.Marshal(e)
	return exceptionJSON, e.Status

}

// returns a auth token as doctor user
func loginDoctor(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json; charset=UTF-8")
	token, valid := loginPerson(REQUEST_LOGIN_DOCTOR, r)
	if valid == true {
		w.WriteHeader(http.StatusOK)
		tokenJSON, _ := json.Marshal(token)
		w.Write(tokenJSON)
	} else {
		exceptionJSON, status := getLoginError(token)
		w.WriteHeader(status)
		w.Write(exceptionJSON)
	}
}

// returns a auth token as pladema user
func loginPladema(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json; charset=UTF-8")
	token, valid := loginPerson(REQUEST_LOGIN_PLADEMA, r)
	if valid == true {
		w.WriteHeader(http.StatusOK)
		tokenJSON, _ := json.Marshal(token)
		w.Write(tokenJSON)
	} else {
		exceptionJSON, status := getLoginError(token)
		w.WriteHeader(status)
		w.Write(exceptionJSON)
	}
}

// returns a auth token as admin user
func loginAdmin(w http.ResponseWriter, r *http.Request) {
	token, valid := loginPerson(REQUEST_LOGIN_ADMIN, r)
	if valid == true {
		w.WriteHeader(http.StatusOK)
		tokenJSON, _ := json.Marshal(token)
		w.Write(tokenJSON)
	} else {
		exceptionJSON, status := getLoginError(token)
		w.WriteHeader(status)
		w.Write(exceptionJSON)
	}
}

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
func allFiles(w http.ResponseWriter, req *http.Request) {}

// add to the hashtable the file that is opened
func openedFile(w http.ResponseWriter, r *http.Request) {}

// remove a file of the hashtable of opened files
func closeFile(w http.ResponseWriter, r *http.Request) {}

// search in the files by some filters on json object and return a json object with the result
func searchFiles(w http.ResponseWriter, r *http.Request) {}
