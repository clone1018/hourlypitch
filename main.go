package main

import (
	"database/sql"
	"fmt"
	"html/template"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"strings"
	"time"

	_ "github.com/mattn/go-sqlite3"
)

var db *sql.DB
var indexTemplate *template.Template
var submitTemplate *template.Template
var adminTemplate *template.Template
var currentIdea Idea

type Idea struct {
	ID       int
	Pitch    string
	Created  int
	Approved sql.NullInt64
	Shown    sql.NullInt64
}

func grabDB() *sql.DB {
	db, err := sql.Open("sqlite3", "./db/hourlypitch.db")
	if err != nil {
		log.Fatal(err)
	}
	return db
}
func loadSchema(schemaFile string) {
	file, err := ioutil.ReadFile(schemaFile)
	if err != nil {
		log.Fatal(err)
	}

	requests := strings.Split(string(file), ";\n")

	for _, request := range requests {
		_, err := db.Exec(request)
		if err != nil {
			log.Println(request)
			log.Fatal(err)
		}
	}
}

func loadTemplates() {
	var err error
	indexTemplate, err = template.ParseFiles("templates/index.html")
	if err != nil {
		log.Fatal(err)
	}
	submitTemplate, err = template.ParseFiles("templates/submit.html")
	if err != nil {
		log.Fatal(err)
	}
	adminTemplate, err = template.ParseFiles("templates/admin.html")
	if err != nil {
		log.Fatal(err)
	}
}

func index(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html")

	err := indexTemplate.Execute(w, currentIdea)
	if err != nil {
		log.Print(err)
	}
}

func submit(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html")

	type Data struct {
		Message string ""
	}
	resp := Data{}

	if r.URL.Query().Get("msg") == "err" {
		resp.Message = "There was a problem saving your idea. 500 chars maximum."
	} else if r.URL.Query().Get("msg") == "good" {
		resp.Message = "Your idea has been submitted, it will show up eventually."
	}

	err := submitTemplate.Execute(w, resp)
	if err != nil {
		log.Print(err)
	}
}

func submitSave(w http.ResponseWriter, r *http.Request) {
	pitch := r.FormValue("pitch")
	if pitch == "" || len(pitch) > 500 {
		http.Redirect(w, r, "/submit?msg=err", 301)
		return
	}

	insert, err := db.Prepare("INSERT INTO ideas(pitch, created) VALUES(?,?)")
	if err != nil {
		log.Print(err.Error())
	}
	res, err := insert.Exec(pitch, time.Now().Unix())
	if err != nil {
		log.Print(err.Error())
	}
	_, err = res.LastInsertId()
	if err != nil {
		log.Print(err.Error())
	}

	http.Redirect(w, r, "/submit?msg=good", 301)
}

func rotate(w http.ResponseWriter, r *http.Request) {
	rows, err := db.Query("SELECT * FROM ideas where approved is not null and shown is null ORDER BY RANDOM() LIMIT 1;")
	if rows.Next() == false {
		// No idea found, let's keep the old one
		fmt.Fprint(w, "Keeping old")
		return
	}

	var i Idea
	err = rows.Scan(&i.ID, &i.Pitch, &i.Created, &i.Approved, &i.Shown)
	if err != nil {
		log.Print(err)
		fmt.Fprint(w, "Not Ok")
		return
	}
	_ = rows.Close()

	deactivate, err := db.Prepare("UPDATE ideas SET shown=? WHERE id=?")
	if err != nil {
		log.Print(err)
		fmt.Fprint(w, "Not Ok")
		return
	}
	_, err = deactivate.Exec(time.Now().Unix(), i.ID)
	if err != nil {
		log.Print(err)
		fmt.Fprint(w, "Not Ok")
		return
	}

	currentIdea = i
}

func admin(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html")

	rows, err := db.Query("SELECT id,pitch FROM ideas WHERE approved IS NULL")
	if err != nil {
		log.Print(err)
	}

	var ideas []Idea
	for rows.Next() {
		var idea Idea
		err = rows.Scan(&idea.ID, &idea.Pitch)
		if err != nil {
			log.Print(err)
		}

		ideas = append(ideas, idea)
	}

	err = adminTemplate.Execute(w, ideas)
	if err != nil {
		log.Print(err)
	}
}

func approve(w http.ResponseWriter, r *http.Request) {
	insForm, err := db.Prepare("UPDATE ideas SET approved=? WHERE id=?")
	if err != nil {
		log.Print(err)
	}
	_, err = insForm.Exec(time.Now().Unix(), r.FormValue("id"))
	if err != nil {
		log.Print(err)
	}

	http.Redirect(w, r, "/admin", 301)
}

func getRecentIdea() Idea {
	rows, err := db.Query("SELECT id,pitch,created,approved,shown FROM ideas where approved is not null ORDER BY shown LIMIT 1;")
	if rows.Next() == false {
		log.Print(err)
	}

	var i Idea
	err = rows.Scan(&i.ID, &i.Pitch, &i.Created, &i.Approved, &i.Shown)
	if err != nil {
		log.Print(err)
	}
	_ = rows.Close()

	return i
}

func auth(fn http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("WWW-Authenticate", "Basic realm=\"Collective\"")

		user, pass, _ := r.BasicAuth()
		if user != "admin" || pass != os.Getenv("PASS") {
			http.Error(w, "Unauthorized.", 401)
			return
		}
		fn(w, r)
	}
}

func main() {
	db = grabDB()
	loadSchema("db/schema.sql")
	loadTemplates()

	currentIdea = getRecentIdea()

	http.HandleFunc("/", index)
	http.HandleFunc("/submit", submit)
	http.HandleFunc("/submit-save", submitSave)

	http.HandleFunc("/admin", auth(admin))
	http.HandleFunc("/admin/approve", auth(approve))
	http.HandleFunc("/admin/rotate", auth(rotate))
	log.Print(http.ListenAndServe(":8080", nil))
}
