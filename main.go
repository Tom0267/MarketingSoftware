package main

import (
	"fmt"
	"log"
	"net/http"

	_ "github.com/mattn/go-sqlite3"
)

func main() {
	//initialize the DB
	db, err := initDB("templates.db")
	if err != nil {
		log.Fatalf("failed to initialize database: %v", err)
	}
	defer db.Close()
	//clear the templates table for testing
	clearTemplates()

	//save test template
	saveTemplate("Welcome", "Welcome to our platform!")

	//serve the composer and templates route
	http.HandleFunc("/composer", composerHandler)
	http.HandleFunc("/templates", templatesHandler)

	//start the web server
	log.Fatal(http.ListenAndServe(":8080", nil))
	fmt.Println("Server started on localhost:8080")
}
