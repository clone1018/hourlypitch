package main

import (
	"database/sql"
	"html/template"
	"log"
	"net/http"

	_ "github.com/mattn/go-sqlite3"
)

var db *sql.DB
var rootTemplate *template.Template
var indexTemplate *template.Template
var submitTemplate *template.Template

type Idea struct {
	ID       int
	Pitch    string
	Created  int
	Approved int
	Shown    int
}

func grabDB() *sql.DB {
	db, err := sql.Open("sqlite3", "./db/hourlypitch.db")
	if err != nil {
		log.Fatal(err)
	}
	return db
}

func loadTemplates() {
	var err error
	indexTemplate, err = rootTemplate.ParseFiles("templates/index.html")
	if err != nil {
		log.Fatal(err)
	}
	submitTemplate, err = rootTemplate.ParseFiles("templates/submit.html")
	if err != nil {
		log.Fatal(err)
	}
}

func index(w http.ResponseWriter, r *http.Request) {
	p := Idea{
		ID:       1,
		Pitch:    "Airbnb for renting, reliable tenants paired with great landlords. App promises landlords and tenants what they want, and uses algorithms to determine if they're reputable or not. A renting marketplace end to end.",
		Created:  0,
		Approved: 0,
		Shown:    0,
	}

	err := indexTemplate.Execute(w, p)
	if err != nil {
		log.Fatal(err)
	}
}

func submit(w http.ResponseWriter, r *http.Request) {
	err := submitTemplate.Execute(w, nil)
	if err != nil {
		log.Fatal(err)
	}
}

func main() {
	db = grabDB()
	loadTemplates()

	http.HandleFunc("/", index)
	http.HandleFunc("/submit", submit)
	log.Fatal(http.ListenAndServe(":8080", nil))
}
