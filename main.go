package main

import (
	"fmt"
	"html/template"
	"io"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"
	"sync"
)

// campaign struct is retained for future use
type campaign struct {
	ID    int    `json:"id"`
	email string `json:"email"`
}

// global variables for synchronisation in file upload
var (
	posts   = make(map[int]campaign) // a map that holds posts in memory
	nextID  = 1                      // a variable to create unique post ids
	postsMu sync.Mutex               // mutex to synchronise access to posts map
)

func main() {
	// register the /composer route
	http.HandleFunc("/composer", composerHandler)

	// start the webserver
	fmt.Println("server is running at http://localhost:8080")
	log.Fatal(http.ListenAndServe(":8080", nil))
}

func composerHandler(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case "POST":
		handlePostComposer(w, r)
	case "GET":
		handleGetComposer(w, r)
	default:
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
	}
}

func handleGetComposer(w http.ResponseWriter, r *http.Request) {
	// set content type to text/html so that the browser renders html correctly
	w.Header().Set("Content-Type", "text/html")

	// parse the compose template
	tmpl, err := template.ParseFiles("templates/compose.tmpl") // load the template from the 'templates' folder
	if err != nil {
		http.Error(w, "error loading template", http.StatusInternalServerError)
		log.Println("error loading template:", err)
		return
	}

	// execute the template with no dynamic data
	err = tmpl.Execute(w, nil)
	if err != nil {
		http.Error(w, "error rendering template", http.StatusInternalServerError)
		log.Println("error rendering template:", err)
	}
}

func handlePostComposer(w http.ResponseWriter, r *http.Request) {
	err := r.ParseMultipartForm(10 * 1024 * 1024) // 10mb max form size
	if err != nil {
		http.Error(w, "form parsing error", http.StatusBadRequest)
		return
	}

	// get recipient, subject and body from the form
	recipientStr := r.FormValue("recipient")
	subject := r.FormValue("subject")
	body := r.FormValue("body")

	// split recipients by comma and trim spaces
	recipients := strings.Split(recipientStr, ",")
	for i := range recipients {
		recipients[i] = strings.TrimSpace(recipients[i])
	}

	// get filename from form data (if file is uploaded using chunking)
	fileName := r.FormValue("filename")
	// define a temporary file name based on the original file name
	outFileName := "temp_uploads/" + fileName
	// create the folder if it doesn't exist
	os.MkdirAll("temp_uploads", os.ModePerm)

	// get chunk index and total chunks from form data
	chunkIndex := r.FormValue("chunk_index")
	totalChunksStr := r.FormValue("total_chunks")

	// if no chunk data is provided then handle attachments normally
	if chunkIndex == "" || totalChunksStr == "" {
		var attachments []string
		attachmentFiles := r.MultipartForm.File["attachments"]
		for _, fileHeader := range attachmentFiles {
			attPath := "temp_uploads/" + fileHeader.Filename
			attachments = append(attachments, attPath)
			os.MkdirAll("temp_uploads", os.ModePerm)
			outFile, err := os.Create(attPath)
			if err != nil {
				http.Error(w, "error creating attachment file", http.StatusInternalServerError)
				return
			}
			defer outFile.Close()

			file, err := fileHeader.Open()
			if err != nil {
				http.Error(w, "error opening attachment file", http.StatusInternalServerError)
				return
			}
			defer file.Close()

			_, err = io.Copy(outFile, file)
			if err != nil {
				http.Error(w, "error saving attachment file", http.StatusInternalServerError)
				return
			}
		}

		// similarly, handle inline images
		images := make(map[string]string)
		imageFiles := r.MultipartForm.File["images"]
		for _, fileHeader := range imageFiles {
			cid := fileHeader.Filename
			imagePath := "temp_uploads/" + fileHeader.Filename
			images[cid] = imagePath

			os.MkdirAll("temp_uploads", os.ModePerm)
			outFile, err := os.Create(imagePath)
			if err != nil {
				http.Error(w, "error creating image file", http.StatusInternalServerError)
				return
			}
			defer outFile.Close()

			file, err := fileHeader.Open()
			if err != nil {
				http.Error(w, "error opening image file", http.StatusInternalServerError)
				return
			}
			defer file.Close()

			_, err = io.Copy(outFile, file)
			if err != nil {
				http.Error(w, "error saving image file", http.StatusInternalServerError)
				return
			}
		}

		err = sendMail(recipients, subject, body, attachments, images)
		if err != nil {
			http.Error(w, "error sending email", http.StatusInternalServerError)
			return
		}
		w.Write([]byte("email sent successfully"))
		return
	}

	// process chunked upload
	totalChunks, err := strconv.Atoi(totalChunksStr)
	if err != nil {
		http.Error(w, "error converting total_chunks to int", http.StatusBadRequest)
		return
	}

	// if this is the first chunk, create a new file to store the upload
	if chunkIndex == "0" {
		outFile, err := os.Create(outFileName)
		if err != nil {
			http.Error(w, "error creating file", http.StatusInternalServerError)
			return
		}
		outFile.Close()
	}

	// read the current chunk and append to the file
	fileChunk, _, err := r.FormFile("attachment_chunk")
	if err != nil {
		http.Error(w, "error reading chunk", http.StatusBadRequest)
		return
	}
	defer fileChunk.Close()

	outFile, err := os.OpenFile(outFileName, os.O_APPEND|os.O_WRONLY, 0644)
	if err != nil {
		http.Error(w, "error opening file", http.StatusInternalServerError)
		return
	}
	defer outFile.Close()

	_, err = io.Copy(outFile, fileChunk)
	if err != nil {
		http.Error(w, "error saving chunk", http.StatusInternalServerError)
		return
	}

	// if this is the last chunk, send the email with the file attached
	if chunkIndex == strconv.Itoa(totalChunks-1) {
		attachments := []string{outFileName}
		images := make(map[string]string) // assume no inline images for chunked upload

		err = sendMail(recipients, subject, body, attachments, images)
		if err != nil {
			http.Error(w, "error sending email", http.StatusInternalServerError)
			return
		}

		err = os.Remove(outFileName)
		if err != nil {
			http.Error(w, "error deleting temporary file", http.StatusInternalServerError)
			return
		}
	}

	w.Write([]byte("chunk received"))
}
