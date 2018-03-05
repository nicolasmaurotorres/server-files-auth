package main

import (
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	s "strings"
	Sync "sync"
	"sync/atomic"

	mgo "gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"
)

var mu Sync.Mutex
var initialized uint32
var instance *database

type database struct {
	DataBaseName    string
	CollectionUsers string
	BasePath        string
	Separator       string
	ModePermitions  os.FileMode
}

//GetDatabaseInstance return the only instance of the database
func GetDatabaseInstance() *database {

	if atomic.LoadUint32(&initialized) == 1 {
		return instance
	}

	mu.Lock()
	defer mu.Unlock()

	if initialized == 0 {
		instance = &database{
			BasePath:        "/home/maro/Desktop/data/pvw/data/",
			CollectionUsers: "users",
			DataBaseName:    "tesis",
			Separator:       string(os.PathSeparator),
			ModePermitions:  0755,
		}
		atomic.StoreUint32(&initialized, 1)
	}

	return instance
}

type User struct {
	Name     string
	Email    string
	Password string
}

type UserDBO struct {
	Category int
	Email    string
	Password string
}

func (db *database) getSession() *mgo.Session {
	session, err := mgo.Dial("localhost:27017")
	if err != nil {
		panic(err)
	}
	return session
}

func (db *database) AdminAddUser(user NewUserRequest) error {
	session := GetDatabaseInstance().getSession()
	defer session.Close()
	if GetDatabaseInstance().ExistsEmail(user.Email) {
		return errors.New(ERROR_EMAIL_ALREADY_EXISTS)
	}
	collection := session.DB(GetDatabaseInstance().DataBaseName).C(GetDatabaseInstance().CollectionUsers)
	var userDBO UserDBO
	userDBO.Email = user.Email
	userDBO.Password = user.Password
	switch user.Category {
	case REQUEST_DOCTOR:
		userDBO.Category = REQUEST_DOCTOR
		os.Mkdir(GetDatabaseInstance().BasePath+userDBO.Email, GetDatabaseInstance().ModePermitions)
		break
	case REQUEST_PLADEMA:
		userDBO.Category = REQUEST_PLADEMA
	default:
		log.Fatal(ERROR_SERVER)
		return errors.New(ERROR_SERVER)
	}
	err := collection.Insert(userDBO)
	if err != nil {
		return errors.New(ERROR_INSERT_NEW_USER)
	}
	return nil
}

func (db *database) ExistsEmail(email string) bool {
	session := GetDatabaseInstance().getSession()
	defer session.Close()
	collection := session.DB(GetDatabaseInstance().DataBaseName).C(GetDatabaseInstance().CollectionUsers)
	var userResult User
	err := collection.Find(bson.M{"email": email}).One(&userResult)
	if err != nil {
		return false
	}
	return true
}

func (db *database) GetUserByEmail(email string) (UserDBO, error) {
	var userToReturn UserDBO // this user is the most general of the three
	session := GetDatabaseInstance().getSession()
	defer session.Close()
	collection := session.DB(GetDatabaseInstance().DataBaseName).C(GetDatabaseInstance().CollectionUsers)
	query := make(map[string]string)
	query["email"] = email
	err := collection.Find(query).One(&userToReturn)
	if err != nil {
		return userToReturn, errors.New(ERROR_NOT_EXISTING_USER)
	}
	return userToReturn, nil
}

func (db *database) DoctorAddFolder(req AddFolderRequest) error {
	email := LogedUsers[req.Token].Email
	errCreate := os.Mkdir(GetDatabaseInstance().BasePath+email+GetDatabaseInstance().Separator+req.Name, GetDatabaseInstance().ModePermitions) //checkeo si puedo crear la carpeta
	if errCreate != nil {
		return errCreate // no se pudo crear la carpeta
	}
	return nil
}

func folderContainOpenFile(token string, path string) bool {
	files, inMap := OpenedFiles[token]
	if inMap {
		// el usuario tiene algun archivo abierto, checkeo si la carpeta "path" esta abierta con algun archivo
		for _, value := range files {
			if s.Contains(value, path) {
				return true // la carpeta tiene algun archivo abierto
			}
		}
	}
	return false
}

func (db *database) DoctorDeleteFolder(req DelFolderRequest) error {
	//TODO: ver como se guuardar en el OpenedFiles
	if folderContainOpenFile(req.Token, req.Folder) {
		return errors.New(ERROR_FOLDER_WITH_OPEN_FILE)
	}
	email := LogedUsers[req.Token].Email
	errDel := os.RemoveAll(GetDatabaseInstance().BasePath + email + GetDatabaseInstance().Separator + req.Folder)
	if errDel != nil {
		return errDel
	}
	return nil
}

func (db *database) AdminDeleteUser(user DelUserRequest) error {
	session := GetDatabaseInstance().getSession()
	defer session.Close()
	collection := session.DB(GetDatabaseInstance().DataBaseName).C(GetDatabaseInstance().CollectionUsers)
	query := bson.M{"email": user.Email}
	userDeleted, errEmail := GetDatabaseInstance().GetUserByEmail(user.Email)
	if errEmail != nil {
		return errEmail
	}
	errDel := collection.Remove(query)
	if errDel != nil {
		return errDel
	}
	tokenDeletedUser := ""
	var theValue *Pair
	theValue = nil
	for value, key := range LogedUsers {
		if key.Email == user.Email {
			tokenDeletedUser = value
			theValue.Email = key.Email
			theValue.TimeLogIn = key.TimeLogIn
			break
		}
	}
	if tokenDeletedUser != "" {
		delete(LogedUsers, tokenDeletedUser) // delete if user loged
	}

	if userDeleted.Category == REQUEST_DOCTOR {
		errDelFolder := os.RemoveAll(GetDatabaseInstance().BasePath + user.Email) //error deleting the folder
		// delete the opened files
		jwtToken, _ := generateToken(UserLoginRequest{Email: userDeleted.Email, Password: userDeleted.Password}, REQUEST_DOCTOR)
		files := OpenedFiles[jwtToken.Token]
		delete(OpenedFiles, jwtToken.Token)

		if errDelFolder != nil {
			if theValue != nil {
				LogedUsers[tokenDeletedUser] = theValue //"rollback"
			}
			OpenedFiles[jwtToken.Token] = files           //add the opened files
			errRollBack := collection.Insert(userDeleted) // re-insert the deleted user
			if errRollBack != nil {
				return errors.New(ERROR_SERVER)
			}
			return errDelFolder
		}
	}
	return nil
}

func (db *database) DoctorRenameFolder(req RenameFolderRequest) error {
	email := LogedUsers[req.Token].Email
	openedFilesForUser, ptrs := OpenedFiles[req.Token]
	found := false
	if ptrs {
		//significa que tiene archivos abiertos, tengo que verificar si la carpeta que quiere renombrar NO ESTE esta aqui
		//TODO: verificar como guardo en openFile operation
		for _, val := range openedFilesForUser {
			if s.Contains(val, req.OldFolder) {
				found = true
				break
			}
		}
	}

	if found {
		return errors.New(ERROR_FOLDER_WITH_OPEN_FILE)
	}
	//el archivo no esta abierto, tengo que cambiarle el nombre
	errRename := os.Rename(GetDatabaseInstance().BasePath+email+GetDatabaseInstance().Separator+req.OldFolder, GetDatabaseInstance().BasePath+email+GetDatabaseInstance().Separator+req.NewFolder)
	if errRename != nil {
		return errRename
	}
	return nil
}

func (db *database) DoctorAddFile(r *http.Request) error {
	var toReturn AddFileRequest
	r.ParseMultipartForm(500 << 20)          // 500mb tamaño maximo
	file, handler, err := r.FormFile("file") // obtengo el archivo del request
	if err != nil {
		return err
	}
	defer file.Close()
	toReturn.Token = r.FormValue("token")
	toReturn.Folder = r.FormValue("folder")
	toReturn.File = handler.Filename
	email := LogedUsers[toReturn.Token].Email
	f, errOpen := os.OpenFile(GetDatabaseInstance().BasePath+email+GetDatabaseInstance().Separator+toReturn.Folder+GetDatabaseInstance().Separator+handler.Filename, os.O_WRONLY|os.O_CREATE, 0666) // creo un archivo con el nombre del que me mandaron
	if errOpen != nil {
		return errOpen
	}
	defer f.Close()
	_, errCopy := io.Copy(f, file) //copio lo del file del request al nuevo lugar
	if errCopy != nil {

		return errCopy
	}
	return nil
}

func (db *database) DoctorDeleteFile(req DelFileRequest) error {
	email := LogedUsers[req.Token].Email
	// checkeo que el archivo que se quiera eliminar NO este abierto
	for _, value := range OpenedFiles[req.Token] {
		if s.Contains(value, req.File) && s.Contains(value, req.Folder) {
			//si contiene el nombre del archivo y el nombre de la carpeta , el archivo esta abierto
			return errors.New(ERROR_FILE_OPENED)
		}
	}
	errDel := os.Remove(GetDatabaseInstance().BasePath + email + GetDatabaseInstance().Separator + req.Folder + GetDatabaseInstance().Separator + req.File)
	if errDel != nil {
		//error al intentar borrarlo del sistema de archivos, ya sea por que no existe o el path es invalido
		return errDel
	}
	return nil
}

type list []interface{}

func (db *database) DoctorRenameFile(req RenameFileDoctorRequest) error {
	files, inMap := OpenedFiles[req.Token]
	if inMap {
		//el usuario tiene algun archivo abierto,
		for _, value := range files {
			if value == req.FileOld {
				return errors.New(ERROR_FILE_OPENED) // el archivo que quiere cambiarle el nombre, esta abierto por algun proceso, no puede cambiarle el nombre
			}
		}
		return errors.New(ERROR_SERVER) // no deberia pasar NUNCA, dado que si se encuentra en el mapa, tiene que haber algun archivo abierto
	}
	email := LogedUsers[req.Token].Email
	errChange := os.Rename(GetDatabaseInstance().BasePath+email+GetDatabaseInstance().Separator+req.Folder+GetDatabaseInstance().Separator+req.FileOld, GetDatabaseInstance().BasePath+email+GetDatabaseInstance().Separator+req.Folder+GetDatabaseInstance().Separator+req.FileNew)
	if errChange != nil {
		return errChange
	}
	return nil
}

func (db *database) DoctorOpenFile(req OpenFileRequest) (string, error) {
	email := LogedUsers[req.Token].Email
	if _, err := os.Stat(GetDatabaseInstance().BasePath + email + GetDatabaseInstance().Separator + req.Folder + GetDatabaseInstance().Separator + req.File); os.IsNotExist(err) {
		return "", errors.New(ERROR_FILE_NOT_EXISTS)
	}
	files, inMap := OpenedFiles[req.Token]
	if inMap {
		// el doctor tiene lagun archivo abierto
		for key, value := range files {
			fmt.Println("valor " + string(key) + " " + value)
			if s.Contains(value, req.Folder+GetDatabaseInstance().Separator+req.File) {
				return "", errors.New(ERROR_FILE_ALREADY_OPENED)
			}
		}
	}
	// no tiene archivos abiertos o no coincide, creo un nuevo arreglo
	pathFile := email + GetDatabaseInstance().Separator + req.Folder + GetDatabaseInstance().Separator + req.File
	OpenedFiles[req.Token] = append(OpenedFiles[req.Token], pathFile)
	return pathFile, nil
}

//elimina un objeto en una posicion del arreglo
func remove(s []string, i int) []string {
	s[len(s)-1], s[i] = s[i], s[len(s)-1]
	return s[:len(s)-1]
}

func (db *database) DoctorCloseFile(req CloseFileRequest) error {
	email := LogedUsers[req.Token].Email
	if _, err := os.Stat(GetDatabaseInstance().BasePath + email + GetDatabaseInstance().Separator + req.Folder + GetDatabaseInstance().Separator + req.File); os.IsNotExist(err) {
		return errors.New(ERROR_FILE_NOT_EXISTS)
	}
	files, inMap := OpenedFiles[req.Token]
	if inMap {
		// tiene algun archivo abierto
		for key, value := range files {
			fmt.Println("value:" + value)
			if s.Contains(value, req.Folder+GetDatabaseInstance().Separator+req.File) {
				// elimino el archivo del array
				OpenedFiles[req.Token] = remove(files, key)
				fmt.Println(OpenedFiles[req.Token])
				return nil
			}
		}
	}
	return errors.New(ERROR_FILE_NOT_OPENED)
}

func (db *database) AdminEditUser(req EditUserRequest) error {
	session := GetDatabaseInstance().getSession()
	defer session.Close()
	collection := session.DB(GetDatabaseInstance().DataBaseName).C(GetDatabaseInstance().CollectionUsers)
	if req.NewEmail != "" && req.NewEmail != req.OldEmail {
		// compruebo si el nuevo email NO existe en la DB
		if GetDatabaseInstance().ExistsEmail(req.NewEmail) {
			return errors.New(ERROR_EMAIL_ALREADY_EXISTS)
		}
		//actualizo el email en la db
		query := make(map[string]string)
		query["email"] = req.OldEmail
		modified := bson.M{"$set": bson.M{"email": req.NewEmail}}
		errUpdate := collection.Update(query, modified)
		if errUpdate != nil {
			return errUpdate
		}
		//renombro la carpeta
		os.Rename(GetDatabaseInstance().BasePath+req.OldEmail, GetDatabaseInstance().BasePath+req.NewEmail)
		cleanDataUserEdited(req.OldEmail)
	}
	if req.NewPassword != "" && req.OldPassword != req.NewPassword {
		query := make(map[string]string)
		if req.NewEmail != req.OldEmail {
			query["email"] = req.NewEmail
		} else {
			query["email"] = req.OldEmail
		}
		update := bson.M{"$set": bson.M{"password": req.NewPassword}}
		errUpdate := collection.Update(query, update)
		if errUpdate != nil {
			return errUpdate
		}
		cleanDataUserEdited(req.OldEmail)
	}
	return nil
}

func cleanDataUserEdited(email string) {
	user, _ := GetDatabaseInstance().GetUserByEmail(email)
	oldToken, _ := generateToken(UserLoginRequest{Email: user.Email, Password: user.Password}, user.Category)
	//lo deslogueo si esta logueado
	for key := range LogedUsers {
		if key == oldToken.Token {
			delete(LogedUsers, key)
			break
		}
	}
	// borro todos los archivos abiertos si es que tiene
	_, inMap := OpenedFiles[oldToken.Token]
	if inMap {
		delete(OpenedFiles, oldToken.Token)
	}
}
