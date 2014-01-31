go-mysql-react
==============

Go-restful working with react, mysql + bcrypt, and bootstrap

	go get "code.google.com/p/go.crypto/bcrypt"
	go get "github.com/emicklei/go-restful"
	go get "github.com/go-sql-driver/mysql"

Make sure to change the connect string:

	"USER:PASSWORD@tcp(127.0.0.1:3306)/DATABASE"

Running:

	go build service.go
	./service

MySQL User Table:

	id: int(16)
	name: varchar(128)
	last: varchar(128)
	email: varchar(128)
	password: varchar(255)
