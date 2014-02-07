package main

import (
	"runtime"
	"code.google.com/p/go.crypto/bcrypt"
	"github.com/duanebester/go-restful"
	"flag"
	"strings"
	"crypto/rand"
	"encoding/base64"
	"log"
	"net/http"
	"path"
	"database/sql"
	_ "github.com/go-sql-driver/mysql"
)

const (
	// Database
	databaseType       = "mysql"
	connectionString   = "duanebester:du4n3b3s@tcp(127.0.0.1:3306)/monty"
	usersSelect        = "SELECT id, name, last, password, email, created FROM user ORDER BY created ASC LIMIT ?"
	userSelect         = "SELECT id, name, last, password, email, created FROM user WHERE id = ?"
	//userSelectByEmail  = "SELECT id, name, last, password, email FROM user WHERE email = ? LIMIT 1"
	userInsert         = "INSERT INTO user(name,last,password,email) VALUES(?,?,?,?)"

	agentInsert         = "INSERT INTO agent(name,user_id,secret,appkey) VALUES(?,?,?,?)"
	agentSelect         = "SELECT id, name, user_id, secret, appkey, created FROM agent WHERE id = ?"

	
	maxConnectionCount = 256
)

var (
	// Database Statements
	userInsertStatement        *sql.Stmt
	userSelectStatement        *sql.Stmt

	usersSelectStatement       *sql.Stmt

	agentInsertStatement       *sql.Stmt
	agentSelectStatement       *sql.Stmt

	// App dir for resources
	rootDir string
)

type User struct {
	Id        uint16 `json:"id"`
	Name      string `json:"name"`
	Last      string `json:"last"`
	Email     string `json:"email"`
	Created   string `json:"created"`
	Password  string `json:"password"`
	LastLogin string `json:"lastlogin"`
}

type Alert struct {
	Id       uint32 `json:"id"`
	Level     uint8 `json:"level"`
	AgentId  uint32 `json:"agentid"`
	Message  string `json:"message"`
	Created  string `json:"created"`
	Category string `json:"category"`
}

type Agent struct {
	Id            uint32 `json:"id"`
	UserId        uint16 `json:"userid"`
	Name          string `json:"name"`
	Secret        string `json:"secret"`
	Appkey        string `json:"appkey"`
	Created       string `json:"created"`

func init() {
	// Command Line Arguments:
	flag.StringVar(&rootDir, "root-dir", "/Users/duanebester/go/src/service", "The root dir where project source is located.")
}

func prepareDatabase() {

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

	usersSelectStatement, err = db.Prepare(usersSelect)
	if err != nil { log.Fatal(err) }

	agentInsertStatement, err = db.Prepare(agentInsert)
	if err != nil { log.Fatal(err) }

	agentSelectStatement, err = db.Prepare(agentSelect)
	if err != nil { log.Fatal(err) }
}

func main() {

	flag.Parse()

	// Max powah
	runtime.GOMAXPROCS(runtime.NumCPU())

	// Start Database
	prepareDatabase()

	// Web Service
	wsContainer := restful.NewContainer()
	initWS := initStatic()
	apiWS := apiService()
	wsContainer.Add(initWS).EnableContentEncoding(true)
	wsContainer.Add(apiWS)
	log.Println("Listening ... 8080")
	log.Fatal(http.ListenAndServe(":8080", wsContainer))

}

func getAuthCode(req *restful.Request, response *restful.Response) {
	potential := randomString(16)

	response.WriteHeader(http.StatusFound)
	response.WriteEntity(potential)
}

// randomString generates authorization codes or tokens with a given strength.
func randomString(strength int) string {
	s := make([]byte, strength)
	if _, err := rand.Read(s); err != nil {
		return ""
	}
	return strings.TrimRight(base64.URLEncoding.EncodeToString(s), "=")
}


// Static Resources
func initStatic() *restful.WebService {
	staticWS := new(restful.WebService)
	staticWS.Filter(staticLog)
	staticWS.Route(staticWS.GET("/").To(serveIndex))
	staticWS.Route(staticWS.GET("/{static}").To(serveStatic))
	return staticWS
}

// Service API
func apiService() *restful.WebService {
	ws := new(restful.WebService)

	ws.
		Path("/api").
		Consumes(restful.MIME_JSON).
		Produces(restful.MIME_JSON)

    ws.Filter(apiAuth)

    ws.Route(ws.GET("/users/{limit}").To(getUsers))
	ws.Route(ws.POST("/user").To(createUser))
	ws.Route(ws.GET("/user/{id}").To(getUserById))

	ws.Route(ws.POST("/agent").To(createAgent))
	ws.Route(ws.GET("/agent/{id}").To(getAgentById))

	ws.Route(ws.GET("/code/new").To(getAuthCode))

	return ws
}

// TODO: RateLimitFilter

func apiAuth(req *restful.Request, resp *restful.Response, chain *restful.FilterChain) {
	
	encoded := req.Request.Header.Get("Authorization")
	
	if len(encoded) == 0 {
		resp.AddHeader("WWW-Authenticate", "Basic realm=Protected Area")
		resp.WriteErrorString(401, "401: Not Authorized")
		log.Println("Error")
		return
	}

	_, err := base64.StdEncoding.DecodeString(encoded[6:])
	if err != nil {
		log.Println("error:", err)
		return
	}

	chain.ProcessFilter(req, resp)
}

// WebService Filter
func staticLog(req *restful.Request, resp *restful.Response, chain *restful.FilterChain) {
	log.Printf("[Static] %s - %s\n", req.Request.Method, req.Request.URL)
	chain.ProcessFilter(req, resp)
}

func getUsers(req *restful.Request, resp *restful.Response) {

	limit := req.PathParameter("limit")
	users := make([]User, 0, 10)
	rows, err := SelectUsers().Query(limit)

	var tempUser User
	defer rows.Close()

	for rows.Next() {
		err := rows.Scan(&tempUser.Id, &tempUser.Name, &tempUser.Last, &tempUser.Password, &tempUser.Email, &tempUser.Created)
		if err != nil {
			log.Fatal(err)
		}
		users = append(users, tempUser)
	}
	err = rows.Err()
	if err != nil {
		log.Fatal(err)
	}

	resp.WriteHeader(http.StatusFound)
	resp.WriteEntity(users)
}

// Returns a User from an ID
func getUserById(req *restful.Request, response *restful.Response) {
	userId := req.PathParameter("id")
	var user User

	err := SelectUser().QueryRow(userId).Scan(&user.Id, &user.Name, &user.Last, &user.Password, &user.Email, &user.Created)
	if err != nil {
		log.Fatal(err)
	}

	response.WriteHeader(http.StatusFound)
	response.WriteEntity(user)
}

// Creates a User from a ajax JSON user object
func createUser(req *restful.Request, response *restful.Response) {

	user := User{Id: 0}
	
	parseErr := req.ReadEntity(&user)
	if parseErr == nil {

		// Hash the password
		hashed, err := bcrypt.GenerateFromPassword([]byte(user.Password), bcrypt.DefaultCost)
		if err != nil {
			panic(err)
		}

		// Insert to database
		res, err := InsertUser().Exec(user.Name, user.Last, hashed, user.Email)
		if err != nil {
			log.Fatal(err)
		}

		// Grab last ID
		lastId, err := res.LastInsertId()
		if err != nil {
			log.Fatal(err)
		}

		log.Println("New User ID: ", lastId)

		response.WriteHeader(http.StatusCreated)
		response.WriteEntity(lastId)
	} else {
		log.Fatal(parseErr.Error())
		response.AddHeader("Content-Type", "text/plain")
		response.WriteErrorString(http.StatusInternalServerError, parseErr.Error())
	}
}

// Returns a Agent from an ID
func getAgentById(req *restful.Request, response *restful.Response) {
	agentId := req.PathParameter("id")
	var agent Agent
	// "SELECT id, name, owner, secret, appkey, created FROM agent WHERE id = ?"
	err := SelectAgent().QueryRow(agentId).Scan(&agent.Id, &agent.Name, &agent.UserId, &agent.Secret, &agent.Appkey, &agent.Created)
	if err != nil {
		log.Fatal(err)
	}

	response.WriteHeader(http.StatusFound)
	response.WriteEntity(agent)
}

// Creates an Agent from a ajax JSON user object
func createAgent(req *restful.Request, response *restful.Response) {

	agent := Agent{Id: 0, Secret: randomString(24), Appkey: randomString(16)}
	
	parseErr := req.ReadEntity(&agent)

	if parseErr == nil {

		// Insert to database
		// INSERT INTO agent(name,user_id,secret,appkey) VALUES(?,?,?,?)
		res, err := InsertAgent().Exec(agent.Name, agent.UserId, agent.Secret, agent.Appkey)
		if err != nil {
			log.Fatal(err)
		}

		// Grab last ID
		lastId, err := res.LastInsertId()
		if err != nil {
			log.Fatal(err)
		}

		log.Println("New Agent ID: ", lastId)

		response.WriteHeader(http.StatusCreated)
		response.WriteEntity(agent.Appkey)
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

// Serve Static 
func serveStatic(req *restful.Request, resp *restful.Response) {
	filename := req.PathParameter("static")
	//log.Println(filename)
	http.ServeFile(
		resp.ResponseWriter,
		req.Request,
		path.Join(rootDir, filename))
}

func InsertUser() *sql.Stmt {
	return userInsertStatement
}

func InsertAgent() *sql.Stmt {
	return agentInsertStatement
}

func SelectUser() *sql.Stmt {
	return userSelectStatement
}

func SelectUsers() *sql.Stmt {
	return usersSelectStatement
}

func SelectAgent() *sql.Stmt {
	return agentSelectStatement
}


















