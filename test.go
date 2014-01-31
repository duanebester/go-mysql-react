package main

import (
    "fmt"
    "log"
    //"io"
    //"github.com/emicklei/go-restful"
    "net/http"
    "code.google.com/p/go.crypto/bcrypt"
    "database/sql"
	_ "github.com/go-sql-driver/mysql"
)

var db *sql.DB

const pw = "p455w0rd"

func handler(w http.ResponseWriter, r *http.Request) {
	
    fmt.Fprintf(w, "Hi there, I love %s!", r.URL.Path[1:])
}

func main() {

	hashed, err := bcrypt.GenerateFromPassword([]byte(pw), bcrypt.DefaultCost)
	if err != nil {
		panic(err)
	}
	log.Println(bcrypt.DefaultCost)
	log.Println(string(hashed))
	err = bcrypt.CompareHashAndPassword(hashed, []byte(pw))
	log.Println(err)

	db, err := sql.Open("mysql",
		"duanebester:du4n3b3s@tcp(127.0.0.1:3306)/brave")
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	var (
		id int
		name string
	)
	rows, err := db.Query("select slinkid, name from cfg_user where slinkid = ?", 100)
	if err != nil {
		log.Fatal(err)
	}
	defer rows.Close()
	for rows.Next() {
		err := rows.Scan(&id, &name)
		if err != nil {
			log.Fatal(err)
		}
		log.Println(id, name)
	}
	err = rows.Err()
	if err != nil {
		log.Fatal(err)
	}

    http.HandleFunc("/", handler)
    http.ListenAndServe(":8080", nil)
}
