package main

import (
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"strconv"
	s "strings"

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

type File struct {
	Name string
}

type Directory struct {
	Path  string `json:"path"`
	Files []File `json:"files"`
}

type User struct {
	Name     string
	Email    string
	Password string
}

type UserDoctorDBO struct {
	Category int8
	Name     string
	Email    string
	Password string
	Folders  []Directory
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

func getSession() *mgo.Session {
	session, err := mgo.Dial("localhost:27017")
	if err != nil {
		panic(err)
	}
	return session
}

// precondicion el email NO EXISTE
func NewUserDAL(user NewUserRequest) error {
	session := getSession()
	defer session.Close()
	if ExistsEmail(user.Email) {
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
		userDoctor.Folders = make([]Directory, 2)
		userDoctor.Folders[0].Files = make([]File, 1)
		userDoctor.Folders[0].Path = userDoctor.Email + PATH_OWN_FILES
		userDoctor.Folders[1].Files = make([]File, 1)
		userDoctor.Folders[1].Path = userDoctor.Email + PATH_MODIFIED_FILES
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

func ExistsEmail(email string) bool {
	session := getSession()
	defer session.Close()
	collection := session.DB(DATABASE_NAME).C(COLLECTION_USERS)
	var userResult User
	//	fmt.Println("llegue ExistsE mail")
	err := collection.Find(bson.M{"email": email}).One(&userResult)
	if err != nil {
		return false
	}
	return true
}

func GetUserByEmail(email string, cat int8) (UserDoctorDBO, error) {
	var userToReturn UserDoctorDBO // this user is the most general of the three
	session := getSession()
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

func ExistsFolderUser(path string, email string) bool {
	return false
}

func ExistsFileInPathUser(path string, file string, email string) bool {
	return false
}

func AddNewFileToPath(path string, file string) error {
	return nil
}

func CreateFolder(req AddFolderRequest) error {
	email := LogedUsers[req.Token].Email
	errCreate := os.Mkdir(BASE_PATH+email+PATH_OWN_FILES+req.Name, MODE_PERMITIONS) //checkeo si puedo crear la carpeta
	if errCreate != nil {
		return errCreate // no se pudo crear la carpeta
	}
	var newFolder Directory
	newFolder.Files = make([]File, 1)
	newFolder.Path = email + PATH_OWN_FILES + req.Name
	session := getSession()
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

func DeleteFolder(req DelFolderRequest) error {
	email := LogedUsers[req.Token].Email
	errDel := os.RemoveAll(BASE_PATH + "/" + email + PATH_OWN_FILES + req.Folder)
	if errDel != nil {
		return errDel
	}
	session := getSession()
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

func DeleteUser(user DelUserRequest) error {
	session := getSession()
	defer session.Close()
	collection := session.DB(DATABASE_NAME).C(COLLECTION_USERS)
	query := bson.M{"email": user.Email}
	userDeleted, _ := GetUserByEmail(user.Email, user.Category)
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

func RenameFolderDB(req RenameFolderRequest) error {
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
	session := getSession()
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

func AddFileDoctorDB(req AddFileDoctorRequest, r *http.Request) error {
	email := LogedUsers[req.Token].Email
	r.ParseMultipartForm(500 << 20) // 500MB max file size
	file, handler, err := r.FormFile("file")
	if err != nil {
		fmt.Println(err)
		return err
	}
	defer file.Close()
	newFile := BASE_PATH + email + SEPARATOR + PATH_OWN_FILES + req.Folder + SEPARATOR + handler.Filename
	if _, err := os.Stat(newFile); err == nil {
		return errors.New(ERROR_FILE_ALREADY_EXISTS) // the file already exists
	}
	f, err := os.OpenFile(newFile, os.O_WRONLY|os.O_CREATE, 0666)
	if err != nil {
		fmt.Println(err)
		return err
	}
	defer f.Close()
	_, errCopy := io.Copy(f, file) // copia el archivo "file" del form a "f"
	if errCopy != nil {
		return errCopy //archivo duplicado en la carpeta
	}
	session := getSession()
	collection := session.DB(DATABASE_NAME).C(COLLECTION_USERS)
	query := make(map[string]string)
	query["email"] = email
	query["directorys.path"] = email + SEPARATOR + PATH_OWN_FILES + req.Folder // carpeta a agregar el archivo
	update := bson.M{"$push": bson.M{"files": bson.M{"name": handler.Filename}}}
	errUpdate := collection.Update(query, update)
	if errUpdate != nil {
		return errUpdate //TODO: tengo que eliminar el archivo que guarde
	}

	return nil
}
