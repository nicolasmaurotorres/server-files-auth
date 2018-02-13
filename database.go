package main

import (
	"errors"
	"log"
	"os"
	"strconv"

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

func GetUserByEmail(email string, cat int8) (User, error) {
	var userToReturn User
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
		return User{Name: "", Email: ""}, errors.New(ERROR_NOT_EXISTING_USER)
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
	valid, err := IsValidToken(req.Token, true)
	if !valid {
		return err
	}
	email := LogedUsers[req.Token].Email
	errCreate := os.Mkdir(BASE_PATH+email+PATH_OWN_FILES+req.Name, MODE_PERMITIONS) //checkeo si puedo crear la carpeta
	if errCreate != nil {
		return errCreate // no se pudo crear la carpeta
	}
	var newFolder Directory
	newFolder.Files = make([]File, 1)
	newFolder.Path = email + PATH_OWN_FILES + req.Name
	session := getSession()
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
	valid, err := IsValidToken(req.Token, true)
	if !valid {
		return err
	}
	email := LogedUsers[req.Token].Email
	errDel := os.RemoveAll(BASE_PATH + "/" + email + PATH_OWN_FILES + req.Folder)
	if errDel != nil {
		return errDel
	}
	session := getSession()
	collection := session.DB(DATABASE_NAME).C(COLLECTION_USERS)
	query := bson.M{"email": email}
	update := bson.M{"$pull": bson.M{"directorys": bson.M{"path": email + PATH_OWN_FILES + req.Folder}}}
	errUpdate := collection.Update(query, update)
	if errUpdate != nil {
		return errUpdate
	}
	return nil
}
