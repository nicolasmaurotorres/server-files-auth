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
	BASE_PATH        = "/home/maro/Desktop/data/pvw/data/"
)

/*

marowark@gmail.com/pepe/archivo1.txt
marowark@gmail.com/pepo/archivo1.txt
marowark@gmail.com/pepe/archivo2.txt

	Name       :marowark
	Email      :marowark@gmail.com
	Password   :123456
	Directorys [{
					Path:pepe,
					Files : [{
						Name : archivo1.txt
					},{
						Name: archivo2.txt
					}]
				},
				{
					Path:pepo,
					Files : [
						Name: archivo1.txt
					]
				}]
*/

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
	collection := session.DB(DB_NAME).C(COLLECTION_USERS)
	switch user.Category {
	case REQUEST_DOCTOR:
		var userDoctor UserDoctorDBO
		userDoctor.Name = user.Name
		userDoctor.Email = user.Email
		userDoctor.Password = user.Password
		userDoctor.Category = REQUEST_DOCTOR
		os.Mkdir(BASE_PATH+userDoctor.Email, 0644)
		userDoctor.Folders = make([]Directory, 1)
		userDoctor.Folders[0].Files = make([]File, 1)
		userDoctor.Folders[0].Path = user.Email
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
	collection := session.DB(DB_NAME).C(COLLECTION_USERS)
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
