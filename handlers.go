package main

import (
	"archive/zip"
	"encoding/json"
	"errors"
	"fmt"
	"io"
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

func ZipFiles(filename string, files []string) error {

	newfile, err := os.Create(filename)
	if err != nil {
		return err
	}
	defer newfile.Close()

	zipWriter := zip.NewWriter(newfile)
	defer zipWriter.Close()

	// Add files to zip
	for _, file := range files {

		zipfile, err := os.Open(file)
		if err != nil {
			return err
		}
		defer zipfile.Close()

		// Get the file information
		info, err := zipfile.Stat()
		if err != nil {
			return err
		}

		header, err := zip.FileInfoHeader(info)
		if err != nil {
			return err
		}

		// Change to deflate to gain better compression
		// see http://golang.org/pkg/archive/zip/#pkg-constants
		header.Method = zip.Deflate

		writer, err := zipWriter.CreateHeader(header)
		if err != nil {
			return err
		}
		_, err = io.Copy(writer, zipfile)
		if err != nil {
			return err
		}
	}
	return nil
}

func generateTokenWithoutControl(user UserLoginRequest, cat int) string {
	var key []byte
	switch cat {
	case REQUEST_SPECIALIST:
		key = SigningKeySpecialist
		break
	case REQUEST_TECHNICIAN:
		key = SigningKeyTechnician
		break
	case REQUEST_ADMIN:
		key = SigningKeyAdmin
		break
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"username": user.Email,
		"category": cat,
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
	case REQUEST_SPECIALIST:
		key = SigningKeySpecialist
		category = REQUEST_SPECIALIST
		break
	case REQUEST_TECHNICIAN:
		key = SigningKeyTechnician
		category = REQUEST_TECHNICIAN
		break
	case REQUEST_ADMIN:
		key = SigningKeyAdmin
		break
	default:
		return toReturn, errors.New(ERROR_SERVER)
	}
	//genero el token del usuario, lo guardo en la hash, y lo devuelvo
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"username": userDBO.Email,
		"category": userDBO.Category,
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

type ResponseLogin struct {
	Message  string `json:"message"`
	Status   int    `json:"status"`
	Category int    `json:"category"`
}

func RouteLogin(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json; charset=UTF-8")
	var responseOK ResponseLogin
	var responseBAD Response
	var tokenJSON []byte
	user, cat, err := GetParserInstance().GetUserAndCategory(r)
	if err != nil {
		w.WriteHeader(http.StatusBadGateway)
		responseBAD.Message = err.Error()
		responseBAD.Status = http.StatusBadGateway
		tokenJSON, _ = json.Marshal(responseBAD)
	} else {
		token, err := generateToken(user, cat)
		if err != nil {
			w.WriteHeader(http.StatusBadGateway)
			responseBAD.Message = err.Error()
			responseBAD.Status = http.StatusBadGateway
			tokenJSON, _ = json.Marshal(responseBAD)
		} else {
			w.WriteHeader(http.StatusOK)
			responseOK.Message = token.Token
			responseOK.Status = http.StatusOK
			responseOK.Category = cat
			tokenJSON, _ = json.Marshal(responseOK)
		}
	}
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
			fmt.Println("error logica")
		} else {
			w.WriteHeader(http.StatusOK)
			response.Status = http.StatusOK
			response.Message = FILE_ADD_SUCCESS
		}
	}
	responseJSON, _ := json.Marshal(response)
	w.Write(responseJSON)
}

type ResponseAdminUsers struct {
	Message string         `json:"message"`
	Status  int            `json:"status"`
	Users   []UserCategory `json:"users"`
}

func AdminViewUsers(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json; charset=UTF-8")
	_, errToken := GetParserInstance().AdminViewRequest(r)
	var responseOK ResponseAdminUsers
	var responseBAD Response
	var responseJSON []byte
	if errToken != nil {
		w.WriteHeader(http.StatusBadRequest)
		responseBAD.Status = http.StatusBadRequest
		responseBAD.Message = errToken.Error()
		responseJSON, _ = json.Marshal(responseBAD)
	} else {
		users := GetDatabaseInstance().AdminViewUsers() // traigo todos los datos de todos los usuarios
		w.WriteHeader(http.StatusOK)
		responseOK.Message = GET_FILES_SUCCESS
		responseOK.Status = http.StatusOK
		responseOK.Users = users
		responseJSON, _ = json.Marshal(responseOK)
	}
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

// add a file modified by some technician engeneeir
func TechnicianAddFile(w http.ResponseWriter, r *http.Request) {}

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
func SpecialistGetFiles(w http.ResponseWriter, req *http.Request) {
	w.Header().Set("Content-Type", "application/json; charset=UTF-8")
	getFiles, err := GetParserInstance().SpecialistGetFilesRequest(req)
	var response ResponseDirectorys
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		response.Message = err.Error()
		response.Status = http.StatusBadRequest
	} else {
		folder := GetDatabaseInstance().SpecialistGetFiles(getFiles)
		w.WriteHeader(http.StatusOK)
		response.Message = "OK"
		response.Status = http.StatusOK
		response.Folders = folder
	}
	responseJSON, _ := json.Marshal(response)
	w.Write(responseJSON)
}

// add to the hashtable the file that is opened
func SpecialistOpenFile(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json; charset=UTF-8")
	openFileRequest, err := GetParserInstance().SpecialistOpenFileRequest(r)
	var response Response
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		response.Message = err.Error()
		response.Status = http.StatusBadRequest
	} else {
		folder, errCreate := GetDatabaseInstance().SpecialistOpenFile(openFileRequest)
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

func SpecialistDownloadFile(w http.ResponseWriter, r *http.Request) {
	getFileRequest, err := GetParserInstance().SpecialistGetFile(r)
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
			http.Error(w, "File not found.", 404) //File not found, send 404
		} else {
			slices := s.Split(getFileRequest.File, GetDatabaseInstance().Separator)
			tempFolder := ""
			for j := 0; j <= len(slices)-2; j++ {
				tempFolder = tempFolder + slices[j] + GetDatabaseInstance().Separator
			}
			fmt.Println(tempFolder)
			FileWithExtention := slices[len(slices)-1]   // obtengo el nombre del archivo con extension
			partsFile := s.Split(FileWithExtention, ".") // partes del archivo
			FileName := partsFile[0]                     // file name
			files := []string{GetDatabaseInstance().BasePath + getFileRequest.File}
			output := GetDatabaseInstance().BasePath + tempFolder + FileName + ".zip"
			fmt.Println(output)
			err := ZipFiles(output, files)
			if err != nil {
				http.Error(w, "Server Error Compressing", 500)
			} else {
				file, err1 := os.Open(output)
				defer file.Close()
				if err1 != nil {
					http.Error(w, "Server Error Searching Compressed file", 500)
				} else {
					w.Header().Set("Content-Type", "application/zip")
					w.Header().Set("Content-Disposition", "attachment; filename='"+FileName+".zip'")
					http.ServeFile(w, r, output)
					defer os.Remove(output)
				}
			}
		}
	}
}

// remove a file of the hashtable of opened files
func SpecialistCloseFile(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json; charset=UTF-8")
	closeFileRequest, err := GetParserInstance().SpecialistCloseFileRequest(r)
	var response Response
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		response.Message = err.Error()
		response.Status = http.StatusBadRequest
	} else {
		errCreate := GetDatabaseInstance().SpecialistCloseFile(closeFileRequest)
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

type ResponseTechnicianDirectorys struct {
	Message string       `json:"message"`
	Status  int          `json:"status"`
	Folders []Directorys `json:"folders"`
}

// search in the files by some filters on json object and return a json object with the result
func TechnicianSearchFiles(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json; charset=UTF-8")
	searchFiles, err := GetParserInstance().TechnicianSearchFilesRequest(r)
	var response ResponseTechnicianDirectorys
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		response.Message = err.Error()
		response.Status = http.StatusBadRequest
	} else {
		directorys := GetDatabaseInstance().TechnicianSearchFiles(searchFiles)
		w.WriteHeader(http.StatusOK)
		response.Message = GET_FILES_SUCCESS
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

func TechnicianGetEmails(w http.ResponseWriter, r *http.Request) {
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

func TechnicianGetFile(w http.ResponseWriter, r *http.Request) {
	getFileRequest, err := GetParserInstance().TechnicianGetFile(r)
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
			http.Error(w, "File not found.", 404) //File not found, send 404
		} else {
			slices := s.Split(getFileRequest.File, GetDatabaseInstance().Separator)
			tempFolder := ""
			for j := 0; j <= len(slices)-2; j++ {
				tempFolder = tempFolder + slices[j] + GetDatabaseInstance().Separator
			}
			fmt.Println(tempFolder)
			FileWithExtention := slices[len(slices)-1]   // obtengo el nombre del archivo con extension
			partsFile := s.Split(FileWithExtention, ".") // partes del archivo
			FileName := partsFile[0]                     // file name
			files := []string{GetDatabaseInstance().BasePath + getFileRequest.File}
			output := GetDatabaseInstance().BasePath + tempFolder + FileName + ".zip"
			fmt.Println(output)
			err := ZipFiles(output, files)
			if err != nil {
				http.Error(w, "Server Error Compressing", 500)
			} else {
				file, err1 := os.Open(output)
				defer file.Close()
				if err1 != nil {
					http.Error(w, "Server Error Searching Compressed file", 500)
				} else {
					w.Header().Set("Content-Type", "application/zip")
					w.Header().Set("Content-Disposition", "attachment; filename='"+FileName+".zip'")
					http.ServeFile(w, r, output)
					defer os.Remove(output)
				}
			}
		}
	}
}
