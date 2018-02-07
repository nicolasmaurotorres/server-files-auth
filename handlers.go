package main

import (
	"encoding/json"
	"net/http"
	"time"

	jwt "github.com/dgrijalva/jwt-go"
)

// Precondicion: el token que se pasa es valido, contiene el user.Email y user.Password y son datos validos
func GenerateToken(user UserLoginRequest, cat int) (JwtToken, bool) {
	var key []byte
	var password string
	switch cat {
	case REQUEST_DOCTOR:
		key = SigningKeyDoctor
		password = "passwordDoctor" // TODO: buscar en MongoDB la clave del usuario doctor
		// TODO: controlar que no este logueado previamente el usuario, por el tema de que un proceso de adueña de un archivo y no lo libera
		break
	case REQUEST_PLADEMA:
		key = SigningKeyPladema
		password = "passwordPladema" // TODO: buscar en MongoDB la clave del usuario pladema
		// TODO: controlar que no este logueado previamente el usuario, por el tema de que un proceso de adueña de un archivo y no lo libera
		break
	case REQUEST_ADMIN:
		if !AdminLoguedIn {
			key = SigningKeyAdmin
			password = "passwordAdmin" // TODO: buscar en MongoDB la clave del usuario administrador
			AdminLoguedIn = true       // 1 solo admin puede haber logueado en el sistema
		} else {
			return JwtToken{Token: ERROR_USER_ALREADY_LOGUED}, false
		}
		break
	default:
		return JwtToken{Token: ERROR_SERVER}, false
	}
	//TODO: control de pass y user que sean iguales
	if password == "" {
		password = "pepe"
	}
	//controlo que el usuario NO este logueado previamente
	_, inMap := LogedUsers[user.Email]
	if inMap {
		return JwtToken{Token: ERROR_USER_ALREADY_LOGUED}, false
	}
	//genero el token del usuario, lo guardo en la hash, y lo devuelvo
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"username": user.Email,
		"password": user.Password,
	})
	tokenString, _ := token.SignedString(key)
	var toReturn JwtToken
	toReturn.Token = tokenString
	LogedUsers[tokenString] = time.Now()
	return toReturn, true
}

func LoginPerson(cat int, r *http.Request) (JwtToken, bool) {
	var user UserLoginRequest
	tokenReponse := ValidateLoginJSONRequest(r)
	if tokenReponse.Token == VALID_DATA_ENTRY {
		// no hay error ni en el token requerido, ni en los datos proporcionados, ni en el logueo
		return GenerateToken(user, cat)
	}
	return tokenReponse, false
}

// returns the error of login
func GetLoginError(token JwtToken) ([]byte, int) {
	var e Response
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
func LoginDoctor(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json; charset=UTF-8")
	token, valid := LoginPerson(REQUEST_DOCTOR, r)
	if valid == true {
		w.WriteHeader(http.StatusOK)
		tokenJSON, _ := json.Marshal(token)
		w.Write(tokenJSON)
	} else {
		exceptionJSON, status := GetLoginError(token)
		w.WriteHeader(status)
		w.Write(exceptionJSON)
	}
}

// returns a auth token as pladema user
func LoginPladema(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json; charset=UTF-8")
	token, valid := LoginPerson(REQUEST_PLADEMA, r)
	if valid == true {
		w.WriteHeader(http.StatusOK)
		tokenJSON, _ := json.Marshal(token)
		w.Write(tokenJSON)
	} else {
		exceptionJSON, status := GetLoginError(token)
		w.WriteHeader(status)
		w.Write(exceptionJSON)
	}
}

// returns a auth token as admin user
func LoginAdmin(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json; charset=UTF-8")
	token, valid := LoginPerson(REQUEST_ADMIN, r)
	if valid == true {
		w.WriteHeader(http.StatusOK)
		tokenJSON, _ := json.Marshal(token)
		w.Write(tokenJSON)

	} else {
		exceptionJSON, status := GetLoginError(token)
		w.WriteHeader(status)
		w.Write(exceptionJSON)
	}
}

// add a new user with the data on a json
func AddUser(w http.ResponseWriter, r *http.Request) {}

// delete some valid user with the data on a json
func DelUser(w http.ResponseWriter, r *http.Request) {}

// change name or email of a valid user
func EditUser(w http.ResponseWriter, r *http.Request) {}

// add a file to visualize
func AddFile(w http.ResponseWriter, r *http.Request) {}

// remove a file of the hashtable of opened files
func DelFile(w http.ResponseWriter, r *http.Request) {}

// returns all files to visualize
func AllFiles(w http.ResponseWriter, req *http.Request) {}

// add to the hashtable the file that is opened
func OpenedFile(w http.ResponseWriter, r *http.Request) {}

// remove a file of the hashtable of opened files
func CloseFile(w http.ResponseWriter, r *http.Request) {}

// search in the files by some filters on json object and return a json object with the result
func SearchFiles(w http.ResponseWriter, r *http.Request) {}

// logout a pladema user
func LogoutPladema(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json; charset=UTF-8")
	valid, tokenString := ValidateLogoutJSONRequest(r, REQUEST_PLADEMA)
	var response Response
	if valid {
		validToken, errorMessage := IsValidToken(tokenString, REQUEST_PLADEMA)
		if validToken {
			_, inMap := LogedUsers[tokenString]
			if inMap {
				w.WriteHeader(http.StatusOK)
				delete(LogedUsers, tokenString) // elimino al usuario
				response.Message = LOGOUT_SUCCESS
				response.Status = http.StatusOK
				responseJSON, _ := json.Marshal(response)
				w.Write(responseJSON)
			} else {
				// el usuario que esta intentando desloguearse, no esta logueado, es un error, es un token valido pero no esta
				// en la hash, tiramos otro mensaje de error para despistar :V
				w.WriteHeader(http.StatusForbidden)
				response.Message = ERROR_NOT_VALID_TOKEN
				response.Status = http.StatusForbidden
				responseJSON, _ := json.Marshal(response)
				w.Write(responseJSON)
			}
		} else {
			// no es un token valido
			w.WriteHeader(http.StatusForbidden)
			response.Message = errorMessage
			response.Status = http.StatusForbidden
			responseJSON, _ := json.Marshal(response)
			w.Write(responseJSON)

		}
	} else {
		// no se encontro el campo "token" en el json que enviaron
		w.WriteHeader(http.StatusBadRequest)
		response.Message = ERROR_NOT_JSON_NEEDED
		response.Status = http.StatusBadRequest
		responseJSON, _ := json.Marshal(response)
		w.Write(responseJSON)

	}
}

// logout a doctor user
func LogoutDoctor(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json; charset=UTF-8")
	valid, tokenString := ValidateLogoutJSONRequest(r, REQUEST_DOCTOR)
	var response Response
	if valid {
		validToken, errorMessage := IsValidToken(tokenString, REQUEST_DOCTOR)
		if validToken {
			_, inMap := LogedUsers[tokenString]
			if inMap {
				w.WriteHeader(http.StatusOK)
				delete(LogedUsers, tokenString) // elimino al usuario
				response.Message = LOGOUT_SUCCESS
				response.Status = http.StatusOK
				responseJSON, _ := json.Marshal(response)
				w.Write(responseJSON)

			} else {
				// el usuario que esta intentando desloguearse, no esta logueado, es un error
				w.WriteHeader(http.StatusForbidden)
				response.Message = ERROR_NOT_VALID_TOKEN
				response.Status = http.StatusForbidden
				responseJSON, _ := json.Marshal(response)
				w.Write(responseJSON)
			}
		} else {
			// no es un token valido
			w.WriteHeader(http.StatusForbidden)
			response.Message = errorMessage
			response.Status = http.StatusForbidden
			responseJSON, _ := json.Marshal(response)
			w.Write(responseJSON)
		}
	} else {
		// no se encontro el campo "token" en el json que enviaron
		w.WriteHeader(http.StatusBadRequest)
		response.Message = ERROR_NOT_JSON_NEEDED
		response.Status = http.StatusBadRequest
		responseJSON, _ := json.Marshal(response)
		w.Write(responseJSON)
	}
}

// logout a admin user
func LogoutAdmin(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json; charset=UTF-8")
	valid, tokenString := ValidateLogoutJSONRequest(r, REQUEST_ADMIN)
	var response Response
	if valid {
		validToken, errorMessage := IsValidToken(tokenString, REQUEST_ADMIN)
		if validToken {
			_, inMap := LogedUsers[tokenString]
			if inMap {
				w.WriteHeader(http.StatusOK)
				AdminLoguedIn = false
				delete(LogedUsers, tokenString) // elimino al usuario
				response.Message = LOGOUT_SUCCESS
				response.Status = http.StatusOK
				responseJSON, _ := json.Marshal(response)
				w.Write(responseJSON)
			} else {
				// el usuario que esta intentando desloguearse, no esta logueado, es un error
				w.WriteHeader(http.StatusForbidden)
				response.Message = ERROR_NOT_VALID_TOKEN
				response.Status = http.StatusForbidden
				responseJSON, _ := json.Marshal(response)
				w.Write(responseJSON)
			}
		} else {
			// no es un token valido
			w.WriteHeader(http.StatusForbidden)
			response.Message = errorMessage
			response.Status = http.StatusForbidden
			responseJSON, _ := json.Marshal(response)
			w.Write(responseJSON)
		}
	} else {
		// no se encontro el campo "token" en el json que enviaron
		w.WriteHeader(http.StatusBadRequest)
		response.Message = ERROR_NOT_JSON_NEEDED
		response.Status = http.StatusBadRequest
		responseJSON, _ := json.Marshal(response)
		w.Write(responseJSON)
	}
}
