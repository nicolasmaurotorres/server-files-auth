package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"math"
	"net/http"
	"sync/atomic"
	"time"

	"github.com/asaskevich/govalidator"
)

type CommonOperation interface {
	GetFolder() string
	GetToken() string
}

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

func (p *parser) UserLoginRequest(r *http.Request, cat int) (UserLoginRequest, error) {
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
	Email    string `json:"email"`
	Password string `json:"password"`
	Category int    `json:"category"` // 0 doctor, 1 pladema
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
	Token  string `json:"token"`
	Folder string `json:"folder"`
}

func (adr *AddFolderRequest) GetToken() string {
	return adr.Token
}

func (adr *AddFolderRequest) GetFolder() string {
	return adr.Folder
}

func (p *parser) AddFolderRequest(r *http.Request) (AddFolderRequest, error) {
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
	if govalidator.IsNull(toReturn.Folder) {
		return toReturn, errors.New(ERROR_BAD_FORMED_NAME)
	}
	return toReturn, nil
}

type DelFolderRequest struct {
	Token  string `json:"token"`
	Folder string `json:"folder"`
}

func (adr *DelFolderRequest) GetToken() string {
	return adr.Token
}

func (adr *DelFolderRequest) GetFolder() string {
	return adr.Folder
}

func (p *parser) DeleteFolderRequest(r *http.Request) (DelFolderRequest, error) {
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
	Token string `json:"token"`
	Email string `json:"email"`
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
	return toReturn, nil
}

type RenameFolderRequest struct {
	Token     string `json:"token"`
	OldFolder string `json:"oldfolder"`
	NewFolder string `json:"newfolder"`
}

func (p *parser) RenameFolderRequest(r *http.Request) (RenameFolderRequest, error) {
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

func (p *parser) DoctorAddFileRequest(r *http.Request) error {
	var toReturn AddFileRequest
	toReturn.Token = r.FormValue("token")
	toReturn.Folder = r.FormValue("folder")
	errToken := isValidToken(toReturn.Token, true)
	if errToken != nil {
		return errToken
	}
	if govalidator.IsNull(toReturn.Folder) {
		toReturn.Folder = "" //intenta agregar un archivo en la carpeta del mismo email
	}
	return nil
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
		toReturn.Folder = "" // carpeta default
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

func (rfr *RenameFileDoctorRequest) GetToken() string {
	return rfr.Token
}

func (rfr *RenameFileDoctorRequest) GetFolder() string {
	return rfr.Folder
}

func (p *parser) RenameFileRequest(r *http.Request) (RenameFileDoctorRequest, error) {
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
		toReturn.Folder = "" //carpeta base
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
		toReturn.Folder = "" //carpeta base
	}
	if govalidator.IsNull(toReturn.File) {
		return toReturn, errors.New(ERROR_BAD_FORMED_FILE_NAME)
	}
	return toReturn, nil
}

type EditUserRequest struct {
	Token       string `json:"token"`
	OldPassword string `json:"oldpassword"`
	OldEmail    string `json:"oldemail"`
	NewName     string `json:"newname"`
	NewPassword string `json:"newpassword"`
	NewEmail    string `json:"newemail"`
}

func (p *parser) AdminEditUserRequest(r *http.Request) (EditUserRequest, error) {
	var toReturn EditUserRequest
	err := json.NewDecoder(r.Body).Decode(&toReturn)
	defer r.Body.Close()
	if err != nil {
		return toReturn, errors.New(ERROR_NOT_JSON_NEEDED)
	}
	errToken := isValidToken(toReturn.Token, true)
	if errToken != nil {
		return toReturn, errToken
	}
	if govalidator.IsNull(toReturn.OldPassword) {
		return toReturn, errors.New(ERROR_BAD_FORMED_PASSWORD)
	}
	_, errOldUser := GetDatabaseInstance().GetUserByEmail(toReturn.OldEmail)
	if errOldUser != nil {
		return toReturn, errOldUser
	}
	if toReturn.NewEmail != "" && !govalidator.IsEmail(toReturn.NewEmail) {
		return toReturn, errors.New(ERROR_BAD_FORMED_EMAIL)
	}
	return toReturn, nil
}

func (p *parser) DoctorGetFilesRequest(r *http.Request) (JwtToken, error) {
	var toReturn JwtToken
	err := json.NewDecoder(r.Body).Decode(&toReturn)
	if err != nil {
		return toReturn, errors.New(ERROR_NOT_JSON_NEEDED)
	}
	errToken := isValidToken(toReturn.Token, true)
	if errToken != nil {
		return toReturn, errToken
	}
	return toReturn, nil
}

type SearchFileRequest struct {
	Token  string   `json:"token"`
	Emails []string `json:"emails"` // si los emails estan vacios, traigo todos los emails
}

func (p *parser) PlademaSearchFilesRequest(r *http.Request) (SearchFileRequest, error) {
	var toReturn SearchFileRequest
	err := json.NewDecoder(r.Body).Decode(&toReturn)
	if err != nil {
		return toReturn, errors.New(ERROR_NOT_JSON_NEEDED)
	}
	errToken := isValidToken(toReturn.Token, true)
	if errToken != nil {
		return toReturn, errToken
	}
	return toReturn, nil
}

type ChangeFileRequest struct {
	Token          string `json:"token"`
	File           string `json:"file"`
	ActualLocation string `json:"actuallocation"`
	NewLocation    string `json:"newlocation"`
}

func (cfr *ChangeFileRequest) GetToken() string {
	return cfr.Token
}

func (cfr *ChangeFileRequest) GetFolder() string {
	return cfr.ActualLocation
}

func (cfr *ChangeFileRequest) GetDestinationFolder() string {
	if LogedUsers[cfr.Token].Category == REQUEST_DOCTOR {
		email := LogedUsers[cfr.Token].Email
		return GetDatabaseInstance().BasePath + email + GetDatabaseInstance().Separator + cfr.NewLocation
	}
	return GetDatabaseInstance().BasePath + cfr.NewLocation //el email puede ser distinto si es un usuario pladema
}

func (p *parser) CopyFileRequest(r *http.Request) (ChangeFileRequest, error) {
	var toReturn ChangeFileRequest
	err := json.NewDecoder(r.Body).Decode(&toReturn)
	if err != nil {
		return toReturn, errors.New(ERROR_NOT_JSON_NEEDED)
	}
	errToken := isValidToken(toReturn.Token, true)
	if errToken != nil {
		return toReturn, errToken
	}
	if govalidator.IsNull(toReturn.File) {
		return toReturn, errors.New(ERROR_BAD_FORMED_FILE_NAME)
	}
	if govalidator.IsNull(toReturn.NewLocation) || govalidator.IsNull(toReturn.ActualLocation) {
		return toReturn, errors.New(ERROR_BAD_FORMED_FOLDER)
	}
	return toReturn, nil
}

type CopyFolderRequest struct {
	Token       string `json:"token"`
	Folder      string `json:"folder"`
	NewLocation string `json:"newlocation"`
}

func (cfr *CopyFolderRequest) GetToken() string {
	return cfr.Token
}

func (cft *CopyFolderRequest) GetFolder() string {
	return cft.Folder
}

func (cft *CopyFolderRequest) GetDestinationFolder() string {
	if LogedUsers[cft.Token].Category == REQUEST_DOCTOR {
		email := LogedUsers[cft.Token].Email
		return GetDatabaseInstance().BasePath + email + GetDatabaseInstance().Separator + cft.NewLocation
	}
	return GetDatabaseInstance().BasePath + cft.NewLocation
}

func (p *parser) CopyFolderRequest(r *http.Request) (CopyFolderRequest, error) {
	var toReturn CopyFolderRequest
	err := json.NewDecoder(r.Body).Decode(&toReturn)
	if err != nil {
		return toReturn, errors.New(ERROR_NOT_JSON_NEEDED)
	}
	errToken := isValidToken(toReturn.Token, true)
	if errToken != nil {
		return toReturn, errToken
	}
	if govalidator.IsNull(toReturn.NewLocation) || govalidator.IsNull(toReturn.Folder) {
		return toReturn, errors.New(ERROR_BAD_FORMED_FOLDER)
	}
	return toReturn, nil
}

func (p *parser) GetEmailsRequest(r *http.Request) error {
	var toReturn JwtToken
	err := json.NewDecoder(r.Body).Decode(&toReturn)
	if err != nil {
		return errors.New(ERROR_NOT_JSON_NEEDED)
	}
	errToken := isValidToken(toReturn.Token, true)
	if errToken != nil {
		return errToken
	}
	return nil
}
