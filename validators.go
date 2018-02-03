package main

import (
	"encoding/json"
	"net/http"

	"github.com/asaskevich/govalidator"
)

func validateFieldsRequest(user userLoginRequest) jwtToken {
	// validate the email is not null
	if !govalidator.IsEmail(user.Email) {
		return jwtToken{Token: ERROR_BAD_FORMED_EMAIL}
	}
	// validate the name is not empty or missing
	if govalidator.IsNull(user.Password) {
		return jwtToken{Token: ERROR_BAD_FORMED_PASSWORD}
	}
	return jwtToken{Token: VALID_DATA_ENTRY}
}

func validateLoginRequest(r *http.Request) jwtToken {
	var user userLoginRequest
	err := json.NewDecoder(r.Body).Decode(&user)
	defer r.Body.Close()
	if err == nil { // no hay errores de json mal formado
		return validateFieldsRequest(user)
	}
	return jwtToken{Token: ERROR_NOT_JSON_NEEDED}
}
