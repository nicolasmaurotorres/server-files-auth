import (
	"errors"
	"net/http"
)

// error of
var (
	// ErrInvalidUUID - error when we have a UUID validation issue
	ErrInvalidEmail = errors.New("invalid email")
	// ErrInvalidName - error when we have an invalid name
	ErrInvalidPassword = errors.New("invalid or empty password")
)

type inputValidation interface {
	validateLoginRequest(r *http.Request) error
}

func (req userLoginRequest) validateLoginRequest(r *http.Request) error {
	// validate the email is not null
	if !govalidator.IsEmail(req.email) {
		return ErrInvalidEmail
	}
	// validate the name is not empty or missing
	if govalidator.IsNull(req.password) {
		return ErrInvalidPassword
	}
	return nil
}