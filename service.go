package main

import (
	"database/sql"
	"runtime"
	"code.google.com/p/go.crypto/bcrypt"
	"github.com/duanebester/go-restful"
	"flag"
	"log"
	"net/http"
	"path"
	_ "github.com/go-sql-driver/mysql"
)

const (
	// Database
	databaseType       = "mysql"
	connectionString   = "USER:PASSWORD@tcp(127.0.0.1:3306)/DATABASE"
	userSelect         = "SELECT id, name, last, password, email FROM user WHERE id = ?"
	userSelectByEmail  = "SELECT id, name, last, password, email FROM user WHERE email = ? LIMIT 1"
	userInsert         = "INSERT INTO user(name,last,password,email) VALUES(?,?,?,?)"
	// worldUpdate        = "UPDATE World SET randomNumber = ? WHERE id = ?"
	maxConnectionCount = 256
)

var (
	// Templates -- I'll get to this later
	// tmpl = template.Must(template.ParseFiles("templates/layout.html", "templates/fortune.html"))

	// Database Statements
	userInsertStatement        *sql.Stmt
	userSelectStatement        *sql.Stmt
	userSelectByEmailStatement *sql.Stmt
)

type User struct {
	Id       uint16 `json:"id"`
	Name     string `json:"name"`
	Last     string `json:"last"`
	Email    string `json:"email"`
	Password string `json:"password"`
}

var rootDir string
var dbUser string
var dbPassword string
var dbName string

func init() {
	// Command Line Args
	flag.StringVar(&rootDir, "root-dir", "/Users/duanebester/go/src/httptest", "specifies the root dir where html and other files will be relative to")
	flag.StringVar(&dbName, "db-name", "DBNAME", "MySQL Database Name")
	flag.StringVar(&dbUser, "db-user", "USERNAME", "MySQL Database Username")
	flag.StringVar(&dbPassword, "db-password", "PASSWORD", "MySQL Database Password")
}

func main() {

	flag.Parse()

	// Max powah
	runtime.GOMAXPROCS(runtime.NumCPU())

	// Database Setup
	db, err := sql.Open(databaseType, connectionString)
	if err != nil {
		log.Fatalf("Error opening database: %v", err)
	}

	db.SetMaxIdleConns(maxConnectionCount)

	// Prepared Database Queries
	userInsertStatement, err = db.Prepare(userInsert)
	if err != nil { log.Fatal(err) }

	userSelectStatement, err = db.Prepare(userSelect)
	if err != nil { log.Fatal(err) }

	userSelectByEmailStatement, err = db.Prepare(userSelectByEmail)
	if err != nil { log.Fatal(err) }

	// Web Service
	wsContainer := restful.NewContainer()
	initWS := initStatic()
	userWS := userService()
	wsContainer.Add(initWS).EnableContentEncoding(true)
	wsContainer.Add(userWS)
	log.Println("Listening ... 8080")
	log.Fatal(http.ListenAndServe(":8080", wsContainer))

}

// Static Resources
func initStatic() *restful.WebService {
	staticWS := new(restful.WebService)
	staticWS.Route(staticWS.GET("/").To(serveIndex))
	staticWS.Route(staticWS.GET("/index").To(serveIndex))
	return staticWS
}

// User Service API
func userService() *restful.WebService {
	ws := new(restful.WebService)

	ws.
		Path("/api").
		Consumes(restful.MIME_JSON).
		Produces(restful.MIME_JSON)

	ws.Route(ws.POST("/users/new").To(createUser))
	ws.Route(ws.GET("/users/{id}").To(getUserById))
	//ws.Route(ws.GET("/users/{email}/{last-page}").To(undef))

	return ws
}

// Returns a User from an ID
func getUserById(request *restful.Request, response *restful.Response) {
	userId := request.PathParameter("id")
	var user User

	err := userSelectStatement.QueryRow(userId).Scan(&user.Id, &user.Name, &user.Last, &user.Password, &user.Email)
	if err != nil {
		log.Fatal(err)
	}

	response.WriteHeader(http.StatusFound)
	response.WriteEntity(user)
}

// Creates a User from a ajax JSON user object
func createUser(request *restful.Request, response *restful.Response) {

	user := User{Id: 0}
	
	parseErr := request.ReadEntity(&user)
	if parseErr == nil {

		// Hash the password
		hashed, err := bcrypt.GenerateFromPassword([]byte(user.Password), bcrypt.DefaultCost)
		if err != nil {
			panic(err)
		}

		// Insert to database
		res, err := userInsertStatement.Exec(user.Name, user.Last, hashed, user.Email)
		if err != nil {
			log.Fatal(err)
		}

		// Grab last ID
		lastId, err := res.LastInsertId()
		if err != nil {
			log.Fatal(err)
		}

		ret := string(lastId)

		response.WriteHeader(http.StatusCreated)
		response.WriteEntity(ret)
	} else {
		log.Fatal(parseErr.Error())
		response.AddHeader("Content-Type", "text/plain")
		response.WriteErrorString(http.StatusInternalServerError, parseErr.Error())
	}
}

// Serve index.html
func serveIndex(req *restful.Request, resp *restful.Response) {
	http.ServeFile(
		resp.ResponseWriter,
		req.Request,
		path.Join(rootDir, "index.html"))
}


