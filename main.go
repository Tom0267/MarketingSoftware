package main

import (
	"fmt"
	"html/template"
	"log"
	"net/http"
	"strings"
	"sync"
)

type campaign struct {
	ID    int    `json:"id"`
	email string `json:"email"`
}

var (
	posts   = make(map[int]campaign) // a map that will hold our posts in memory
	nextID  = 1                      // a variable to help create unique post IDs
	postsMu sync.Mutex               // a mutex to synchronize access to the posts map
)

func main() {
	//register the /Composer route
	http.HandleFunc("/composer", composerHandler)

	//start the webserver
	fmt.Println("Server is running at http://localhost:8080")
	log.Fatal(http.ListenAndServe(":8080", nil))
}

func composerHandler(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case "POST":
		handlePostComposer(w, r)
	case "GET":
		handleGetComposer(w, r)
	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

func handleGetComposer(w http.ResponseWriter, r *http.Request) {
	//set content type to text/html to ensure the browser renders HTML properly
	w.Header().Set("Content-Type", "text/html")

	//parse the compose template
	tmpl, err := template.ParseFiles("templates/compose.tmpl") //load the template from the 'templates' folder
	if err != nil {
		http.Error(w, "Error loading template", http.StatusInternalServerError)
		return
	}

	//execute the template with the provided data, writing the result to the ResponseWriter
	err = tmpl.Execute(w, nil)
	if err != nil {
		http.Error(w, "Error rendering template", http.StatusInternalServerError)
	}
}

func handlePostComposer(w http.ResponseWriter, r *http.Request) {
	//parse the form data
	err := r.ParseForm()
	if err != nil {
		http.Error(w, "Error parsing form", http.StatusInternalServerError)
		return
	}
	//retrieve the form data from the request and send the email
	recipient := r.FormValue("recipient")
	recipients := strings.Split(recipient, ",") //split the recipient string by commas
	fmt.Println(recipients)

	sendMail(recipients, r.FormValue("subject"), r.FormValue("body"))

	//respond back to the user
	w.Write([]byte("Email sent successfully!"))
}
