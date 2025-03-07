package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"
	"text/template"
	"time"
)

// get db connection from the main function
var db *sql.DB

type errorMessage struct {
	Message string `json:"message"`
}

// templatesHandler returns all email templates as json
func templatesHandler(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case "POST":
		log.Println("Received request to save template in templatesHandler")
		saveTemplateHandler(w, r)
	case "GET":
		templateGetter(w, r)
	default:
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
	}
}

// templateHandler returns all email templates as json
func templateGetter(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	// get templates from the database
	templates, err := getTemplates()
	if err != nil {
		http.Error(w, fmt.Sprintf(`{"error": "failed to fetch templates: %v"}`, err), http.StatusInternalServerError)
		return
	}

	// wrap response in an object with a "templates" key
	response := map[string]interface{}{
		"templates": templates,
	}

	log.Println("Sending templates:", response)

	// encode and send response
	json.NewEncoder(w).Encode(response)
}

// saveTemplateHandler allows saving a new email template
func saveTemplateHandler(w http.ResponseWriter, r *http.Request) {

	var template struct {
		Title   string `json:"Title"`
		Content string `json:"Content"`
	}
	log.Println("Received request to save template")
	err := json.NewDecoder(r.Body).Decode(&template)
	if err != nil {
		json.NewEncoder(w).Encode(errorMessage{Message: "failed to decode request body"})
		return
	}

	// save the template to the database
	err = saveTemplate(template.Title, template.Content)
	if err != nil {
		json.NewEncoder(w).Encode(errorMessage{Message: "failed to save template"})
		return
	}
	log.Println("Template saved successfully")
	json.NewEncoder(w).Encode(map[string]string{"success": "true", "message": "Template saved successfully!"})
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
		json.NewEncoder(w).Encode(map[string]string{"message": "error parsing form data (max 10mb)"})
		return
	}

	// get recipient, subject and body from the form
	recipientStr := r.FormValue("recipients")
	campaigns := r.FormValue("campaigns")
	subject := r.FormValue("subject")
	body := r.FormValue("body")

	// split recipients by comma and trim spaces
	recipients := []string{}
	for _, r := range strings.Split(recipientStr, ",") {
		r = strings.TrimSpace(r)
		if r != "" { // Only add non-empty values
			recipients = append(recipients, r)
		}
	}

	// split campaigns by comma and trim spaces
	type Campaign struct {
		Name string `json:"name"`
	}
	rawCampaigns := strings.Split(campaigns, ",")
	campaignList := make([]Campaign, 0, len(rawCampaigns))

	for _, rawCampaign := range rawCampaigns {
		if cleanedCampaign := strings.TrimSpace(rawCampaign); cleanedCampaign != "" {
			campaignList = append(campaignList, Campaign{Name: cleanedCampaign})
		}
	}

	// get the campaign subscribers
	for _, campaign := range campaignList {
		subscribers, err := getSubscribers(campaign.Name)
		log.Println("Subscribers for campaign", campaign.Name, ":", subscribers)
		if err != nil {
			json.NewEncoder(w).Encode(map[string]string{"message": "error fetching subscribers for campaign: " + campaign.Name})
			return
		}
		recipients = append(recipients, subscribers...)
	}
	log.Println("Recipients:", recipients)
	for i := range recipients {
		recipients[i] = strings.TrimSpace(recipients[i])
	}
	log.Println("Recipients after trimming:", recipients)

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
		attachmentFiles := r.MultipartForm.File["attachments[]"]
		// check for duplicate attachments
		attachmentMap := make(map[string]bool)
		for _, fileHeader := range attachmentFiles {
			//remove duplicate attachments
			if !attachmentMap[fileHeader.Filename] {
				attachmentMap[fileHeader.Filename] = true
			}
		}

		for fileName := range attachmentMap {
			attPath := "temp_uploads/" + fileName
			attachments = append(attachments, attPath)
			os.MkdirAll("temp_uploads", os.ModePerm)
			outFile, err := os.Create(attPath)
			if err != nil {
				json.NewEncoder(w).Encode(map[string]string{"message": "error creating attachment file"})
				return
			}
			defer outFile.Close()

			for _, fileHeader := range r.MultipartForm.File["attachments[]"] {
				if fileHeader.Filename == fileName {
					file, err := fileHeader.Open()
					if err != nil {
						json.NewEncoder(w).Encode(map[string]string{"message": "error opening attachment file"})
						return
					}
					defer file.Close()

					_, err = io.Copy(outFile, file)
					if err != nil {
						json.NewEncoder(w).Encode(map[string]string{"message": "error saving attachment file"})
						return
					}
				}
			}
		}

		// handle inline images
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

		err := sendMail(recipients, subject, body, attachments, images)
		if err != nil {
			json.NewEncoder(w).Encode(map[string]string{"success": "false", "message": "error sending email"})
			return
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]string{"success": "true", "message": "email sent successfully"})

		clearTempFiles() // clear temp files after 30 seconds
		return
	}

	// process chunked upload
	totalChunks, err := strconv.Atoi(totalChunksStr)
	if err != nil {
		json.NewEncoder(w).Encode(map[string]string{"success": "false", "message": "error converting total_chunks to int"})
		return
	}

	// if this is the first chunk, create a new file to store the upload
	if chunkIndex == "0" {
		outFile, err := os.Create(outFileName)
		if err != nil {
			json.NewEncoder(w).Encode(map[string]string{"success": "false", "message": "error creating file"})
			return
		}
		outFile.Close()
	}

	// read the current chunk and append to the file
	fileChunk, _, err := r.FormFile("attachment_chunk")
	if err != nil {
		json.NewEncoder(w).Encode(map[string]string{"success": "false", "message": "error reading chunk"})
		return
	}
	defer fileChunk.Close()

	outFile, err := os.OpenFile(outFileName, os.O_APPEND|os.O_WRONLY, 0644)
	if err != nil {
		json.NewEncoder(w).Encode(map[string]string{"success": "false", "message": "error opening file"})
		return
	}
	defer outFile.Close()

	_, err = io.Copy(outFile, fileChunk)
	if err != nil {
		json.NewEncoder(w).Encode(map[string]string{"success": "false", "message": "error saving chunk"})
		return
	}

	// if this is the last chunk, send the email with the file attached
	if chunkIndex == strconv.Itoa(totalChunks-1) {
		attachments := []string{outFileName}
		images := make(map[string]string) // assume no inline images for chunked upload

		err = sendMail(recipients, subject, body, attachments, images)
		if err != nil {
			json.NewEncoder(w).Encode(map[string]string{"success": "false", "message": "error sending email"})
			return
		}

		err = os.Remove(outFileName)
		if err != nil {
			json.NewEncoder(w).Encode(map[string]string{"success": "false", "message": "error deleting files"})
			return
		}
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{"success": "true", "message": "chunk uploaded successfully"})
}

func clearTempFiles() {
	//create a thread to clean up the temp folder after 30 seconds
	go func() {
		// sleep for 30 seconds
		<-time.After(30 * time.Second)
		err := os.RemoveAll("temp_uploads")
		if err != nil {
			log.Println("error cleaning up temp folder:", err)
		}
	}()
}

func campaignHandler(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case "POST":
		handlePostCampaign(w, r)
	case "GET":
		handleGetCampaign(w, r)
	default:
		json.NewEncoder(w).Encode(map[string]string{"success": "false", "message": "method not allowed"})
	}
}

func handleGetCampaign(w http.ResponseWriter, r *http.Request) {
	//get the name of the requested campaign
	campaignName := r.URL.Query().Get("name")
	if campaignName == "" {
		json.NewEncoder(w).Encode(map[string]string{"success": "false", "message": "missing campaign name"})
		return
	}

	//get the campaign from the database
	campaign, err := getSubscribers(campaignName)
	if err != nil {
		json.NewEncoder(w).Encode(map[string]string{"success": "false", "message": "error fetching campaign"})
		return
	}
	json.NewEncoder(w).Encode(campaign)
}

func handlePostCampaign(w http.ResponseWriter, r *http.Request) {
	//parse the request body
	var campaign struct {
		Name       string   `json:"campaignName"`
		Recipients []string `json:"mailingList"`
	}
	err := json.NewDecoder(r.Body).Decode(&campaign)
	if err != nil {
		fmt.Printf("error decoding campaign\n")
		json.NewEncoder(w).Encode(map[string]string{"success": "false", "message": "Campaign failed to saved"})
		return
	}

	//save the campaign to the database
	err = saveCampaign(campaign.Name, campaign.Recipients)
	if err != nil {
		json.NewEncoder(w).Encode(map[string]string{"success": "false", "message": "Campaign failed to saved"})
		return
	}

	json.NewEncoder(w).Encode(map[string]string{"success": "true", "message": "Campaign saved successfully!"})
}

func listHandler(w http.ResponseWriter, r *http.Request) {
	// query all campaigns from the database
	campaigns, err := getAllCampaigns()
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]string{"success": "false", "message": "error fetching campaigns"})
		return
	}
	// return campaigns
	json.NewEncoder(w).Encode(map[string]interface{}{"campaigns": campaigns})
}
