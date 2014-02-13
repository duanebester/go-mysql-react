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
	./service -db-name=YOUR_DB_NAME -db-user=USERNAME -db-pass=PASSWORD

MySQL User Table:

	+----------+--------------+------+-----+-------------------+
	| Field    | Type         | Null | Key | Default           |
	+----------+--------------+------+-----+-------------------+
	| id       | int(32)      | NO   | PRI | NULL              |
	| name     | varchar(128) | NO   |     | NULL              |
	| last     | varchar(128) | NO   |     | NULL              |
	| email    | varchar(128) | NO   | UNI | NULL              |
	| password | varchar(255) | NO   |     | NULL              |
	| created  | timestamp    | NO   |     | CURRENT_TIMESTAMP |
	+----------+--------------+------+-----+-------------------+

MySQL Agent Table

	+---------+-------------+------+-----+-------------------+
	| Field   | Type        | Null | Key | Default           |
	+---------+-------------+------+-----+-------------------+
	| id      | int(32)     | NO   | PRI | NULL              |
	| name    | varchar(45) | NO   |     | NULL              |
	| secret  | varchar(45) | NO   |     | NULL              |
	| appkey  | varchar(45) | NO   |     | NULL              |
	| created | timestamp   | NO   |     | CURRENT_TIMESTAMP |
	| user_id | int(32)     | NO   | MUL | NULL              |
	+---------+-------------+------+-----+-------------------+

MySQL Alert Table

	+----------+--------------+------+-----+-------------------+
	| Field    | Type         | Null | Key | Default           |
	+----------+--------------+------+-----+-------------------+
	| id       | int(32)      | NO   | PRI | NULL              |
	| created  | timestamp    | NO   |     | CURRENT_TIMESTAMP |
	| message  | varchar(512) | NO   |     | NULL              |
	| category | varchar(64)  | NO   |     | NULL              |
	| level    | int(8)       | NO   |     | NULL              |
	| agent_id | int(32)      | NO   | MUL | NULL              |
	+----------+--------------+------+-----+-------------------+

