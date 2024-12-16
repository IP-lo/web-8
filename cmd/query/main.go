package main

import (
	"database/sql"
	"fmt"
	"log"
	"net/http"

	_ "github.com/lib/pq"
)

const (
	host     = "localhost"
	port     = 5432
	user     = "postgres"
	password = "BKMZ661248082"
	dbname   = "Sandbox"
)

type Handlers struct {
	dbProvider DatabaseProvider
}

type DatabaseProvider struct {
	db *sql.DB
}

func (h *Handlers) postQuery(w http.ResponseWriter, r *http.Request) {
	name := r.URL.Query().Get("name")
	if name == "" {
		http.Error(w, "IT'S EMPTY: ", http.StatusBadRequest)
		return
	}
	err := h.dbProvider.insertName(name)
	if err != nil {
		http.Error(w, "Database error: "+err.Error(), http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("name is post"))
}

func (h *Handlers) GetQuery(w http.ResponseWriter, r *http.Request) {
	name := r.URL.Query().Get("name")
	if name == "" {
		fmt.Fprintf(w,"name is empty")
		return
	}
	err := h.dbProvider.SelectName(name)
	if err != nil {
		fmt.Fprint(w, "user does not exist!")
		return
	}
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("Hello, " + name + "!"))
	
}

func (dp *DatabaseProvider) SelectName(name string) error {
	var exist string
	err := dp.db.QueryRow("SELECT name FROM query WHERE name = ($1)",name).Scan(&exist)
	if err != nil {
		return err
	}
	return nil
}

func (dp *DatabaseProvider) insertName(name string) error {
	_, err := dp.db.Exec("INSERT INTO query (name) VALUES ($1) ", name)
	return err
}

func main() {
	psqlInfo := fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=disable", host, port, user, password, dbname)
	db, err := sql.Open("postgres", psqlInfo)
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	_, err = db.Exec(`
	CREATE TABLE IF NOT EXISTS query (
        id SERIAL PRIMARY KEY,
        name VARCHAR(50)
	)
	`)
	if err != nil {
		log.Fatal(err)
	}
	dp := DatabaseProvider{db: db}
	h := Handlers{dbProvider: dp}

	http.HandleFunc("/api/user/get", h.GetQuery)
	http.HandleFunc("/api/user/post", h.postQuery)

	if err := http.ListenAndServe(":9000", nil); err != nil{
		log.Println("Серверная ошибка:",err)
	}
}