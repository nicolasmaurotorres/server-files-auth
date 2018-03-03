package main

import (
	"errors"
	"fmt"
	"log"
	"os"
	"strconv"
	s "strings"
	Sync "sync"
	"sync/atomic"

	mgo "gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"
)

const (
	DATABASE_NAME       = "tesis"
	COLLECTION_USERS    = "users"
	BASE_PATH           = "/home/maro/Desktop/data/pvw/data/"
	PATH_OWN_FILES      = "/own/"
	PATH_MODIFIED_FILES = "/modified/"
	MODE_PERMITIONS     = 0755
	SEPARATOR           = string(os.PathSeparator)
)

var mu Sync.Mutex
var initialized uint32
var instance *database

type database struct {
	DataBaseName      string
	CollectionUsers   string
	BasePath          string
	PathOwnFiles      string
	PathModifiedFiles string
}

//GetDataBase return the only instance of the database
func GetDatabaseInstance() *database {

	if atomic.LoadUint32(&initialized) == 1 {
		return instance
	}

	mu.Lock()
	defer mu.Unlock()

	if initialized == 0 {
		instance = &database{
			BasePath:          BASE_PATH,
			CollectionUsers:   COLLECTION_USERS,
			DataBaseName:      DATABASE_NAME,
			PathModifiedFiles: PATH_MODIFIED_FILES,
			PathOwnFiles:      PATH_OWN_FILES,
		}
		atomic.StoreUint32(&initialized, 1)
	}

	return instance
}

type Directory struct {
	Path  string   `json:"path"`
	Files []string `json:"files"`
}

type User struct {
	Name     string
	Email    string
	Password string
}

type UserDoctorDBO struct {
	Category   int8
	Name       string
	Email      string
	Password   string
	Directorys []Directory
}

type UserPlademaDBO struct {
	Category int8
	Name     string
	Email    string
	Password string
}

type UserAdminDBO struct {
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
	collection := session.DB(DATABASE_NAME).C(COLLECTION_USERS)
	switch user.Category {
	case REQUEST_DOCTOR:
		var userDoctor UserDoctorDBO
		userDoctor.Name = user.Name
		userDoctor.Email = user.Email
		userDoctor.Password = user.Password
		userDoctor.Category = REQUEST_DOCTOR
		os.Mkdir(BASE_PATH+userDoctor.Email, MODE_PERMITIONS)
		os.Mkdir(BASE_PATH+userDoctor.Email+PATH_OWN_FILES, MODE_PERMITIONS)      // files sended by the doctor
		os.Mkdir(BASE_PATH+userDoctor.Email+PATH_MODIFIED_FILES, MODE_PERMITIONS) // files modified by the pladema user
		userDoctor.Directorys = make([]Directory, 2)
		userDoctor.Directorys[0].Files = make([]string, 1)
		userDoctor.Directorys[0].Files[0] = ""
		userDoctor.Directorys[0].Path = userDoctor.Email + PATH_OWN_FILES
		userDoctor.Directorys[1].Files = make([]string, 1)
		userDoctor.Directorys[1].Files[0] = ""
		userDoctor.Directorys[1].Path = userDoctor.Email + PATH_MODIFIED_FILES
		err := collection.Insert(userDoctor)
		if err != nil {
			return errors.New(ERROR_INSERT_NEW_DOCTOR)
		}
		return nil
	case REQUEST_PLADEMA:
		var userPladema UserPlademaDBO
		userPladema.Email = user.Email
		userPladema.Name = user.Name
		userPladema.Password = user.Password
		userPladema.Category = REQUEST_PLADEMA
		err := collection.Insert(userPladema)
		if err != nil {
			log.Fatal(ERROR_INSERT_NEW_PLADEMA)
			return errors.New(ERROR_INSERT_NEW_PLADEMA)
		}
		return nil
	default:
		log.Fatal(ERROR_SERVER)
		return errors.New(ERROR_SERVER)
	}
}

func (db *database) ExistsEmail(email string) bool {
	session := GetDatabaseInstance().getSession()
	defer session.Close()
	collection := session.DB(DATABASE_NAME).C(COLLECTION_USERS)
	var userResult User
	err := collection.Find(bson.M{"email": email}).One(&userResult)
	if err != nil {
		return false
	}
	return true
}

func (db *database) GetUserByEmail(email string, cat int8) (UserDoctorDBO, error) {
	var userToReturn UserDoctorDBO // this user is the most general of the three
	session := GetDatabaseInstance().getSession()
	defer session.Close()
	collection := session.DB(DATABASE_NAME).C(COLLECTION_USERS)
	query := make(map[string]string)
	query["email"] = email
	switch cat {
	case REQUEST_DOCTOR:
		query["category"] = strconv.Itoa(int(REQUEST_DOCTOR))
		break
	case REQUEST_PLADEMA:
		query["category"] = strconv.Itoa(int(REQUEST_PLADEMA))
		break
	case REQUEST_ADMIN:
		query["category"] = strconv.Itoa(int(REQUEST_ADMIN))
		break
	}
	err := collection.Find(query).One(&userToReturn)
	if err != nil {
		return UserDoctorDBO{Name: "", Email: ""}, errors.New(ERROR_NOT_EXISTING_USER)
	}
	return userToReturn, nil
}

func (db *database) DoctorAddFolder(req AddFolderRequest) error {
	email := LogedUsers[req.Token].Email
	errCreate := os.Mkdir(BASE_PATH+email+PATH_OWN_FILES+req.Name, MODE_PERMITIONS) //checkeo si puedo crear la carpeta
	if errCreate != nil {
		return errCreate // no se pudo crear la carpeta
	}
	var newFolder Directory
	newFolder.Files = make([]string, 1)
	newFolder.Files[0] = ""
	newFolder.Path = email + PATH_OWN_FILES + req.Name
	session := GetDatabaseInstance().getSession()
	defer session.Close()
	collection := session.DB(DATABASE_NAME).C(COLLECTION_USERS)
	query := bson.M{"email": email}
	update := bson.M{"$push": bson.M{"directorys": newFolder}}
	errUpdate := collection.Update(query, update) // actualizo al usuario
	if errUpdate != nil {
		return errUpdate // error en la actualizacion
	}
	return nil
}

func (db *database) DoctorDeleteFolder(req DelFolderRequest) error {
	email := LogedUsers[req.Token].Email
	errDel := os.RemoveAll(BASE_PATH + "/" + email + PATH_OWN_FILES + req.Folder)
	if errDel != nil {
		return errDel
	}
	session := GetDatabaseInstance().getSession()
	defer session.Close()
	collection := session.DB(DATABASE_NAME).C(COLLECTION_USERS)
	query := bson.M{"email": email}
	update := bson.M{"$pull": bson.M{"directorys": bson.M{"path": email + PATH_OWN_FILES + req.Folder}}}
	errUpdate := collection.Update(query, update)
	if errUpdate != nil {
		return errUpdate
	}
	return nil
}

func (db *database) AdminDeleteUser(user DelUserRequest) error {
	session := GetDatabaseInstance().getSession()
	defer session.Close()
	collection := session.DB(DATABASE_NAME).C(COLLECTION_USERS)
	query := bson.M{"email": user.Email}
	userDeleted, errEmail := GetDatabaseInstance().GetUserByEmail(user.Email, user.Category)
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

	if user.Category == REQUEST_DOCTOR {
		errDelFolder := os.RemoveAll(BASE_PATH + SEPARATOR + user.Email) //error deleting the folder
		if errDelFolder != nil {
			if theValue != nil {
				LogedUsers[tokenDeletedUser] = theValue //"rollback"
			}
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
	errRename := os.Rename(BASE_PATH+email+PATH_OWN_FILES+req.OldFolder, BASE_PATH+email+PATH_OWN_FILES+req.NewFolder)
	if errRename != nil {
		return errRename
	}
	session := GetDatabaseInstance().getSession()
	defer session.Close()
	collection := session.DB(DATABASE_NAME).C(COLLECTION_USERS)
	query := make(map[string]string)
	query["email"] = email
	query["directorys.path"] = email + PATH_OWN_FILES + req.OldFolder
	update := bson.M{"$set": bson.M{"directorys.$.path": email + PATH_OWN_FILES + req.NewFolder}}
	// hacer el update en la base de datos, el nombre repetido ya se ataja en el os.Rename dado que una carpeta no puede contener 2 carpetas con el mismo nombre, me aseguro
	// de que ese error no va a pasar en la DB
	errUpdate := collection.Update(query, update)
	if errUpdate != nil {
		os.Rename(BASE_PATH+email+PATH_OWN_FILES+req.NewFolder, BASE_PATH+email+PATH_OWN_FILES+req.OldFolder) // vuelvo atras con el renombre
		return errUpdate
	}

	return nil
}

func (db *database) DoctorAddFile(req AddFileRequest) error {
	email := LogedUsers[req.Token].Email
	session := GetDatabaseInstance().getSession()
	defer session.Clone()
	collection := session.DB(DATABASE_NAME).C(COLLECTION_USERS)
	query := make(map[string]string)
	query["email"] = email
	query["directorys.path"] = email + PATH_OWN_FILES + req.Folder // carpeta a agregar el archivo
	update := bson.M{"$addToSet": bson.M{"directorys.$.files": req.File}}
	errUpdate := collection.Update(query, update)
	if errUpdate != nil {
		fmt.Println(errUpdate)
		os.Remove(BASE_PATH + email + PATH_OWN_FILES + req.Folder + SEPARATOR + req.File) //elimino el archivo que guarde
		return errUpdate
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
	errDel := os.Remove(BASE_PATH + email + PATH_OWN_FILES + req.Folder + SEPARATOR + req.File)
	if errDel != nil {
		//error al intentar borrarlo del sistema de archivos, ya sea por que no existe o el path es invalido
		return errDel
	}
	session := GetDatabaseInstance().getSession()
	defer session.Clone()
	collection := session.DB(DATABASE_NAME).C(COLLECTION_USERS)
	query := make(map[string]string)
	query["email"] = email
	query["directorys.path"] = email + PATH_OWN_FILES + req.Folder // carpeta a agregar el archivo
	update := bson.M{"$pull": bson.M{"directorys.$.files": req.File}}
	errUpdate := collection.Update(query, update)
	if errUpdate != nil {
		return errUpdate
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
	errChange := os.Rename(BASE_PATH+email+PATH_OWN_FILES+req.Folder+SEPARATOR+req.FileOld, BASE_PATH+email+PATH_OWN_FILES+req.Folder+SEPARATOR+req.FileNew)
	if errChange != nil {
		return errChange
	}
	session := GetDatabaseInstance().getSession()
	defer session.Clone()
	collection := session.DB(DATABASE_NAME).C(COLLECTION_USERS)
	query := make(map[string]string)
	query["email"] = email
	query["directorys.path"] = email + PATH_OWN_FILES + req.Folder
	query["directorys.files"] = req.FileOld
	update := bson.M{"$set": bson.M{"directorys.$[].files.$[selectedFile]": req.FileNew}}
	filters := list{bson.M{"selectedFile": req.FileOld}}
	errUpdate := collection.UpdateArrayFilters(query, update, filters)
	if errUpdate != nil {
		//vuelvo atras con el nombre que tenia antes el archivo
		os.Rename(BASE_PATH+email+PATH_OWN_FILES+req.Folder+SEPARATOR+req.FileNew, BASE_PATH+email+PATH_OWN_FILES+req.Folder+SEPARATOR+req.FileOld)
		return errUpdate
	}
	return nil
}

func (db *database) DoctorOpenFile(req OpenFileRequest) (string, error) {
	email := LogedUsers[req.Token].Email
	if _, err := os.Stat(BASE_PATH + email + PATH_OWN_FILES + req.Folder + SEPARATOR + req.File); os.IsNotExist(err) {
		return "", errors.New(ERROR_FILE_NOT_EXISTS)
	}
	files, inMap := OpenedFiles[req.Token]
	if inMap {
		// el doctor tiene lagun archivo abierto
		for key, value := range files {
			fmt.Println("valor " + string(key) + " " + value)
			if s.Contains(value, req.Folder+SEPARATOR+req.File) {
				return "", errors.New(ERROR_FILE_ALREADY_OPENED)
			}
		}
		// ninguno de los archivos abierto coincide, lo agrego
		OpenedFiles[req.Token] = append(OpenedFiles[req.Token], email+PATH_OWN_FILES+req.Folder+SEPARATOR+req.File)
		return BASE_PATH + email + PATH_OWN_FILES + req.Folder + SEPARATOR + req.File, nil
	}
	// no tiene archivos abiertos, creo un nuevo arreglo
	OpenedFiles[req.Token] = append(OpenedFiles[req.Token], email+PATH_OWN_FILES+req.Folder+SEPARATOR+req.File)
	return BASE_PATH + email + PATH_OWN_FILES + req.Folder + SEPARATOR + req.File, nil
}

//elimina un objeto en una posicion del arreglo
func remove(s []string, i int) []string {
	s[len(s)-1], s[i] = s[i], s[len(s)-1]
	return s[:len(s)-1]
}

func (db *database) DoctorCloseFile(req CloseFileRequest) error {
	email := LogedUsers[req.Token].Email
	if _, err := os.Stat(BASE_PATH + email + PATH_OWN_FILES + req.Folder + SEPARATOR + req.File); os.IsNotExist(err) {
		return errors.New(ERROR_FILE_NOT_EXISTS)
	}
	files, inMap := OpenedFiles[req.Token]
	if inMap {
		// tiene algun archivo abierto
		for key, value := range files {
			fmt.Println("value:" + value)
			if s.Contains(value, req.Folder+SEPARATOR+req.File) {
				// elimino el archivo del array
				OpenedFiles[req.Token] = remove(files, key)
				fmt.Println(OpenedFiles[req.Token])
				return nil
			}
		}
	}
	return errors.New(ERROR_FILE_NOT_OPENED)
}
