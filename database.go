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
	DB_NAME          = "tesis"
	COLLECTION_USERS = "users"
	BASE_PATH        = "/home/maro/Desktop/data/pvw/"
)

type Directory struct {
	Path  string   `json:"path"`
	Files []string `json:"files"`
}

type Directorys []Directory

func getSession() *mgo.Session {
	session, err := mgo.Dial("localhost:27017")
	if err != nil {
		panic(err)
	}
	return session
}

type User struct {
	Name       string
	Email      string
	Password   string
	Directorys []Directory
}

type UserDoctorDBO struct {
	Name      string
	Email     string
	Password  string
	Directory Directorys
}

type UserPlademaDBO struct {
	Name     string
	Email    string
	Password string
}

type UserAdminDBO struct {
	Email    string
	Password string
}

// precondicion el email NO EXISTE
func NewUserDAL(user NewUserRequest) error {
	session := getSession()
	defer session.Close()
	if ExistsEmail(user.Email) {
		return errors.New(ERROR_EMAIL_ALREADY_EXISTS)
	}
	collection := session.DB(DB_NAME).C(COLLECTION_USERS)
	switch user.Category {
	case REQUEST_DOCTOR:
		var userDoctor UserDoctorDBO
		userDoctor.Name = user.Name
		userDoctor.Email = user.Email
		userDoctor.Password = user.Password
		os.Mkdir(BASE_PATH+userDoctor.Email, 0666)
		var directory Directory
		directory.Path = BASE_PATH + userDoctor.Email
		directory.Files = make([]string, 1)
		directory.Files[0] = ""
		var directorys Directorys
		directorys = append(directorys, directory)
		userDoctor.Directory = directorys
		err := collection.Insert(userDoctor)
		if err != nil {
			log.Fatal(ERROR_INSERT_NEW_DOCTOR)
			return errors.New(ERROR_INSERT_NEW_DOCTOR)
		}
		return nil
	case REQUEST_PLADEMA:
		var userPladema UserPlademaDBO
		userPladema.Email = user.Email
		userPladema.Name = user.Name
		userPladema.Password = user.Password
		err := collection.Insert(userPladema)
		if err != nil {
			log.Fatal(ERROR_INSERT_NEW_PLADEMA)
			return errors.New(ERROR_INSERT_NEW_PLADEMA)
		}
		return nil
	case REQUEST_ADMIN:
		log.Fatal(ERROR_NEW_ADMIN)
		return errors.New(ERROR_NEW_ADMIN)
	default:
		log.Fatal(ERROR_SERVER)
		return errors.New(ERROR_SERVER)
	}
}

func ExistsEmail(email string) bool {
	session := getSession()
	defer session.Close()
	collection := session.DB(DB_NAME).C(COLLECTION_USERS)
	var userResult User
	err := collection.Find(bson.M{"email": "{$eq:" + email + "}"}).One(&userResult)
	if err != nil {
		log.Fatal(err)
		return false
	}
	return true
}

func GetUserByEmail(email string, cat int8) (User, error) {
	var userToReturn User
	session := getSession()
	defer session.Close()
	collection := session.DB(DB_NAME).C(COLLECTION_USERS)
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

func ExistsPathForUser(path string, email string) bool {
	return false
}

func ExistsFileInPathUser(path string, file string, email string) bool {
	return false
}

func AddNewFileToPath(path string, file string) error {
	return nil
}
