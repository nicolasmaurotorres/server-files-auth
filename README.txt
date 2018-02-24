Referential Links:
- https://www.digitalocean.com/community/tutorials/como-instalar-mongodb-en-ubuntu-16-04-es
- https://medium.com/@patdhlk/how-to-install-go-1-9-1-on-ubuntu-16-04-ee64c073cd79
- https://www.thepolyglotdeveloper.com/2017/03/authenticate-a-golang-api-with-json-web-tokens/
- https://thenewstack.io/make-a-restful-json-api-go/

**** Installation of MongoDB 3.6****

- sudo apt-key adv --keyserver hkp://keyserver.ubuntu.com:80 --recv 2930ADAE8CAF5059EE73BB4B58712A2291FA4AD5
- echo "deb [ arch=amd64,arm64 ] https://repo.mongodb.org/apt/ubuntu xenial/mongodb-org/3.6 multiverse" | sudo tee /etc/apt/sources.list.d/mongodb-org-3.6.list
- sudo apt-get update
- sudo apt-get install -y mongodb-org
- sudo mkdir -p /data/db //carpeta para mongod
- sudo chown -R `id -u` /data/db // permisos para mongod
- sudo su // para activar el servicio de mongod
- nano /lib/systemd/system/mongodb.service
		[Unit]
		Description=MongoDB Database Service
		Wants=network.target
		After=network.target

		[Service]
		ExecStart=/usr/bin/mongod --config /etc/mongod.conf
		ExecReload=/bin/kill -HUP $MAINPID
		Restart=always
		User=mongodb
		Group=mongodb
		StandardOutput=syslog
		StandardError=syslog

		[Install]
		WantedBy=multi-user.target

- sudo systemctl start mongodb // start mongod
- sudo systemctl enable mongodb.service // start mongod on startup

*** DEFAULT DATA IN DB ***
- In the console you have to add the default user admin/admin 
	* mongo
	* use tesis
	* db.createCollection("users")
	* db.users.insert({"category":"2","email":"admin","password":"admin"})

- defaults users only for testing
	* db.users.insert({"category":"1","email":"pladema@pladema.com","password":"pladema","name":"pladema"})
	* db.users.insert({"category":"0","email":"doctor@doctor.com","password":"doctor","name":"doctor",
	"directorys":
	[{
		"path":"doctor@doctor.com/own/",
		"files":[""]
	},
	{
		"path":"doctor@doctor.com/modified/",
		"files":[""]
	}
	]})
	
**** Installation of Golang ****
- sudo apt-get update
- sudo apt-get -y upgrade
- sudo curl -O https://storage.googleapis.com/golang/go1.9.1.linux-amd64.tar.gz
- sudo tar -xvf go1.9.1.linux-amd64.tar.gz
- sudo mv go /usr/local
- mkdir $HOME/work
- sudo nano ~/.profile
	export PATH=$PATH:/usr/local/go/bin
	export GOPATH=$HOME/work
- source ~/.profile
- cd $GOPATH 
- mkdir src/hello
- nano src/hello/hello.go
	package main
	import "fmt"
	func main() {
    		fmt.Printf("Welcome To ITzGeek\n")
	}
	
** modifiyed file ** - snavarro89
- package gopkg.in/mgo.v2
- file "session.go" , add this funcion
func (c *Collection) UpdateArrayFilters(selector interface{}, update interface{}, arrayfilters []interface{}) error {

	if selector == nil {
		selector = bson.D{}
	}
	op := updateOp{
		Collection:   c.FullName,
		Selector:     selector,
		Update:       update,
		ArrayFilters: arrayfilters,
	}
	
	lerr, err := c.writeOp(&op, true)
	if err == nil && lerr != nil && !lerr.UpdatedExisting {
		return ErrNotFound
	}
	return err
	}
- file "socket.go", struct "updateOp" add this field 
ArrayFilters []interface{} `bson:"arrayFilters,omitempty"`
