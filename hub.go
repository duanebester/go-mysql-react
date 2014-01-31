package main

import 
(
	"database/sql"
	"encoding/json"
	"html/template"
	"log"
	"math/rand"
	"net/http"
	"runtime"
	"sort"
	"strconv"
	"code.google.com/p/go.crypto/bcrypt"
	_ "github.com/go-sql-driver/mysql"
)

type User struct 
{
	Id       uint16 `json:"id"`
	Name     string `json:"name"`
	Last     string `json:"last"`
	Email    string `json:"email"`
	Password string `json:"password"`
}

const 
(
	// Database
	databaseType       = "mysql"
	connectionString   = "duanebester:du4n3b3s@tcp(127.0.0.1:3306)/monty"
	userSelect         = "SELECT id, name, last, password, email FROM user WHERE id = ?"
	userSelectByEmail  = "SELECT id, name, last, password, email FROM user WHERE email = ? LIMIT 1"
	userInsert         = "INSERT INTO user(name,last,password,email) VALUES(?,?,?,?)"
	// worldUpdate        = "UPDATE World SET randomNumber = ? WHERE id = ?"
	maxConnectionCount = 256
)

var 
(
	// Templates
	// tmpl = template.Must(template.ParseFiles("templates/layout.html", "templates/fortune.html"))

	// Database
	userInsertStatement        *sql.Stmt
	userSelectStatement        *sql.Stmt
	userSelectByEmailStatement *sql.Stmt
)

func main() 
{
	runtime.GOMAXPROCS(runtime.NumCPU())

	db, err := sql.Open(databaseType, connectionString)
	if err != nil {
		log.Fatalf("Error opening database: %v", err)
	}
	db.SetMaxIdleConns(maxConnectionCount)
	userInsertStatement, err = db.Prepare(userInsert)
	if err != nil {
		log.Fatal(err)
	}
	userSelectStatement, err = db.Prepare(userSelect)
	if err != nil {
		log.Fatal(err)
	}
	userSelectByEmailStatement, err = db.Prepare(userSelectByEmail)
	if err != nil {
		log.Fatal(err)
	}

	http.HandleFunc("/users", usersHandler)

	http.HandleFunc("/db", dbHandler)
	http.HandleFunc("/queries", queriesHandler)
	http.HandleFunc("/json", jsonHandler)
	http.HandleFunc("/fortune", fortuneHandler)
	http.HandleFunc("/update", updateHandler)
	http.HandleFunc("/plaintext", plaintextHandler)
	http.ListenAndServe(":8080", nil)
}

// Get Users
func usersHandler(w http.ResponseWriter, r *http.Request) {
	n := 1
	if nStr := r.URL.Query().Get("queries"); len(nStr) > 0 {
		n, _ = strconv.Atoi(nStr)
	}

	w.Header().Set("Content-Type", "application/json")
	encoder := json.NewEncoder(w)

	if n <= 1 {
		var world World
		worldStatement.QueryRow(rand.Intn(worldRowCount)+1).Scan(&world.Id, &world.RandomNumber)
		world.RandomNumber = uint16(rand.Intn(worldRowCount) + 1)
		updateStatement.Exec(world.RandomNumber, world.Id)
		encoder.Encode(&world)
	} else {
		world := make([]World, n)
		for i := 0; i < n; i++ {
			if err := worldStatement.QueryRow(rand.Intn(worldRowCount)+1).Scan(&world[i].Id, &world[i].RandomNumber); err != nil {
				log.Fatalf("Error scanning world row: %s", err.Error())
			}
			world[i].RandomNumber = uint16(rand.Intn(worldRowCount) + 1)
			if _, err := updateStatement.Exec(world[i].RandomNumber, world[i].Id); err != nil {
				log.Fatalf("Error updating world row: %s", err.Error())
			}
		}
		encoder.Encode(world)
	}
}

// Test 1: JSON serialization
func jsonHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/javascript")
	json.NewEncoder(w).Encode(&Message{helloWorldString})
}

// Test 2: Single database query
func dbHandler(w http.ResponseWriter, r *http.Request) {
	var world World
	err := worldStatement.QueryRow(rand.Intn(worldRowCount)+1).Scan(&world.Id, &world.RandomNumber)
	if err != nil {
		log.Fatalf("Error scanning world row: %s", err.Error())
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(&world)
}

// Test 3: Multiple database queries
func queriesHandler(w http.ResponseWriter, r *http.Request) {
	n := 1
	if nStr := r.URL.Query().Get("queries"); len(nStr) > 0 {
		n, _ = strconv.Atoi(nStr)
	}

	if n <= 1 {
		dbHandler(w, r)
		return
	}

	world := make([]World, n)
	for i := 0; i < n; i++ {
		err := worldStatement.QueryRow(rand.Intn(worldRowCount)+1).Scan(&world[i].Id, &world[i].RandomNumber)
		if err != nil {
			log.Fatalf("Error scanning world row: %s", err.Error())
		}
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(world)
}

// Test 4: Fortunes
func fortuneHandler(w http.ResponseWriter, r *http.Request) {
	rows, err := fortuneStatement.Query()
	if err != nil {
		log.Fatalf("Error preparing statement: %v", err)
	}

	fortunes := make(Fortunes, 0, 16)
	for rows.Next() { //Fetch rows
		fortune := Fortune{}
		if err := rows.Scan(&fortune.Id, &fortune.Message); err != nil {
			log.Fatalf("Error scanning fortune row: %s", err.Error())
		}
		fortunes = append(fortunes, &fortune)
	}
	fortunes = append(fortunes, &Fortune{Message: "Additional fortune added at request time."})

	sort.Sort(ByMessage{fortunes})
	w.Header().Set("Content-Type", "text/html")
	if err := tmpl.Execute(w, fortunes); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

// Test 5: Database updates
func updateHandler(w http.ResponseWriter, r *http.Request) {
	n := 1
	if nStr := r.URL.Query().Get("queries"); len(nStr) > 0 {
		n, _ = strconv.Atoi(nStr)
	}

	w.Header().Set("Content-Type", "application/json")
	encoder := json.NewEncoder(w)

	if n <= 1 {
		var world World
		worldStatement.QueryRow(rand.Intn(worldRowCount)+1).Scan(&world.Id, &world.RandomNumber)
		world.RandomNumber = uint16(rand.Intn(worldRowCount) + 1)
		updateStatement.Exec(world.RandomNumber, world.Id)
		encoder.Encode(&world)
	} else {
		world := make([]World, n)
		for i := 0; i < n; i++ {
			if err := worldStatement.QueryRow(rand.Intn(worldRowCount)+1).Scan(&world[i].Id, &world[i].RandomNumber); err != nil {
				log.Fatalf("Error scanning world row: %s", err.Error())
			}
			world[i].RandomNumber = uint16(rand.Intn(worldRowCount) + 1)
			if _, err := updateStatement.Exec(world[i].RandomNumber, world[i].Id); err != nil {
				log.Fatalf("Error updating world row: %s", err.Error())
			}
		}
		encoder.Encode(world)
	}
}

// Test 6: Plaintext
func plaintextHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/plain")
	w.Write(helloWorldBytes)
}

type Fortunes []*Fortune

func (s Fortunes) Len() int      { return len(s) }
func (s Fortunes) Swap(i, j int) { s[i], s[j] = s[j], s[i] }

type ByMessage struct{ Fortunes }

func (s ByMessage) Less(i, j int) bool { return s.Fortunes[i].Message < s.Fortunes[j].Message }