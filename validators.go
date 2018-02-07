package main

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/asaskevich/govalidator"
	jwt "github.com/dgrijalva/jwt-go"
)

func ValidateFieldsLoginRequest(user UserLoginRequest) JwtToken {
	// validate the email is not null
	if !govalidator.IsEmail(user.Email) {
		return JwtToken{Token: ERROR_BAD_FORMED_EMAIL}
	}
	// validate the name is not empty or missing
	if govalidator.IsNull(user.Password) {
		return JwtToken{Token: ERROR_BAD_FORMED_PASSWORD}
	}
	return JwtToken{Token: VALID_DATA_ENTRY}
}

func ValidateLoginJSONRequest(r *http.Request) JwtToken {
	var user UserLoginRequest
	err := json.NewDecoder(r.Body).Decode(&user)
	defer r.Body.Close()
	if err == nil { // no hay errores de json mal formado
		return ValidateFieldsLoginRequest(user)
	}
	return JwtToken{Token: ERROR_NOT_JSON_NEEDED}
}

func ValidateLogoutJSONRequest(r *http.Request, cat int) (bool, string) {
	var userLogoutRequest JwtToken
	err := json.NewDecoder(r.Body).Decode(&userLogoutRequest)
	defer r.Body.Close()
	if err == nil {
		return true, userLogoutRequest.Token
	}
	return false, ERROR_NOT_JSON_NEEDED
}

func ValidDataLogoutJSONRequest(token string, cat int) (bool, string) {
	validToken, errorMessage := IsValidToken(token, cat)
	if !validToken {
		return false, errorMessage
	}
	return true, ""
}

func IsValidToken(tokenString string, cat int) (bool, string) {
	var key []byte
	switch cat {
	case REQUEST_ADMIN:
		key = SigningKeyAdmin
		break
	case REQUEST_PLADEMA:
		key = SigningKeyPladema
		break
	case REQUEST_DOCTOR:
		key = SigningKeyDoctor
		break
	default:
		return false, ERROR_SERVER
	}
	token, error := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("There was an error")
		}
		return key, nil
	})
	if error != nil {
		return false, error.Error()
	}
	if token.Valid {
		return true, ""
	}
	return false, ERROR_NOT_VALID_TOKEN
}
