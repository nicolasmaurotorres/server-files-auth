package main

import (
	"encoding/json"
	"fmt"
	"net/http"

	jwt "github.com/dgrijalva/jwt-go"
)

func validateLoginRequest(r *http.Request, v InputValidation) error {
	if err := json.NewDecoder(r.Body).Decode(v); err != nil {
		return err
	}
	defer r.Body.Close()
	// peform validation on the InputValidation implementation
	return v.Validate(r)
}

func generateToken(cat int, ) jwtToken {
	var key []byte

	switch cat {
	case categoryDoctor:
		key = signingKeyDoctor
		//TODO buscar en MongoDB la clave del usuario doctor
		break
	case categoryPladema:
		key = signingKeyPladema
		//TODO buscar en MongoDB la clave del usuario pladema
		break
	case categoryAdmin:
		if !adminLoguedIn {
			key = signingKeyAdmin
			//TODO buscar en MongoDB la clave del usuario
		} else {
			panic("DOUBLE ADMIN LOGUED")
		}
		break
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"username": user.Username,
		"password": user.Password,
	})
	tokenString, error := token.SignedString(key)
	if error != nil {
		fmt.Println(error)
	}
	json.NewEncoder(w).Encode(JwtToken{Token: tokenString})

	return toReturn
}

func isValidToken(token jwtToken, cat int) bool {
	return false
}

func loginPerson(cat int) jwtToken{
	user userLoginRequest
	err := validateLoginRequest(r,user)
	if (err == nil) {
		token := generateToken(user,REQUEST_LOGIN_DOCTOR)
	} else {
		// la entrada no es valida, por falta de email, por falta de contraseña,
		w.WriteHeader(http.StatusBadRequest) // falla en los parametros de entrada
		w.WriteHeader(http.StatusForbidden) // mala contraseña o email
	}
	return token
}


// returns a auth token as doctor user
func loginDoctor(w http.ResponseWriter, r *http.Request) {


	
}

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

	w.Header().Set("Content-Type", "application/json; charset=UTF-8")
	w.WriteHeader(http.StatusOK)

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
