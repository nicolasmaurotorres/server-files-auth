package main

import (
	"encoding/json"
	"errors"
	"net/http"
	"os"
	s "strings"
	"time"

	jwt "github.com/dgrijalva/jwt-go"
)

type Response struct {
	Message string `json:"message"`
	Status  int    `json:"status"`
}

func generateTokenWithoutControl(user UserLoginRequest, cat int) string {
	var key []byte
	switch cat {
	case REQUEST_DOCTOR:
		key = SigningKeyDoctor
		break
	case REQUEST_PLADEMA:
		key = SigningKeyPladema
		break
	case REQUEST_ADMIN:
		key = SigningKeyAdmin
		break
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"username": user.Email,
		"password": user.Password,
	})
	tokenString, _ := token.SignedString(key)
	return tokenString
}

// Precondicion: el token que se pasa es valido, contiene el user.Email y user.Password y son datos validos
func generateToken(user UserLoginRequest, cat int) (JwtToken, error) {
	var key []byte
	var toReturn JwtToken
	var category int
	userDBO, err := GetDatabaseInstance().GetUserByEmail(user.Email)
	if err != nil {
		return JwtToken{Token: ""}, err
	}
	// control de pass y user que sean iguales
	// TODO: hacer un hashing de la pass al enviarla desde el cleinte y al dar de alta en el servidor
	if user.Password != userDBO.Password || user.Email != userDBO.Email || userDBO.Category != cat {
		return toReturn, errors.New(ERROR_MISSMATCH_USER_PASSWORD)
	}
	switch cat {
	case REQUEST_DOCTOR:
		key = SigningKeyDoctor
		category = REQUEST_DOCTOR
		break
	case REQUEST_PLADEMA:
		key = SigningKeyPladema
		category = REQUEST_PLADEMA
		break
	case REQUEST_ADMIN:
		key = SigningKeyAdmin
		break
	default:
		return toReturn, errors.New(ERROR_SERVER)
	}
	//genero el token del usuario, lo guardo en la hash, y lo devuelvo
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"username": user.Email,
		"password": user.Password,
	})
	tokenString, _ := token.SignedString(key)
	//controlo que el usuario NO este logueado previamente
	_, inMap := LogedUsers[tokenString]
	if inMap {
		return toReturn, errors.New(ERROR_USER_ALREADY_LOGUED)
	}
	toReturn.Token = tokenString
	LogedUsers[tokenString] = &Pair{TimeLogIn: time.Now(), Email: user.Email, Category: category}
	return toReturn, nil
}

func LoginPerson(cat int, r *http.Request) (JwtToken, error) {
	user, err := GetParserInstance().UserLoginRequest(r, cat)
	var toReturn JwtToken
	if err != nil {
		return toReturn, err
	}
	token, err := generateToken(user, cat)
	if err != nil {
		return toReturn, err
	}
	return token, nil
}

// returns a auth token as doctor user
func DoctorLogin(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json; charset=UTF-8")
	var response Response
	token, err := LoginPerson(REQUEST_DOCTOR, r)
	if err != nil {
		w.WriteHeader(http.StatusBadGateway)
		response.Message = err.Error()
		response.Status = http.StatusBadGateway
	} else {
		w.WriteHeader(http.StatusOK)
		response.Message = token.Token
		response.Status = http.StatusOK
	}
	tokenJSON, _ := json.Marshal(response)
	w.Write(tokenJSON)
}

// returns a auth token as pladema user
func PlademaLogin(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json; charset=UTF-8")
	var response Response
	token, err := LoginPerson(REQUEST_PLADEMA, r)
	if err != nil {
		w.WriteHeader(http.StatusBadGateway)
		response.Message = err.Error()
		response.Status = http.StatusBadGateway
	} else {
		w.WriteHeader(http.StatusOK)
		response.Message = token.Token
		response.Status = http.StatusOK
	}
	tokenJSON, _ := json.Marshal(response)
	w.Write(tokenJSON)
}

// returns a auth token as admin user
func AdminLogin(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json; charset=UTF-8")
	var response Response
	token, err := LoginPerson(REQUEST_ADMIN, r)
	if err != nil {
		w.WriteHeader(http.StatusBadGateway)
		response.Message = err.Error()
		response.Status = http.StatusBadGateway
	} else {
		w.WriteHeader(http.StatusOK)
		response.Message = token.Token
		response.Status = http.StatusOK
	}
	tokenJSON, _ := json.Marshal(response)
	w.Write(tokenJSON)
}

// add a new user with the data on a json
func AdminAddUser(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json; charset=UTF-8")
	newUser, erro := GetParserInstance().AdminAddUser(r)
	var response Response
	if erro == nil {
		err := GetDatabaseInstance().AdminAddUser(newUser)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			response.Status = http.StatusInternalServerError
			response.Message = err.Error()
		} else {
			w.WriteHeader(http.StatusOK)
			response.Status = http.StatusOK
			response.Message = USER_CREATED_SUCCESS
		}
	} else {
		w.WriteHeader(http.StatusBadRequest)
		response.Status = http.StatusBadRequest
		response.Message = erro.Error()
	}
	responseJSON, _ := json.Marshal(response)
	w.Write(responseJSON)
}

// delete some valid user with the data on a json
func AdminDeleteUser(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json; charset=UTF-8")
	delUser, err := GetParserInstance().AdminDeleteUserRequest(r)
	var response Response
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		response.Status = http.StatusBadRequest
		response.Message = err.Error()
	} else {
		errDel := GetDatabaseInstance().AdminDeleteUser(delUser)
		if errDel != nil {
			w.WriteHeader(http.StatusBadRequest)
			response.Status = http.StatusBadRequest
			response.Message = err.Error()
		} else {
			w.WriteHeader(http.StatusOK)
			response.Status = http.StatusOK
			response.Message = DELETE_USER_SUCCESS
		}
	}
	responseJSON, _ := json.Marshal(response)
	w.Write(responseJSON)
}

// change name or email of a valid user
func AdminEditUser(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json; charset=UTF-8")
	editRequest, errToken := GetParserInstance().AdminEditUserRequest(r)
	var response Response
	if errToken != nil {
		w.WriteHeader(http.StatusBadRequest)
		response.Status = http.StatusBadRequest
		response.Message = errToken.Error()
	} else {
		err := GetDatabaseInstance().AdminEditUser(editRequest)
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			response.Status = http.StatusBadRequest
			response.Message = err.Error()
		} else {
			w.WriteHeader(http.StatusOK)
			response.Status = http.StatusOK
			response.Message = FILE_ADD_SUCCESS
		}
	}
	responseJSON, _ := json.Marshal(response)
	w.Write(responseJSON)
}

func AddFile(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json; charset=UTF-8")
	errToken := GetParserInstance().AddFileRequest(r)
	var response Response
	if errToken != nil {
		w.WriteHeader(http.StatusBadRequest)
		response.Status = http.StatusBadRequest
		response.Message = errToken.Error()
	} else {
		err := GetDatabaseInstance().AddFile(r)
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			response.Status = http.StatusBadRequest
			response.Message = err.Error()
		} else {
			w.WriteHeader(http.StatusOK)
			response.Status = http.StatusOK
			response.Message = FILE_ADD_SUCCESS
		}
	}
	responseJSON, _ := json.Marshal(response)
	w.Write(responseJSON)
}

// add a file modified by some pladema engeneeir
func PlademaAddFile(w http.ResponseWriter, r *http.Request) {}

// remove a file of the hashtable of opened files
func DeleteFile(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json; charset=UTF-8")
	delFileRequest, errToken := GetParserInstance().DeleteFileRequest(r)
	var response Response
	if errToken != nil {
		w.WriteHeader(http.StatusBadRequest)
		response.Status = http.StatusBadRequest
		response.Message = errToken.Error()
	} else {
		err := GetDatabaseInstance().DeleteFile(delFileRequest)
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			response.Status = http.StatusBadRequest
			response.Message = err.Error()
		} else {
			w.WriteHeader(http.StatusOK)
			response.Status = http.StatusOK
			response.Message = FILE_DELETED_SUCCESS
		}
	}
	responseJSON, _ := json.Marshal(response)
	w.Write(responseJSON)

}

type ResponseDirectorys struct {
	Message string     `json:"message"`
	Status  int        `json:"status"`
	Folders Directorys `json:"folders"`
}

// returns all files to visualize
func DoctorGetFiles(w http.ResponseWriter, req *http.Request) {
	w.Header().Set("Content-Type", "application/json; charset=UTF-8")
	getFiles, err := GetParserInstance().DoctorGetFilesRequest(req)
	var response ResponseDirectorys
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		response.Message = err.Error()
		response.Status = http.StatusBadRequest
	} else {
		folder := GetDatabaseInstance().DoctorGetFiles(getFiles)
		w.WriteHeader(http.StatusOK)
		response.Message = "OK"
		response.Status = http.StatusOK
		response.Folders = folder
	}
	responseJSON, _ := json.Marshal(response)
	w.Write(responseJSON)
}

// add to the hashtable the file that is opened
func DoctorOpenFile(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json; charset=UTF-8")
	openFileRequest, err := GetParserInstance().DoctorOpenFileRequest(r)
	var response Response
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		response.Message = err.Error()
		response.Status = http.StatusBadRequest
	} else {
		folder, errCreate := GetDatabaseInstance().DoctorOpenFile(openFileRequest)
		if errCreate != nil {
			w.WriteHeader(http.StatusForbidden)
			response.Message = errCreate.Error()
			response.Status = http.StatusForbidden
		} else {
			w.WriteHeader(http.StatusOK)
			response.Message = folder
			response.Status = http.StatusOK
		}
	}
	responseJSON, _ := json.Marshal(response)
	w.Write(responseJSON)
}

// remove a file of the hashtable of opened files
func DoctorCloseFile(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json; charset=UTF-8")
	closeFileRequest, err := GetParserInstance().DoctorCloseFileRequest(r)
	var response Response
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		response.Message = err.Error()
		response.Status = http.StatusBadRequest
	} else {
		errCreate := GetDatabaseInstance().DoctorCloseFile(closeFileRequest)
		if errCreate != nil {
			w.WriteHeader(http.StatusForbidden)
			response.Message = errCreate.Error()
			response.Status = http.StatusForbidden
		} else {
			w.WriteHeader(http.StatusOK)
			response.Message = FILE_CLOSE_SUCCESS
			response.Status = http.StatusOK
		}
	}
	responseJSON, _ := json.Marshal(response)
	w.Write(responseJSON)
}

type ResponsePlademaDirectorys struct {
	Message string       `json:"message"`
	Status  int          `json:"status"`
	Folders []Directorys `json:"folders"`
}

// search in the files by some filters on json object and return a json object with the result
func PlademaSearchFiles(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json; charset=UTF-8")
	searchFiles, err := GetParserInstance().PlademaSearchFilesRequest(r)
	var response ResponsePlademaDirectorys
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		response.Message = err.Error()
		response.Status = http.StatusBadRequest
	} else {
		directorys := GetDatabaseInstance().PlademaSearchFiles(searchFiles)
		w.WriteHeader(http.StatusOK)
		response.Message = CREATE_FOLDER_SUCCESS
		response.Status = http.StatusOK
		response.Folders = directorys
	}
	responseJSON, _ := json.Marshal(response)
	w.Write(responseJSON)
}

func Logout(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json; charset=UTF-8")
	token, err := GetParserInstance().LogoutRequest(r)
	var response Response
	if err != nil {
		// hay un error
		w.WriteHeader(http.StatusBadRequest)
		response.Message = err.Error()
		response.Status = http.StatusBadRequest
	} else {
		// checkeo si el token esta en memoria
		_, inMap := LogedUsers[token.Token]
		if inMap {
			w.WriteHeader(http.StatusOK)
			delete(LogedUsers, token.Token) // elimino al usuario
			response.Message = LOGOUT_SUCCESS
			response.Status = http.StatusOK
		} else {
			// el usuario que esta intentando desloguearse, no esta logueado, es un error, es un token valido pero no esta
			// en la hash, tiramos otro mensaje de error para despistar :V
			w.WriteHeader(http.StatusForbidden)
			response.Message = ERROR_NOT_VALID_TOKEN
			response.Status = http.StatusForbidden
		}
		// elimino todos los archivos abiertos
		delete(OpenedFiles, token.Token)
	}
	responseJSON, _ := json.Marshal(response)
	w.Write(responseJSON)
}

func AddFolder(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json; charset=UTF-8")
	folderRequest, err := GetParserInstance().AddFolderRequest(r)
	var response Response
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		response.Message = err.Error()
		response.Status = http.StatusBadRequest
	} else {
		errCreate := GetDatabaseInstance().AddFolder(folderRequest)
		if errCreate != nil {
			w.WriteHeader(http.StatusForbidden)
			response.Message = errCreate.Error()
			response.Status = http.StatusForbidden
		} else {
			w.WriteHeader(http.StatusOK)
			response.Message = CREATE_FOLDER_SUCCESS
			response.Status = http.StatusOK
		}
	}
	responseJSON, _ := json.Marshal(response)
	w.Write(responseJSON)
}

func DeleteFolder(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json; charset=UTF-8")
	delFolderRequest, err := GetParserInstance().DeleteFolderRequest(r)
	var response Response
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		response.Message = err.Error()
		response.Status = http.StatusBadRequest
	} else {
		errFolder := GetDatabaseInstance().DeleteFolder(delFolderRequest)
		if errFolder != nil {
			w.WriteHeader(http.StatusForbidden)
			response.Message = errFolder.Error()
			response.Status = http.StatusForbidden
		} else {
			w.WriteHeader(http.StatusOK)
			response.Message = DELETE_FOLDER_SUCCESS
			response.Status = http.StatusOK
		}
	}
	responseJSON, _ := json.Marshal(response)
	w.Write(responseJSON)
}

func RenameFolder(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json; charset=UTF-8")
	renameFolderRequest, err := GetParserInstance().RenameFolderRequest(r)
	var response Response
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		response.Message = err.Error()
		response.Status = http.StatusBadRequest
	} else {
		errRenamFolder := GetDatabaseInstance().RenameFolder(renameFolderRequest)
		if errRenamFolder != nil {
			w.WriteHeader(http.StatusBadRequest)
			response.Message = errRenamFolder.Error()
			response.Status = http.StatusBadRequest
		} else {
			w.WriteHeader(http.StatusOK)
			response.Message = RENAME_FOLDER_SUCCESS
			response.Status = http.StatusOK
		}
	}
	responseJSON, _ := json.Marshal(response)
	w.Write(responseJSON)
}

func RenameFile(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json; charset=UTF-8")
	renameFileRequest, err := GetParserInstance().RenameFileRequest(r)
	var response Response
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		response.Message = err.Error()
		response.Status = http.StatusBadRequest
	} else {
		errRenamFolder := GetDatabaseInstance().RenameFile(renameFileRequest)
		if errRenamFolder != nil {
			w.WriteHeader(http.StatusBadRequest)
			response.Message = errRenamFolder.Error()
			response.Status = http.StatusBadRequest
		} else {
			w.WriteHeader(http.StatusOK)
			response.Message = RENAME_FILE_SUCCESS
			response.Status = http.StatusOK
		}
	}
	responseJSON, _ := json.Marshal(response)
	w.Write(responseJSON)
}

func CopyFile(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json; charset=UTF-8")
	changeFileRequest, err := GetParserInstance().CopyFileRequest(r)
	var response Response
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		response.Message = err.Error()
		response.Status = http.StatusBadRequest
	} else {
		errRenamFolder := GetDatabaseInstance().CopyFile(changeFileRequest)
		if errRenamFolder != nil {
			w.WriteHeader(http.StatusBadRequest)
			response.Message = errRenamFolder.Error()
			response.Status = http.StatusBadRequest
		} else {
			w.WriteHeader(http.StatusOK)
			response.Message = COPY_FILE_SUCESS
			response.Status = http.StatusOK
		}
	}
	responseJSON, _ := json.Marshal(response)
	w.Write(responseJSON)
}

func CopyFolder(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json; charset=UTF-8")
	changeFileRequest, err := GetParserInstance().CopyFolderRequest(r)
	var response Response
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		response.Message = err.Error()
		response.Status = http.StatusBadRequest
	} else {
		errRenamFolder := GetDatabaseInstance().CopyFolder(changeFileRequest)
		if errRenamFolder != nil {
			w.WriteHeader(http.StatusBadRequest)
			response.Message = errRenamFolder.Error()
			response.Status = http.StatusBadRequest
		} else {
			w.WriteHeader(http.StatusOK)
			response.Message = COPY_FOLDER_SUCCESS
			response.Status = http.StatusOK
		}
	}
	responseJSON, _ := json.Marshal(response)
	w.Write(responseJSON)
}

type ResponseEmails struct {
	Message string   `json:"message"`
	Status  int      `json:"status"`
	Emails  []string `json:"emails"`
}

func PlademaGetEmails(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json; charset=UTF-8")
	err := GetParserInstance().GetEmailsRequest(r)
	var response ResponseEmails
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		response.Message = err.Error()
		response.Status = http.StatusBadRequest
	} else {
		emails := GetDatabaseInstance().GetEmails()
		response.Status = http.StatusBadRequest
		w.WriteHeader(http.StatusOK)
		response.Message = GET_EMAILS_SUCCESS
		response.Status = http.StatusOK
		response.Emails = emails
	}
	responseJSON, _ := json.Marshal(response)
	w.Write(responseJSON)
}

func PlademaGetFile(w http.ResponseWriter, r *http.Request) {
	getFileRequest, err := GetParserInstance().PlademaGetFile(r)
	if err != nil {
		w.Header().Set("Content-Type", "application/json; charset=UTF-8")
		var response Response
		response.Message = err.Error()
		response.Status = http.StatusBadRequest
		responseJSON, _ := json.Marshal(response)
		w.Write(responseJSON)
	} else {
		//Check if file exists and open
		Openfile, err := os.Open(GetDatabaseInstance().BasePath + getFileRequest.File)
		defer Openfile.Close() //Close after function return
		if err != nil {
			//File not found, send 404
			http.Error(w, "File not found.", 404)
		} else {

			w.Header().Set("Content-Type", "application/octet-stream")
			slices := s.Split(getFileRequest.File, GetDatabaseInstance().Separator)
			FIleName := slices[len(slices)-1] //obtengo el nombre del archivo
			//Get the Content-Type of the file
			//Create a buffer to store the header of the file in
			FileHeader := make([]byte, 512)
			//Copy the headers into the FileHeader buffer
			Openfile.Read(FileHeader)
			//Get content type of file
			//FileContentType := http.DetectContentType(FileHeader)
			//Get the file size
			//FileStat, _ := Openfile.Stat() //Get info from file
			//FileSize := strconv.FormatInt(FileStat.Size(), 10) //Get file size as a string
			//Send the headers
			// tell the browser the returned content should be downloaded
			//w.Header().Add("Content-Disposition", "Attachment")
			w.Header().Add("Content-Disposition", "attachment; filename="+FIleName)
			//w.Header().Add("Content-Type", FileContentType)
			//w.Header().Add("Content-Length", FileSize)
			//Send the file
			//We read 512 bytes from the file already so we reset the offset back to 0
			//	Openfile.Seek(0, 0)
			//	io.Copy(w, Openfile) //'Copy' the file to the client
			modtime := time.Now()
			http.ServeContent(w, r, FIleName, modtime, Openfile)
		}
	}
}
