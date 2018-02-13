Referential Links:
- https://www.digitalocean.com/community/tutorials/como-instalar-mongodb-en-ubuntu-16-04-es
- https://medium.com/@patdhlk/how-to-install-go-1-9-1-on-ubuntu-16-04-ee64c073cd79
- https://www.thepolyglotdeveloper.com/2017/03/authenticate-a-golang-api-with-json-web-tokens/
- https://thenewstack.io/make-a-restful-json-api-go/

**** Installation of MongoDB ****

- sudo apt-key adv --keyserver hkp://keyserver.ubuntu.com:80 --recv EA312927
- echo "deb http://repo.mongodb.org/apt/ubuntu xenial/mongodb-org/3.2 multiverse" | sudo tee /etc/apt/sources.list.d/mongodb-org-3.2.list
- sudo apt-get update
- sudo apt-get install -y mongodb-org
- sudo nano /etc/systemd/system/mongodb.service
	[Unit]
	Description=High-performance, schema-free document-oriented database
	After=network.target

	[Service]
	User=mongodb
	ExecStart=/usr/bin/mongod --quiet --config /etc/mongod.conf

	[Install]
	WantedBy=multi-user.target
- sudo systemctl start mongodb  # StartMongoDB
- sudo systemctl status mongodb # Check MongoDB Status
- sudo systemctl enable mongodb # Autostart on start up

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

- in the console you have to add the default user admin/admin 
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
		"files":
		[{
			"name":""
		}]
	},
	{
		"path":"doctor@doctor.com/modified/",
		"files":
		[{
			"name":""
		}]
	}
	]})


