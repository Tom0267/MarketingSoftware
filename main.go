package main

import (
	"fmt"
	"log"
	"mime"
	"net/http"

	_ "github.com/mattn/go-sqlite3"
)

func main() {
	err := mime.AddExtensionType(".js", "application/javascript")
	if err != nil {
		log.Fatalf("Error adding MIME type: %v", err)
	}

	//initialize the DB
	db, err := initDB("templates.db")
	if err != nil {
		log.Fatalf("failed to initialize database: %v", err)
	}
	defer db.Close()

	//clearCampaigns() //clear campaigns for testing
	//clearDatabase() //clear database for testing

	//save test template
	//saveTemplate("Welcome", "Welcome to our platform!")
	//createMailingList("test")
	//create a test user
	//addUser("121year@gmail.com")
	//addSubscriber("test", "121year@gmail.com")

	//serve the routes
	http.Handle("/JavaScript/", http.StripPrefix("/JavaScript/", http.FileServer(http.Dir("./JavaScript"))))
	http.HandleFunc("/JavaScript/script.js", scriptHandler)
	http.HandleFunc("/composer", composerHandler)
	http.HandleFunc("/templates", templatesHandler)
	http.HandleFunc("/campaigns/list", listHandler)
	http.HandleFunc("/campaigns", campaignHandler)

	//start the web server
	fmt.Println("Server started on localhost:8080")
	log.Fatal(http.ListenAndServe(":8080", nil))
}
