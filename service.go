package main

import (
	"code.google.com/p/go.crypto/bcrypt"
	"crypto/rand"
	"database/sql"
	"encoding/base64"
	"flag"
	"github.com/duanebester/go-restful"
	_ "github.com/go-sql-driver/mysql"
	"log"
	"net/http"
	"runtime"
	"strings"
)

const (
	// Database
	databaseType = "mysql"

	usersSelect = "SELECT id, name, last, password, email, created FROM user ORDER BY created ASC LIMIT ?"

	userInsert = "INSERT INTO user(name,last,password,email) VALUES(?,?,?,?)"
	userSelect = "SELECT id, name, last, password, email, created FROM user WHERE id = ?"

	agentInsert = "INSERT INTO agent(name,user_id,secret,appkey) VALUES(?,?,?,?)"
	agentSelect = "SELECT id, name, user_id, secret, appkey, created FROM agent WHERE id = ?"

	agentsSelectByUser = "SELECT id, name, user_id, created FROM agent WHERE user_id = ? ORDER BY created"

	agentSelectIdByKey = "SELECT id FROM agent WHERE appkey = ?"

	alertInsert = "INSERT INTO alert(message, category, level, agent_id) VALUES(?,?,?,?)"
	alertSelect = "SELECT id, message,category, level, agent_id, created FROM alert WHERE id = ?"

	alertsSelectByAgent = "SELECT id,message,category,level,agent_id, created FROM alert WHERE agent_id = ? ORDER BY created LIMIT ?"

	maxConnectionCount = 256
)

var (
	// Database Statements
	userInsertStatement *sql.Stmt
	userSelectStatement *sql.Stmt

	usersSelectStatement *sql.Stmt

	agentInsertStatement        *sql.Stmt
	agentSelectStatement        *sql.Stmt
	agentsSelectByUserStatement *sql.Stmt

	agentSelectIdByKeyStatement *sql.Stmt

	alertInsertStatement *sql.Stmt
	alertSelectStatement *sql.Stmt

	alertsSelectByAgentStatement *sql.Stmt

	// App dir for resources
	rootDir string

	// Command line args for DB
	dbUsername string
	dbName     string
	dbPassword string
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
	Level    uint8  `json:"level"`
	Appkey   string `json:"appkey"`
	AgentId  uint32 `json:"agentid"`
	Message  string `json:"message"`
	Created  string `json:"created"`
	Category string `json:"category"`
}

type Agent struct {
	Id      uint32 `json:"id"`
	UserId  uint16 `json:"userid"`
	Name    string `json:"name"`
	Secret  string `json:"secret"`
	Appkey  string `json:"appkey"`
	Created string `json:"created"`
}

func init() {
	// Command Line Arguments:
	// EX: ./service -db-user=SomeName -db-pass=SomePass -db-user=SomeUser
	flag.StringVar(&rootDir, "root-dir", "/Users/duanebester/go/src/service", "The root dir where project source is located.")
	flag.StringVar(&dbUsername, "db-user", "USERNAME", "Username to connect to your MySQL DB.")
	flag.StringVar(&dbName, "db-name", "DBNAME", "Name of your MySQL DB.")
	flag.StringVar(&dbPassword, "db-pass", "PASSWORD", "Password to connect to your MySQL DB.")
}

func prepareDatabase() {

	// Database Setup
	//connectionString   := "USER:PASS@tcp(127.0.0.1:3306)/DATABASE"

	connectionString := strings.Join([]string{dbUsername, ":", dbPassword, "@tcp(127.0.0.1:3306)/", dbName}, "")

	db, err := sql.Open(databaseType, connectionString)
	if err != nil {
		log.Fatalf("Error opening database: %v", err)
	}

	db.SetMaxIdleConns(maxConnectionCount)

	// Prepared Database Queries
	userInsertStatement, err = db.Prepare(userInsert)
	if err != nil {
		log.Fatal(err)
	}

	userSelectStatement, err = db.Prepare(userSelect)
	if err != nil {
		log.Fatal(err)
	}

	usersSelectStatement, err = db.Prepare(usersSelect)
	if err != nil {
		log.Fatal(err)
	}

	agentInsertStatement, err = db.Prepare(agentInsert)
	if err != nil {
		log.Fatal(err)
	}

	agentSelectStatement, err = db.Prepare(agentSelect)
	if err != nil {
		log.Fatal(err)
	}

	agentsSelectByUserStatement, err = db.Prepare(agentsSelectByUser)
	if err != nil {
		log.Fatal(err)
	}

	agentSelectIdByKeyStatement, err = db.Prepare(agentSelectIdByKey)
	if err != nil {
		log.Fatal(err)
	}

	alertInsertStatement, err = db.Prepare(alertInsert)
	if err != nil {
		log.Fatal(err)
	}

	alertSelectStatement, err = db.Prepare(alertSelect)
	if err != nil {
		log.Fatal(err)
	}

	alertsSelectByAgentStatement, err = db.Prepare(alertsSelectByAgent)
	if err != nil {
		log.Fatal(err)
	}
}

func main() {

	flag.Parse()

	// Max powah
	runtime.GOMAXPROCS(runtime.NumCPU())

	// Start Database
	prepareDatabase()

	// Web Service
	wsContainer := restful.NewContainer()
	apiWS := apiService()
	wpiWS := wpiService()
	wsContainer.Add(apiWS)
	wsContainer.Add(wpiWS)
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

// Service API
func apiService() *restful.WebService {
	ws := new(restful.WebService)

	ws.
		Path("/api").
		Consumes(restful.MIME_JSON).
		Produces(restful.MIME_JSON)

	ws.Filter(apiAuth).Filter(apiLog)

	ws.Route(ws.GET("/alerts").To(getAlertsByAgentKey))
	ws.Route(ws.POST("/alert").To(createAlert))

	return ws
}

// Private API for Web Apps to CRUD users / agents and such
func wpiService() *restful.WebService {
	ws := new(restful.WebService)

	ws.
		Path("/wpi").
		Consumes(restful.MIME_JSON).
		Produces(restful.MIME_JSON)

	ws.Filter(wpiAuth).Filter(apiLog)

	ws.Route(ws.GET("/users/{limit}").To(getUsers))
	ws.Route(ws.POST("/user").To(createUser))
	ws.Route(ws.GET("/user/{id}").To(getUserById))
	ws.Route(ws.GET("/user/{id}/agents").To(getAgentsByUser))

	ws.Route(ws.POST("/agent").To(createAgent))
	ws.Route(ws.GET("/agent/{id}").To(getAgentById))

	ws.Route(ws.GET("/agent/{id}/alerts").To(getAlertsByAgentId))

	ws.Route(ws.GET("/code/new").To(getAuthCode))

	return ws
}

func apiAuth(req *restful.Request, response *restful.Response, chain *restful.FilterChain) {

	appkey := req.Request.Header.Get("Authorization")

	log.Println(appkey)

	if len(appkey) == 0 {
		response.WriteHeader(http.StatusPaymentRequired)
		log.Println("Error")
		return
	}

	var agentId uint32

	err := SelectAgentIdByKey().QueryRow(appkey).Scan(&agentId)
	if err != nil {
		log.Fatal(err)
	}

	req.SetAttribute("agentId", agentId)

	chain.ProcessFilter(req, response)
}

func wpiAuth(req *restful.Request, response *restful.Response, chain *restful.FilterChain) {

	log.Println(req.Request.RemoteAddr)

	ip := strings.Split(req.Request.RemoteAddr, ":")[0]

	if len(ip) == 0 {
		response.WriteHeader(http.StatusPaymentRequired)
		log.Println("Error")
		return
	}

	chain.ProcessFilter(req, response)
}

// TODO: RateLimitFilter

// Logging Filter
func apiLog(req *restful.Request, resp *restful.Response, chain *restful.FilterChain) {
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

func getAgentsByUser(req *restful.Request, resp *restful.Response) {

	userId := req.PathParameter("id")
	var agents []Agent
	// "SELECT id, name, user_id, created FROM agent WHERE user_id = ? ORDER BY created"
	rows, err := SelectAgentsByUser().Query(userId)

	var temp Agent
	defer rows.Close()

	for rows.Next() {
		err := rows.Scan(&temp.Id, &temp.Name, &temp.UserId, &temp.Created)
		if err != nil {
			log.Fatal(err)
		}
		agents = append(agents, temp)
	}
	err = rows.Err()
	if err != nil {
		log.Fatal(err)
	}

	resp.WriteHeader(http.StatusFound)
	resp.WriteEntity(agents)
}

// TODO: UpdateAgentById

// TODO: DeleteAgentById

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

// Creates an Alert
func createAlert(req *restful.Request, response *restful.Response) {

	alert := Alert{Id: 0}

	parseErr := req.ReadEntity(&alert)

	if parseErr == nil {

		if req.Attribute("agentId") == nil {
			response.WriteHeader(http.StatusBadRequest)
			return
		}

		log.Println(alert.Message, alert.Category, alert.Level, req.Attribute("agentId"))
		// ~-----   "INSERT INTO alert(message, category, level, agent_id)"
		res, err := InsertAlert().Exec(alert.Message, alert.Category, alert.Level, req.Attribute("agentId"))
		if err != nil {
			log.Println(err)
			response.WriteHeader(http.StatusBadRequest)
			return
		}

		// Grab last ID
		lastId, err := res.LastInsertId()
		if err != nil {
			log.Fatal(err)
		}

		log.Println("New Alert ID: ", lastId)
		response.WriteHeader(http.StatusCreated)
		response.WriteEntity(lastId)

	} else {
		log.Fatal(parseErr.Error())
		response.AddHeader("Content-Type", "text/plain")
		response.WriteErrorString(http.StatusInternalServerError, parseErr.Error())
	}
}

func getAlertsByAgentKey(req *restful.Request, resp *restful.Response) {

	if req.Attribute("agentId") == nil {
		resp.WriteHeader(http.StatusBadRequest)
		return
	}

	alerts := make([]Alert, 0, 10)
	// "SELECT id,message,category,level,agent_id, created FROM alert WHERE agent_id = ? LIMIT ?"
	rows, err := SelectAlertsByAgent().Query(req.Attribute("agentId"), 10)

	var tempAlert Alert
	defer rows.Close()

	for rows.Next() {
		err := rows.Scan(&tempAlert.Id, &tempAlert.Message, &tempAlert.Category, &tempAlert.Level, &tempAlert.AgentId, &tempAlert.Created)
		if err != nil {
			log.Fatal(err)
		}
		alerts = append(alerts, tempAlert)
	}
	err = rows.Err()
	if err != nil {
		log.Fatal(err)
	}

	resp.WriteHeader(http.StatusFound)
	resp.WriteEntity(alerts)
}

func getAlertsByAgentId(req *restful.Request, resp *restful.Response) {

	agentId := req.PathParameter("id")
	alerts := make([]Alert, 0, 10)

	// "SELECT id,message,category,level,agent_id, created FROM alert WHERE agent_id = ? LIMIT ?"
	rows, err := SelectAlertsByAgent().Query(agentId, 10)

	var tempAlert Alert
	defer rows.Close()

	for rows.Next() {
		err := rows.Scan(&tempAlert.Id, &tempAlert.Message, &tempAlert.Category, &tempAlert.Level, &tempAlert.AgentId, &tempAlert.Created)
		if err != nil {
			log.Fatal(err)
		}
		alerts = append(alerts, tempAlert)
	}
	err = rows.Err()
	if err != nil {
		log.Fatal(err)
	}

	resp.WriteHeader(http.StatusFound)
	resp.WriteEntity(alerts)
}

func InsertUser() *sql.Stmt {
	return userInsertStatement
}

func InsertAgent() *sql.Stmt {
	return agentInsertStatement
}

func InsertAlert() *sql.Stmt {
	return alertInsertStatement
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

func SelectAgentsByUser() *sql.Stmt {
	return agentsSelectByUserStatement
}

func SelectAgentIdByKey() *sql.Stmt {
	return agentSelectIdByKeyStatement
}

func SelectAlert() *sql.Stmt {
	return alertSelectStatement
}

func SelectAlertsByAgent() *sql.Stmt {
	return alertsSelectByAgentStatement
}
