package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"time"

	jwt "github.com/dgrijalva/jwt-go"
)

type Response struct {
	Message string `json:"message"`
	Status  int    `json:"status"`
}

// Precondicion: el token que se pasa es valido, contiene el user.Email y user.Password y son datos validos
func GenerateToken(user UserLoginRequest, cat int8) (JwtToken, error) {
	var key []byte
	userDBO, err := GetUserByEmail(user.Email, cat)
	if err != nil {
		return JwtToken{Token: ""}, err
	}
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
	default:
		return JwtToken{Token: ""}, errors.New(ERROR_SERVER)
	}
	// control de pass y user que sean iguales
	// TODO: hacer un hashing de la pass al enviarla desde el cleinte y al dar de alta en el servidor
	if user.Password != userDBO.Password || user.Email != userDBO.Email {
		return JwtToken{Token: ""}, errors.New(ERROR_SERVER)
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
		return JwtToken{Token: ""}, errors.New(ERROR_USER_ALREADY_LOGUED)
	}
	var toReturn JwtToken
	toReturn.Token = tokenString
	LogedUsers[tokenString] = &Pair{TimeLogIn: time.Now(), Email: user.Email}
	return toReturn, nil
}

func LoginPerson(cat int8, r *http.Request) (JwtToken, error) {
	user, err := GetLoginJSONRequest(r, cat)
	if err != nil {
		fmt.Println(err.Error())
		return JwtToken{Token: ""}, err
	}
	token, err := GenerateToken(user, cat)
	if err != nil {
		return JwtToken{Token: ""}, err
	}
	return token, nil
}

// returns the error of login
func GetLoginError(err string) ([]byte, int) {
	var e Response
	switch err {
	case ERROR_BAD_FORMED_EMAIL:
		e.Status = http.StatusBadRequest
		break
	case ERROR_BAD_FORMED_PASSWORD:
		e.Status = http.StatusBadRequest
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
	default:
		e.Status = http.StatusForbidden
		break
	}
	e.Message = err
	exceptionJSON, _ := json.Marshal(e)
	return exceptionJSON, e.Status
}

// returns a auth token as doctor user
func LoginDoctor(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json; charset=UTF-8")
	token, err := LoginPerson(REQUEST_DOCTOR, r)
	if err != nil {
		exceptionJSON, status := GetLoginError(err.Error())
		w.WriteHeader(status)
		w.Write(exceptionJSON)
	} else {
		w.WriteHeader(http.StatusOK)
		tokenJSON, _ := json.Marshal(token)
		w.Write(tokenJSON)
	}
}

// returns a auth token as pladema user
func LoginPladema(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json; charset=UTF-8")
	token, err := LoginPerson(REQUEST_PLADEMA, r)
	if err != nil {
		exceptionJSON, status := GetLoginError(err.Error())
		w.WriteHeader(status)
		w.Write(exceptionJSON)
	} else {
		w.WriteHeader(http.StatusOK)
		tokenJSON, _ := json.Marshal(token)
		w.Write(tokenJSON)
	}
}

// returns a auth token as admin user
func LoginAdmin(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json; charset=UTF-8")
	token, err := LoginPerson(REQUEST_ADMIN, r)
	if err != nil {
		exceptionJSON, status := GetLoginError(err.Error())
		w.WriteHeader(status)
		w.Write(exceptionJSON)

	} else {
		w.WriteHeader(http.StatusOK)
		tokenJSON, _ := json.Marshal(token)
		w.Write(tokenJSON)
	}
}

// add a new user with the data on a json
func AddUser(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json; charset=UTF-8")
	newUser, erro := GetNewUserJSONRequest(r)
	var response Response
	if erro == nil {
		//	fmt.Println("pase por addUser sin error")
		err := NewUserDAL(newUser)
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
func DelUser(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json; charset=UTF-8")
	delUser, err := GetDelUserRequestFromJSONRequest(r)
	var response Response
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		response.Status = http.StatusBadRequest
		response.Message = err.Error()
	} else {
		errDel := DeleteUser(delUser)
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
func EditUser(w http.ResponseWriter, r *http.Request) {}

// add a file to visualize
func AddFileDoctor(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json; charset=UTF-8")
	addFileRequest, errToken := GetAddFileDoctorFromJSONRequest(r)
	var response Response
	if errToken != nil {
		w.WriteHeader(http.StatusBadRequest)
		response.Status = http.StatusBadRequest
		response.Message = errToken.Error()
	} else {
		err := AddFileDoctorDB(addFileRequest)
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
func AddFilePladema(w http.ResponseWriter, r *http.Request) {}

// remove a file of the hashtable of opened files
func DelFile(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json; charset=UTF-8")
	delFileRequest, errToken := GetDelFileDoctorFromJSONRequest(r)
	var response Response
	if errToken != nil {
		w.WriteHeader(http.StatusBadRequest)
		response.Status = http.StatusBadRequest
		response.Message = errToken.Error()
	} else {
		err := DelFileDoctorDB(delFileRequest)
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

// returns all files to visualize
func AllFiles(w http.ResponseWriter, req *http.Request) {}

// add to the hashtable the file that is opened
func OpenFile(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json; charset=UTF-8")
	openFileRequest, err := GetOpenFileRequestFromJSONRequest(r)
	var response Response
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		response.Message = err.Error()
		response.Status = http.StatusBadRequest
	} else {
		folder, errCreate := OpenFileBL(openFileRequest)
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
func CloseFile(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json; charset=UTF-8")
	closeFileRequest, err := GetCloseFileRequestFromJSONRequest(r)
	var response Response
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		response.Message = err.Error()
		response.Status = http.StatusBadRequest
	} else {
		errCreate := CloseFileBL(closeFileRequest)
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

// search in the files by some filters on json object and return a json object with the result
func SearchFiles(w http.ResponseWriter, r *http.Request) {}

func RoutineLogout(w http.ResponseWriter, tokenString string) bool {
	var response Response
	validToken, errorMessage := IsValidToken(tokenString, false)
	if validToken {
		_, inMap := LogedUsers[tokenString]
		if inMap {
			w.WriteHeader(http.StatusOK)
			delete(LogedUsers, tokenString) // elimino al usuario
			response.Message = LOGOUT_SUCCESS
			response.Status = http.StatusOK
			responseJSON, _ := json.Marshal(response)
			w.Write(responseJSON)
			return true
		}
		// el usuario que esta intentando desloguearse, no esta logueado, es un error, es un token valido pero no esta
		// en la hash, tiramos otro mensaje de error para despistar :V
		w.WriteHeader(http.StatusForbidden)
		response.Message = ERROR_NOT_VALID_TOKEN
		response.Status = http.StatusForbidden
		responseJSON, _ := json.Marshal(response)
		w.Write(responseJSON)
		return false
	}
	// no es un token valido
	w.WriteHeader(http.StatusForbidden)
	response.Message = errorMessage.Error()
	response.Status = http.StatusForbidden
	responseJSON, _ := json.Marshal(response)
	w.Write(responseJSON)
	return false
}

// logout a admin user
func Logout(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json; charset=UTF-8")
	valid, tokenString := GetTokenStringFromLogoutRequest(r)
	var response Response
	if valid {
		inMap, err := IsValidToken(tokenString, false)
		if inMap {
			RoutineLogout(w, tokenString)
		} else {
			w.WriteHeader(http.StatusForbidden)
			response.Message = err.Error()
			response.Status = http.StatusBadRequest
			responseJSON, _ := json.Marshal(response)
			w.Write(responseJSON)
		}
	} else {
		w.WriteHeader(http.StatusBadRequest)
		response.Message = ERROR_NOT_JSON_NEEDED
		response.Status = http.StatusBadRequest
		responseJSON, _ := json.Marshal(response)
		w.Write(responseJSON)
	}
}

func AddFolder(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json; charset=UTF-8")
	folderRequest, err := GetAddFolderRequestFromJSONRequest(r)
	var response Response
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		response.Message = err.Error()
		response.Status = http.StatusBadRequest
	} else {
		errCreate := CreateFolder(folderRequest)
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

func DelFolder(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json; charset=UTF-8")
	delFolderRequest, err := GetDelFolderRequestFromJSONRequest(r)
	var response Response
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		response.Message = err.Error()
		response.Status = http.StatusBadRequest
	} else {
		errFolder := DeleteFolder(delFolderRequest)
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
	renameFolderRequest, err := RenameFolderRequestFromJSONRequest(r)
	var response Response
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		response.Message = err.Error()
		response.Status = http.StatusBadRequest
	} else {
		errRenamFolder := RenameFolderDB(renameFolderRequest)
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

func RenameFileDoctor(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json; charset=UTF-8")
	renameFileRequest, err := RenameFileRequestFromJSONRequest(r)
	var response Response
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		response.Message = err.Error()
		response.Status = http.StatusBadRequest
	} else {
		errRenamFolder := RenameFileDoctorDB(renameFileRequest)
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

func MoveFileToFolder(w http.ResponseWriter, r *http.Request) {}
