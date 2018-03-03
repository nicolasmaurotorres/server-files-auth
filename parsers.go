package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"math"
	"net/http"
	"os"
	"sync/atomic"
	"time"

	"github.com/asaskevich/govalidator"
)

var initializedParser uint32
var instanceParser *parser

type parser struct{}

func GetParserInstance() *parser {

	if atomic.LoadUint32(&initializedParser) == 1 {
		return instanceParser
	}

	mu.Lock()
	defer mu.Unlock()

	if initializedParser == 0 {
		instanceParser = &parser{}
		atomic.StoreUint32(&initializedParser, 1)
	}

	return instanceParser
}

// iValidToken returns if a token is valid or not
func isValidToken(tokenString string, checkTimeStamp bool) error {
	value, inMap := LogedUsers[tokenString]
	if inMap {
		if checkTimeStamp {
			diff := time.Now().Sub(value.TimeLogIn)
			minutesDiff := math.Abs(diff.Minutes())
			if minutesDiff > 5 {
				delete(LogedUsers, tokenString) // lo elimino por tiempo invalido
				return errors.New(ERROR_REQUIRE_LOGIN_AGAIN)
			}
			LogedUsers[tokenString].TimeLogIn = time.Now() // actualizo el tiempo que se checkeo el token
			return nil
		}
		return nil
	}
	return errors.New(ERROR_NOT_LOGUED_USER)
}

type UserLoginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

func (p *parser) UserLoginRequest(r *http.Request, cat int8) (UserLoginRequest, error) {
	var user UserLoginRequest
	err := json.NewDecoder(r.Body).Decode(&user)
	defer r.Body.Close()
	if err == nil { // no hay errores de json mal formado
		if cat != REQUEST_ADMIN {
			// usuario pladema o doctor
			if !govalidator.IsEmail(user.Email) {
				return user, errors.New(ERROR_BAD_FORMED_EMAIL)
			}
		} else {
			// usuario admin
			if govalidator.IsNull(user.Email) {
				return user, errors.New(ERROR_BAD_FORMED_EMAIL_EMPTY)
			}
		}
		// validate the password is not empty or missing
		if govalidator.IsNull(user.Password) {
			return user, errors.New(ERROR_BAD_FORMED_PASSWORD)
		}
		return user, nil
	}
	// hay error de parsing json
	return user, err
}

type NewUserRequest struct {
	Token    string `json:"token"`
	Name     string `json:"name"`
	Email    string `json:"email"`
	Password string `json:"password"`
	Category int8   `json:"category"` // 0 doctor, 1 pladema
}

// GetNewUserJSONRequest returns a user with valid data
func (p *parser) AdminAddUser(r *http.Request) (NewUserRequest, error) {
	var newUserRequest NewUserRequest
	err := json.NewDecoder(r.Body).Decode(&newUserRequest)
	defer r.Body.Close()
	if err != nil {
		fmt.Println(err)
		//the json input is valid but we have to check the data values
		return newUserRequest, errors.New(ERROR_NOT_JSON_NEEDED)
	}
	errorMessage := isValidToken(newUserRequest.Token, true)
	if errorMessage != nil {
		return newUserRequest, errorMessage
	}

	if !govalidator.IsEmail(newUserRequest.Email) {
		return newUserRequest, errors.New(ERROR_BAD_FORMED_EMAIL)
	}

	if GetDatabaseInstance().ExistsEmail(newUserRequest.Email) {
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

func (p *parser) LogoutRequest(r *http.Request) (JwtToken, error) {
	var userLogoutRequest JwtToken
	err := json.NewDecoder(r.Body).Decode(&userLogoutRequest)
	defer r.Body.Close()
	if err == nil {
		return userLogoutRequest, err
	}
	errValid := isValidToken(userLogoutRequest.Token, false)
	if errValid != nil {
		return userLogoutRequest, errValid
	}
	return userLogoutRequest, nil
}

type AddFolderRequest struct {
	Token string `json:"token"`
	Name  string `json:"name"`
}

func (p *parser) DoctorAddFolderRequest(r *http.Request) (AddFolderRequest, error) {
	var toReturn AddFolderRequest
	err := json.NewDecoder(r.Body).Decode(&toReturn)
	defer r.Body.Close()
	if err != nil {
		return toReturn, errors.New(ERROR_NOT_JSON_NEEDED)
	}
	errValid := isValidToken(toReturn.Token, true)
	if errValid != nil {
		return toReturn, errValid
	}
	if govalidator.IsNull(toReturn.Token) {
		return toReturn, errors.New(ERROR_BAD_FORMED_TOKEN)
	}
	if govalidator.IsNull(toReturn.Name) {
		return toReturn, errors.New(ERROR_BAD_FORMED_NAME)
	}
	return toReturn, nil
}

type DelFolderRequest struct {
	Token  string `json:"token"`
	Folder string `json:"folder"`
}

func (p *parser) DoctorDeleteFolderRequest(r *http.Request) (DelFolderRequest, error) {
	var toReturn DelFolderRequest
	err := json.NewDecoder(r.Body).Decode(&toReturn)
	defer r.Body.Close()
	if err != nil {
		return toReturn, errors.New(ERROR_NOT_JSON_NEEDED)
	}
	errValid := isValidToken(toReturn.Token, true)
	if errValid != nil {
		return toReturn, errValid
	}
	if govalidator.IsNull(toReturn.Token) {
		return toReturn, errors.New(ERROR_BAD_FORMED_TOKEN)
	}
	if govalidator.IsNull(toReturn.Folder) {
		return toReturn, errors.New(ERROR_BAD_FORMED_NAME)
	}
	return toReturn, nil
}

type DelUserRequest struct {
	Token    string `json:"token"`
	Email    string `json:"email"`
	Category int8   `json:"category"` // 0 doctor, 1 pladema
}

func (p *parser) AdminDeleteUserRequest(r *http.Request) (DelUserRequest, error) {
	var toReturn DelUserRequest
	err := json.NewDecoder(r.Body).Decode(&toReturn)
	defer r.Body.Close()
	if err != nil {
		return toReturn, errors.New(ERROR_NOT_JSON_NEEDED)
	}
	errToken := isValidToken(toReturn.Token, true)
	if errToken != nil {
		return toReturn, errToken
	}
	if !govalidator.IsEmail(toReturn.Email) {
		return toReturn, errors.New(ERROR_BAD_FORMED_EMAIL)
	}
	if !GetDatabaseInstance().ExistsEmail(toReturn.Email) {
		return toReturn, errors.New(ERROR_EMAIL_ALREADY_EXISTS)
	}
	if !(toReturn.Category == REQUEST_PLADEMA || toReturn.Category == REQUEST_DOCTOR) {
		return toReturn, errors.New(ERROR_BAD_CATEGORY)
	}
	return toReturn, nil
}

type RenameFolderRequest struct {
	Token     string `json:"token"`
	OldFolder string `json:"oldfolder"`
	NewFolder string `json:"newfolder"`
}

func (p *parser) DoctorRenameFolderRequest(r *http.Request) (RenameFolderRequest, error) {
	var toReturn RenameFolderRequest
	err := json.NewDecoder(r.Body).Decode(&toReturn)
	defer r.Body.Close()
	if err != nil {
		return toReturn, errors.New(ERROR_NOT_JSON_NEEDED)
	}
	errToken := isValidToken(toReturn.Token, true)
	if errToken != nil {
		return toReturn, errToken
	}
	if govalidator.IsNull(toReturn.OldFolder) {
		return toReturn, errors.New(ERROR_BAD_FORMED_OLD_NAME)
	}
	if govalidator.IsNull(toReturn.NewFolder) {
		return toReturn, errors.New(ERROR_BAD_FORMED_NEW_NAME)
	}
	return toReturn, nil
}

type AddFileRequest struct {
	Token  string
	Folder string
	File   string
}

func (p *parser) DoctorAddFileRequest(r *http.Request) (AddFileRequest, error) {
	var toReturn AddFileRequest
	r.ParseMultipartForm(500 << 20)          // 500mb tamaño maximo
	file, handler, err := r.FormFile("file") // obtengo el archivo del request
	if err != nil {
		fmt.Println(err)
		return toReturn, err
	}
	defer file.Close()
	toReturn.Token = r.FormValue("token")
	toReturn.Folder = r.FormValue("folder")
	toReturn.File = handler.Filename

	errToken := isValidToken(toReturn.Token, true)
	if errToken != nil {
		return toReturn, errToken
	}
	if govalidator.IsNull(toReturn.Folder) {
		return toReturn, errors.New(ERROR_BAD_FORMED_FOLDER)
	}
	if govalidator.IsNull(toReturn.File) {
		return toReturn, errors.New(ERROR_BAD_FORMED_FILE_NAME)
	}
	email := LogedUsers[toReturn.Token].Email
	f, err := os.OpenFile(GetDatabaseInstance().BasePath+email+GetDatabaseInstance().Separator+toReturn.Folder+GetDatabaseInstance().Separator+handler.Filename, os.O_WRONLY|os.O_CREATE, 0666) // creo un archivo con el nombre del que me mandaron
	if err != nil {
		fmt.Println(err)
		return toReturn, err
	}
	defer f.Close()
	_, errCopy := io.Copy(f, file) //copio lo del file del request al nuevo lugar
	if errCopy != nil {
		fmt.Println("error al copiar")
		return toReturn, errCopy
	}
	return toReturn, nil
}

type DelFileRequest struct {
	Token  string `json:"token"`
	Folder string `json:"folder"`
	File   string `json:"file"`
}

func (p *parser) DoctorDeleteFileRequest(r *http.Request) (DelFileRequest, error) {
	var toReturn DelFileRequest
	err := json.NewDecoder(r.Body).Decode(&toReturn)
	defer r.Body.Close()
	if err != nil {
		return toReturn, errors.New(ERROR_NOT_JSON_NEEDED)
	}
	errToken := isValidToken(toReturn.Token, true)
	if errToken != nil {
		return toReturn, errToken
	}
	if govalidator.IsNull(toReturn.Folder) {
		return toReturn, errors.New(ERROR_BAD_FORMED_FOLDER)
	}

	if govalidator.IsNull(toReturn.File) {
		return toReturn, errors.New(ERROR_BAD_FORMED_FILE_NAME)
	}
	return toReturn, nil
}

type RenameFileDoctorRequest struct {
	Token   string `json:"token"`
	FileOld string `json:"fileold"`
	FileNew string `json:"filenew"`
	Folder  string `json:"folder"`
}

func (p *parser) DoctorRenameFileRequest(r *http.Request) (RenameFileDoctorRequest, error) {
	var toReturn RenameFileDoctorRequest
	err := json.NewDecoder(r.Body).Decode(&toReturn)
	defer r.Body.Close()
	if err != nil {
		return toReturn, errors.New(ERROR_NOT_JSON_NEEDED)
	}
	errToken := isValidToken(toReturn.Token, true)
	if errToken != nil {
		return toReturn, errToken
	}
	if govalidator.IsNull(toReturn.Folder) {
		return toReturn, errors.New(ERROR_BAD_FORMED_FOLDER)
	}
	if govalidator.IsNull(toReturn.FileOld) || govalidator.IsNull(toReturn.FileNew) {
		return toReturn, errors.New(ERROR_BAD_FORMED_FILE_NAME)
	}
	return toReturn, nil
}

type OpenFileRequest struct {
	Token  string `json:"token"`
	File   string `json:"file"`
	Folder string `json:"folder"`
}

func (p *parser) DoctorOpenFileRequest(r *http.Request) (OpenFileRequest, error) {
	var toReturn OpenFileRequest
	err := json.NewDecoder(r.Body).Decode(&toReturn)
	defer r.Body.Close()
	if err != nil {
		return toReturn, errors.New(ERROR_NOT_JSON_NEEDED)
	}
	errToken := isValidToken(toReturn.Token, true)
	if errToken != nil {
		return toReturn, errToken
	}
	if govalidator.IsNull(toReturn.Folder) {
		return toReturn, errors.New(ERROR_BAD_FORMED_FOLDER)
	}
	if govalidator.IsNull(toReturn.File) {
		return toReturn, errors.New(ERROR_BAD_FORMED_FILE_NAME)
	}
	return toReturn, nil
}

type CloseFileRequest struct {
	Token  string `json:"token"`
	File   string `json:"file"`
	Folder string `json:"folder"`
}

func (p *parser) DoctorCloseFileRequest(r *http.Request) (CloseFileRequest, error) {
	var toReturn CloseFileRequest
	err := json.NewDecoder(r.Body).Decode(&toReturn)
	defer r.Body.Close()
	if err != nil {
		return toReturn, errors.New(ERROR_NOT_JSON_NEEDED)
	}
	errToken := isValidToken(toReturn.Token, true)
	if errToken != nil {
		return toReturn, errToken
	}
	if govalidator.IsNull(toReturn.Folder) {
		return toReturn, errors.New(ERROR_BAD_FORMED_FOLDER)
	}
	if govalidator.IsNull(toReturn.File) {
		return toReturn, errors.New(ERROR_BAD_FORMED_FILE_NAME)
	}
	return toReturn, nil
}
