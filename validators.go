package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"

	"github.com/asaskevich/govalidator"
)

type UserLoginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

func GetLoginJSONRequest(r *http.Request, cat int8) (UserLoginRequest, error) {
	var user UserLoginRequest
	err := json.NewDecoder(r.Body).Decode(&user)
	defer r.Body.Close()
	if err == nil { // no hay errores de json mal formado
		if cat != REQUEST_ADMIN {
			if !govalidator.IsEmail(user.Email) {
				return user, errors.New(ERROR_BAD_FORMED_EMAIL)
			}
		} else {
			if govalidator.IsNull(user.Email) {
				return user, errors.New(ERROR_BAD_FORMED_PASSWORD)
			}
		}
		// validate the name is not empty or missing
		if govalidator.IsNull(user.Password) {
			return user, errors.New(ERROR_BAD_FORMED_PASSWORD)
		}
		return user, nil
	}
	return user, errors.New(ERROR_NOT_JSON_NEEDED)
}

type NewUserRequest struct {
	Token    string `json:"token"`
	Name     string `json:"name"`
	Email    string `json:"email"`
	Password string `json:"password"`
	Category int8   `json:"category"` // 0 doctor, 1 pladema
}

// GetNewUserJSONRequest returns a user with valid data
func GetNewUserJSONRequest(r *http.Request) (NewUserRequest, error) {
	var newUserRequest NewUserRequest
	err := json.NewDecoder(r.Body).Decode(&newUserRequest)
	defer r.Body.Close()
	if err != nil {
		fmt.Println(err)
		//the json input is valid but we have to check the data values
		return newUserRequest, errors.New(ERROR_NOT_JSON_NEEDED)
	}
	valid, errorMessage := IsValidToken(newUserRequest.Token)
	if !valid {
		return newUserRequest, errors.New(errorMessage.Error())
	}

	if !govalidator.IsEmail(newUserRequest.Email) {
		return newUserRequest, errors.New(ERROR_BAD_FORMED_EMAIL)
	}

	if ExistsEmail(newUserRequest.Email) {
		return newUserRequest, errors.New(ERROR_EMAIL_ALREADY_EXISTS)
	}

	if govalidator.IsNull(newUserRequest.Password) {
		return newUserRequest, errors.New(ERROR_BAD_FORMED_PASSWORD)
	}

	if govalidator.IsNull(newUserRequest.Name) {
		return newUserRequest, errors.New(ERROR_BAD_FORMED_NAME)
	}

	if !(newUserRequest.Category == REQUEST_DOCTOR || newUserRequest.Category == REQUEST_PLADEMA) {
		return newUserRequest, errors.New(ERROR_BAD_CATEGORY)
	}
	return newUserRequest, nil
}

type JwtToken struct {
	Token string `json:"token"`
}

func GetTokenStringFromLogoutRequest(r *http.Request) (bool, string) {
	var userLogoutRequest JwtToken
	err := json.NewDecoder(r.Body).Decode(&userLogoutRequest)
	defer r.Body.Close()
	if err == nil {
		return true, userLogoutRequest.Token
	}
	return false, ERROR_NOT_JSON_NEEDED
}

// IsValidToken returns if a token is valid or not, depends on the category of the token
func IsValidToken(tokenString string) (bool, error) {
	_, inMap := LogedUsers[tokenString]
	if inMap {
		return true, nil
	}
	return false, errors.New(ERROR_NOT_LOGUED_USER)
}
