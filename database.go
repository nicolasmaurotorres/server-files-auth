package main

import (
	"errors"
	"fmt"
	"io"
	"io/ioutil"
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

func getPathOperation(req CommonOperation) string {
	var path string
	if LogedUsers[req.GetToken()].Category == REQUEST_DOCTOR {
		email := LogedUsers[req.GetToken()].Email
		path = GetDatabaseInstance().BasePath + email + GetDatabaseInstance().Separator + req.GetFolder()
	} else {
		path = GetDatabaseInstance().BasePath + req.GetFolder() //el pladema ya me manda el path con el email
	}
	return path
}

func (db *database) AddFolder(req AddFolderRequest) error {
	param := &req
	path := getPathOperation(param)
	errCreate := os.Mkdir(path, GetDatabaseInstance().ModePermitions) //checkeo si puedo crear la carpeta
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

func (db *database) DeleteFolder(req DelFolderRequest) error {
	param := &req
	path := getPathOperation(param)
	if folderContainOpenFile(req.Token, req.Folder) {
		return errors.New(ERROR_FOLDER_WITH_OPEN_FILE)
	}
	//email := LogedUsers[req.Token].Email
	if _, err := os.Stat(path); os.IsNotExist(err) {
		return errors.New(ERROR_FOLDER_NOT_EXISTS)
	}
	errDel := os.RemoveAll(path)
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
		tokenUser := generateTokenWithoutControl(UserLoginRequest{Email: userDeleted.Email, Password: userDeleted.Password}, REQUEST_DOCTOR)
		files := OpenedFiles[tokenUser]
		delete(OpenedFiles, tokenUser)

		if errDelFolder != nil {
			if theValue != nil {
				LogedUsers[tokenDeletedUser] = theValue //"rollback"
			}
			OpenedFiles[tokenUser] = files                //add the opened files
			errRollBack := collection.Insert(userDeleted) // re-insert the deleted user
			if errRollBack != nil {
				return errors.New(ERROR_SERVER)
			}
			return errDelFolder
		}
	}
	return nil
}

func (db *database) RenameFolder(req RenameFolderRequest) error {
	tokenUser := req.Token
	if LogedUsers[req.Token].Category == REQUEST_PLADEMA {
		//el usuario que genero esta operacion, es un usuario pladema, tengo que generar el token del doctor al cual voy a renombrar la carpeta
		slices := s.Split(req.OldFolder, GetDatabaseInstance().Separator)
		if GetDatabaseInstance().ExistsEmail(slices[0]) {
			user, _ := GetDatabaseInstance().GetUserByEmail(slices[0])
			token := generateTokenWithoutControl(UserLoginRequest{Email: user.Email, Password: user.Password}, REQUEST_DOCTOR)
			tokenUser = token
		} else {
			return errors.New(ERROR_EMAIL_NOT_EXISTS) // el email del path no existe
		}
	}
	//obtengo los archivos abiertos del usuario doctor
	openedFilesForUser, ptrs := OpenedFiles[tokenUser]
	email := LogedUsers[tokenUser].Email
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

func (db *database) AddFile(r *http.Request) error {
	r.ParseMultipartForm(500 << 20)          // 500mb tamaño maximo
	file, handler, err := r.FormFile("file") // obtengo el archivo del request
	if err != nil {
		return err
	}
	defer file.Close()
	var toReturn AddFileRequest
	toReturn.Token = r.FormValue("token")
	toReturn.Folder = r.FormValue("folder")
	toReturn.File = handler.Filename
	path := getPathOperation(&toReturn)
	if _, err := os.Stat(path); err == nil {
		return errors.New(ERROR_FILE_ALREADY_EXISTS)
	}
	f, errOpen := os.OpenFile(path, os.O_WRONLY|os.O_CREATE, 0666) // creo un archivo con el nombre del que me mandaron
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

func (db *database) DeleteFile(req DelFileRequest) error {
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

func (db *database) RenameFile(req RenameFileDoctorRequest) error {
	token := req.Token
	path := getPathOperation(&req)
	if LogedUsers[req.Token].Category == REQUEST_PLADEMA {
		slices := s.Split(path, GetDatabaseInstance().Separator)
		if GetDatabaseInstance().ExistsEmail(slices[0]) {
			user, _ := GetDatabaseInstance().GetUserByEmail(slices[0])
			tokenUser := generateTokenWithoutControl(UserLoginRequest{Email: user.Email, Password: user.Password}, REQUEST_DOCTOR)
			token = tokenUser
		} else {
			return errors.New(ERROR_EMAIL_NOT_EXISTS) // el email del path no existe
		}
	}
	files, inMap := OpenedFiles[token]
	if inMap {
		//el usuario tiene algun archivo abierto,
		for _, value := range files {
			if s.Contains(value, req.FileOld) {
				return errors.New(ERROR_FILE_OPENED_RENAMED) // el archivo que quiere cambiarle el nombre, esta abierto por algun proceso, no puede cambiarle el nombre
			}
		}
		return errors.New(ERROR_SERVER) // no deberia pasar NUNCA, dado que si se encuentra en el mapa, tiene que haber algun archivo abierto
	}
	email := LogedUsers[req.Token].Email
	newFilePath := GetDatabaseInstance().BasePath + email + GetDatabaseInstance().Separator + req.Folder + GetDatabaseInstance().Separator + req.FileNew
	oldFilePath := GetDatabaseInstance().BasePath + email + GetDatabaseInstance().Separator + req.Folder + GetDatabaseInstance().Separator + req.FileOld
	if _, err := os.Stat(newFilePath); err == nil {
		return errors.New(ERROR_FILE_ALREADY_EXISTS)
	}
	errChange := os.Rename(oldFilePath, newFilePath)
	if errChange != nil {
		return errChange
	}
	return nil
}

func (db *database) DoctorOpenFile(req OpenFileRequest) (string, error) {
	email := LogedUsers[req.Token].Email
	pathFile := GetDatabaseInstance().BasePath + email + GetDatabaseInstance().Separator
	if req.Folder == "" { // carpeta base
		pathFile = pathFile + req.File
	} else {
		pathFile = pathFile + req.Folder + GetDatabaseInstance().Separator + req.File
	}
	if _, err := os.Stat(pathFile); os.IsNotExist(err) {
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
	filePath := GetDatabaseInstance().BasePath + email + GetDatabaseInstance().Separator
	if req.Folder == "" {
		filePath = filePath + req.File
	} else {
		filePath = filePath + req.Folder + GetDatabaseInstance().Separator + req.File
	}
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
	if req.NewPassword != "" {
		query := make(map[string]string)
		if req.NewEmail != req.OldEmail { //por si cambio el email
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
	oldToken := generateTokenWithoutControl(UserLoginRequest{Email: user.Email, Password: user.Password}, user.Category)
	//lo deslogueo si esta logueado
	for key := range LogedUsers {
		if key == oldToken {
			delete(LogedUsers, key)
			break
		}
	}
	// borro todos los archivos abiertos si es que tiene
	_, inMap := OpenedFiles[oldToken]
	if inMap {
		delete(OpenedFiles, oldToken)
	}
}

type Directorys struct {
	Folder     string
	Files      []string
	SubFolders []Directorys
}

func (db *database) DoctorGetFiles(req JwtToken) Directorys {
	email := LogedUsers[req.Token].Email
	basePath := email
	base := Directorys{Folder: basePath, Files: make([]string, 0), SubFolders: make([]Directorys, 0)}
	return DFSFolders(base)
}

func DFSFolders(acum Directorys) Directorys {
	files, _ := ioutil.ReadDir(GetDatabaseInstance().BasePath + acum.Folder)
	for _, f := range files {
		if f.IsDir() {
			//es un directorio, busco en las prufundidades
			aux := Directorys{Folder: acum.Folder + GetDatabaseInstance().Separator + f.Name(), Files: make([]string, 0), SubFolders: make([]Directorys, 0)}
			subFolder := DFSFolders(aux)
			acum.SubFolders = append(acum.SubFolders, subFolder)
		} else {
			//es un archivo, lo agrego al arreglo de archivos
			acum.Files = append(acum.Files, f.Name())
		}
	}
	return acum
}

func (db *database) AdminViewUsers() []string {
	files, _ := ioutil.ReadDir(GetDatabaseInstance().BasePath)
	var toReturn []string
	toReturn = make([]string, 0)

	for _, item := range files {
		toReturn = append(toReturn, item.Name())
	}

	return toReturn
}

func (db *database) PlademaSearchFiles(req SearchFileRequest) []Directorys {
	files, _ := ioutil.ReadDir(GetDatabaseInstance().BasePath)
	var toReturn []Directorys
	for _, item := range files {
		if len(req.Emails) == 0 {
			toExplore := item.Name()
			aux := Directorys{Folder: toExplore, Files: make([]string, 0), SubFolders: make([]Directorys, 0)}
			subFolder := DFSFolders(aux)
			toReturn = append(toReturn, subFolder)
		} else {
			for _, email := range req.Emails {
				if s.Contains(item.Name(), email) {
					aux := Directorys{Folder: item.Name(), Files: make([]string, 0), SubFolders: make([]Directorys, 0)}
					subFolder := DFSFolders(aux)
					toReturn = append(toReturn, subFolder)
				}
			}
		}
	}
	return toReturn
}

func (db *database) CopyFile(req ChangeFileRequest) error {
	path := getPathOperation(&req)
	sFile, err := os.Open(path + GetDatabaseInstance().Separator + req.File)
	defer sFile.Close()
	if err != nil {
		fmt.Println(path)
		return err // the file does not exits
	}
	newLocation := req.GetDestinationFolder()
	if _, err := os.Stat(newLocation); os.IsNotExist(err) { // carpeta destino no existe
		return errors.New(ERROR_FOLDER_NOT_EXISTS)
	}
	newFileLocation := newLocation + GetDatabaseInstance().Separator + req.File
	if _, err := os.Stat(newFileLocation); err == nil { // archivo en el destino ya existe
		return errors.New(ERROR_FILE_ALREADY_EXISTS)
	}

	eFile, err := os.Create(newFileLocation)
	defer eFile.Close()
	if err != nil {
		return err
	}
	_, err = io.Copy(eFile, sFile) // first var shows number of bytes
	if err != nil {
		return err
	}
	err = eFile.Sync()
	if err != nil {
		return err
	}
	return nil
}

func (db *database) CopyFolder(req CopyFolderRequest) error {
	path := getPathOperation(&req) //carpeta que voy a copiar
	dest := req.GetDestinationFolder()
	return CopyDir(path, dest)
}

func CopyDir(source string, dest string) (err error) {

	// get properties of source dir
	fi, err := os.Stat(source)
	if err != nil {
		return err
	}

	if !fi.IsDir() {
		return errors.New(ERROR_FOLDER_NOT_EXISTS)
	}

	// ensure dest dir does not already exist

	_, err = os.Open(dest)
	if !os.IsNotExist(err) {
		return errors.New(ERROR_FOLDER_EXISTS)
	}

	// create dest dir

	err = os.MkdirAll(dest, fi.Mode())
	if err != nil {
		return err
	}

	entries, err := ioutil.ReadDir(source)

	for _, entry := range entries {

		sfp := source + "/" + entry.Name()
		dfp := dest + "/" + entry.Name()
		if entry.IsDir() {
			err = CopyDir(sfp, dfp)
			if err != nil {
				return err
			}
		} else {
			// perform copy
			err = CopyTheFile(sfp, dfp)
			if err != nil {
				return err
			}
		}

	}
	return nil
}

// Copies file source to destination dest.
func CopyTheFile(source string, dest string) (err error) {
	sf, err := os.Open(source)
	if err != nil {
		return err
	}
	defer sf.Close()
	df, err := os.Create(dest)
	if err != nil {
		return err
	}
	defer df.Close()
	_, err = io.Copy(df, sf)
	if err == nil {
		si, err := os.Stat(source)
		if err != nil {
			err = os.Chmod(dest, si.Mode())
		}

	}
	return nil
}

func (db *database) GetEmails() []string {
	var toReturn []string
	entries, err := ioutil.ReadDir(GetDatabaseInstance().BasePath)
	if err != nil {
		log.Fatal(err)
	}
	for _, file := range entries {
		toReturn = append(toReturn, file.Name())
	}
	return toReturn
}
