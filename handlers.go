package main

import (
	"encoding/json"
	"errors"
	"net/http"
	"time"

	jwt "github.com/dgrijalva/jwt-go"
)

type Response struct {
	Message string `json:"message"`
	Status  int    `json:"status"`
}

// Precondicion: el token que se pasa es valido, contiene el user.Email y user.Password y son datos validos
func generateToken(user UserLoginRequest, cat int8) (JwtToken, error) {
	var key []byte
	userDBO, err := GetDatabaseInstance().GetUserByEmail(user.Email, cat)
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
	tokenJSON, _ := json.Marshal(token)
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
	tokenJSON, _ := json.Marshal(token)
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
	tokenJSON, _ := json.Marshal(token)
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
func AdminEditUser(w http.ResponseWriter, r *http.Request) {}

// add a file to visualize
func DoctorAddFile(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json; charset=UTF-8")
	addFileRequest, errToken := GetParserInstance().DoctorAddFileRequest(r)
	var response Response
	if errToken != nil {
		w.WriteHeader(http.StatusBadRequest)
		response.Status = http.StatusBadRequest
		response.Message = errToken.Error()
	} else {
		err := GetDatabaseInstance().DoctorAddFile(addFileRequest)
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
func DoctorDeleteFile(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json; charset=UTF-8")
	delFileRequest, errToken := GetParserInstance().DoctorDeleteFileRequest(r)
	var response Response
	if errToken != nil {
		w.WriteHeader(http.StatusBadRequest)
		response.Status = http.StatusBadRequest
		response.Message = errToken.Error()
	} else {
		err := GetDatabaseInstance().DoctorDeleteFile(delFileRequest)
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
func DoctorGetFiles(w http.ResponseWriter, req *http.Request) {}

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

// search in the files by some filters on json object and return a json object with the result
func SearchFiles(w http.ResponseWriter, r *http.Request) {}

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
	}
	responseJSON, _ := json.Marshal(response)
	w.Write(responseJSON)
}

func DoctorAddFolder(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json; charset=UTF-8")
	folderRequest, err := GetParserInstance().DoctorAddFolderRequest(r)
	var response Response
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		response.Message = err.Error()
		response.Status = http.StatusBadRequest
	} else {
		errCreate := GetDatabaseInstance().DoctorAddFolder(folderRequest)
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

func DoctorDeleteFolder(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json; charset=UTF-8")
	delFolderRequest, err := GetParserInstance().DoctorDeleteFolderRequest(r)
	var response Response
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		response.Message = err.Error()
		response.Status = http.StatusBadRequest
	} else {
		errFolder := GetDatabaseInstance().DoctorDeleteFolder(delFolderRequest)
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

func DoctorRenameFolder(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json; charset=UTF-8")
	renameFolderRequest, err := GetParserInstance().DoctorRenameFolderRequest(r)
	var response Response
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		response.Message = err.Error()
		response.Status = http.StatusBadRequest
	} else {
		errRenamFolder := GetDatabaseInstance().DoctorRenameFolder(renameFolderRequest)
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

func DoctorRenameFile(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json; charset=UTF-8")
	renameFileRequest, err := GetParserInstance().DoctorRenameFileRequest(r)
	var response Response
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		response.Message = err.Error()
		response.Status = http.StatusBadRequest
	} else {
		errRenamFolder := GetDatabaseInstance().DoctorRenameFile(renameFileRequest)
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

func DoctorChangeFileFolder(w http.ResponseWriter, r *http.Request) {}
