package main

import (
	"database/sql"
	"fmt"
	"log"

	_ "github.com/mattn/go-sqlite3"
)

var database *sql.DB

// emailTemplate represents a template email
type EmailTemplate struct {
	ID      int    `json:"Id"`
	Title   string `json:"Title"`
	Content string `json:"Content"`
}

// initDB initialises the sqlite database and creates the email_templates table if needed
func initDB(dbPath string) (*sql.DB, error) {
	db, err := sql.Open("sqlite3", dbPath)
	database = db
	if err != nil {
		return nil, fmt.Errorf("error opening database: %v", err)
	}

	createTableSQL := `CREATE TABLE IF NOT EXISTS email_templates (
		"id" INTEGER NOT NULL PRIMARY KEY AUTOINCREMENT,		
		"title" TEXT,
		"content" TEXT
	);`

	// create mailing lists table
	createMailingListsSQL := `CREATE TABLE IF NOT EXISTS mailing_lists (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		name TEXT UNIQUE
	);`

	// create subscribers table
	createSubscribersSQL := `CREATE TABLE IF NOT EXISTS subscribers (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		mailing_list_id INTEGER,
		email TEXT,
		FOREIGN KEY (mailing_list_id) REFERENCES mailing_lists(id)
	);`

	_, err = database.Exec(createTableSQL)
	if err != nil {
		return nil, fmt.Errorf("error creating table: %v", err)
	}
	_, err = database.Exec(createMailingListsSQL)
	if err != nil {
		return nil, fmt.Errorf("error creating mailing_lists table: %v", err)
	}
	_, err = database.Exec(createSubscribersSQL)
	if err != nil {
		return nil, fmt.Errorf("error creating subscribers table: %v", err)
	}
	if database == nil {
		log.Fatal("database connection is not initialized")
	}
	return database, nil
}

// saveTemplate inserts a new email template into the database
func saveTemplate(title, content string) error {
	insertSQL := `INSERT INTO email_templates(title, content) VALUES (?, ?)`
	_, err := database.Exec(insertSQL, title, content)
	return err
}

// getTemplates retrieves all email templates from the database
func getTemplates() ([]EmailTemplate, error) {
	rows, err := database.Query(`SELECT id, title, content FROM email_templates`)
	if err != nil {
		return nil, fmt.Errorf("error querying templates: %v", err)
	}
	defer rows.Close()

	var templates []EmailTemplate
	for rows.Next() {
		var t EmailTemplate
		if err := rows.Scan(&t.ID, &t.Title, &t.Content); err != nil {
			return nil, fmt.Errorf("error scanning row: %v", err)
		}
		templates = append(templates, t)
	}
	return templates, nil
}

func addMailingList(name string) error {
	_, err := database.Exec(`INSERT INTO mailing_lists (name) VALUES (?)`, name)
	if err != nil {
		return fmt.Errorf("error adding mailing list: %v", err)
	}
	return nil
}

func addSubscriber(listName, email string) error {
	var listID int
	err := database.QueryRow(`SELECT id FROM mailing_lists WHERE name = ?`, listName).Scan(&listID)
	if err != nil {
		return fmt.Errorf("mailing list not found: %v", err)
	}

	_, err = database.Exec(`INSERT INTO subscribers (mailing_list_id, email) VALUES (?, ?)`, listID, email)
	if err != nil {
		return fmt.Errorf("error adding subscriber: %v", err)
	}
	return nil
}

func getSubscribers(listName string) ([]string, error) {
	var listID int
	err := database.QueryRow(`SELECT id FROM mailing_lists WHERE name = ?`, listName).Scan(&listID)
	if err != nil {
		return nil, fmt.Errorf("mailing list not found: %v", err)
	}

	rows, err := database.Query(`SELECT email FROM subscribers WHERE mailing_list_id = ?`, listID)
	if err != nil {
		return nil, fmt.Errorf("error retrieving subscribers: %v", err)
	}
	defer rows.Close()

	var subscribers []string
	for rows.Next() {
		var email string
		if err := rows.Scan(&email); err != nil {
			return nil, fmt.Errorf("error scanning subscriber: %v", err)
		}
		subscribers = append(subscribers, email)
	}
	return subscribers, nil
}

func clearTemplates() error {
	_, err := database.Exec(`DELETE FROM email_templates`)
	if err != nil {
		return fmt.Errorf("error clearing database: %v", err)
	}
	return nil
}
