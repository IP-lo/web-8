package main

import (
	"database/sql"
	"fmt"
	"log"
	"net/http"
	"strconv"

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

func (h *Handlers) postCount(w http.ResponseWriter, r *http.Request) {
	r.ParseForm()
	val, err := strconv.Atoi(r.FormValue("count"))
	if err != nil {
		http.Error(w, "IT'S NOT NUMBER: "+err.Error(), http.StatusBadRequest)
		return
	}
	err = h.dbProvider.incrementCount(val)
	if err != nil {
		http.Error(w, "Database error: "+err.Error(), http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("введено значение " + strconv.Itoa(val)))
}

func (h *Handlers) GetCount(w http.ResponseWriter, r *http.Request) {
	val, err := h.dbProvider.SelectCount()
	if err != nil {
		http.Error(w, "Database error: "+err.Error(), http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("получено значение " + strconv.Itoa(val)))
}

func (dp *DatabaseProvider) SelectCount() (int, error) {
	var val int
	err := dp.db.QueryRow("SELECT value FROM count WHERE id = 1").Scan(&val)
	if err != nil {
		return 0, err
	}
	return val, nil
}

func (dp *DatabaseProvider) incrementCount(n int) error {
	_, err := dp.db.Exec("UPDATE count SET value = value + ($1) WHERE id = 1", n)
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
	CREATE TABLE IF NOT EXISTS count (
		id SERIAL PRIMARY KEY,
		value INTEGER NOT NULL DEFAULT 0
	)
	`)
	if err != nil {
		log.Fatal(err)
	}

	_, err = db.Exec("INSERT INTO count (id, value) VALUES (1, 0) ON CONFLICT (id) DO NOTHING")
	if err != nil {
		log.Fatal(err)
	}

	dp := DatabaseProvider{db: db}
	h := Handlers{dbProvider: dp}

	http.HandleFunc("/count", func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case "GET":
			h.GetCount(w, r)
		case "POST":
			h.postCount(w, r)
		default:
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	})

	if err := http.ListenAndServe(":3333", nil); err != nil {
		log.Fatal(err)
	}
}
