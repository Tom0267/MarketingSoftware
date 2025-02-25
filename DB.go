package main

import (
	"database/sql"
	"fmt"

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

	// create users table
	createUsersSQL := `CREATE TABLE IF NOT EXISTS users (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		email TEXT UNIQUE
	);`

	// create campaigns table
	createCampaignsSQL := `CREATE TABLE IF NOT EXISTS campaigns (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		name TEXT UNIQUE
	);`

	// create campaign_subscribers table to link subscribers to campaigns
	createCampaignSubscribersSQL := `CREATE TABLE IF NOT EXISTS campaign_subscribers (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		campaign_id INTEGER,
		subscriber_id INTEGER,
		joinDate DATE DEFAULT CURRENT_DATE,
		FOREIGN KEY (campaign_id) REFERENCES campaigns(id),
		FOREIGN KEY (subscriber_id) REFERENCES users(id)
	);`

	for _, sql := range []string{createTableSQL, createMailingListsSQL, createUsersSQL, createCampaignsSQL, createCampaignSubscribersSQL} {
		_, err := db.Exec(sql)
		if err != nil {
			return nil, fmt.Errorf("error creating table: %v", err)
		}
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

func createMailingList(name string) error {
	_, err := database.Exec(`INSERT INTO mailing_lists (name) VALUES (?)`, name)
	if err != nil {
		return fmt.Errorf("error creating mailing list: %v", err)
	}
	fmt.Println("Mailing list created")
	return nil
}

func addSubscriber(listName, email string) error {
	// check if the subscriber already exists
	var id int
	err := database.QueryRow(`SELECT id FROM users WHERE email = ?`, email).Scan(&id)
	if err != nil {
		return fmt.Errorf("error adding subscriber: %v", err)
	}
	//check campaign exists
	err = database.QueryRow(`SELECT id FROM campaigns WHERE name = ?`, listName).Scan(&id)
	if err != nil {
		return fmt.Errorf("error adding subscriber: %v", err)
	}
	// check if the subscriber is already in the mailing list
	err = database.QueryRow(`SELECT id FROM campaign_subscribers WHERE campaign_id = (SELECT id FROM campaigns WHERE name = ?) AND subscriber_id = (SELECT id FROM users WHERE email = ?)`, listName, email).Scan(&id)
	if err == nil {
		return fmt.Errorf("subscriber already exists in mailing list")
	}
	// add the subscriber to the mailing list
	_, err = database.Exec(`INSERT INTO campaign_subscribers (campaign_id, subscriber_id) SELECT c.id, u.id FROM campaigns c, users u WHERE c.name = ? AND u.email = ?`, listName, email)
	if err != nil {
		return fmt.Errorf("error adding subscriber to mailing list: %v", err)
	}

	return nil
}

func getSubscribers(campaignName string) ([]string, error) {
	rows, err := database.Query(`SELECT u.email FROM users u JOIN campaign_subscribers cs ON u.id = cs.subscriber_id JOIN campaigns c ON cs.campaign_id = c.id WHERE c.name = ?`, campaignName)
	if err != nil {
		return nil, fmt.Errorf("error querying subscribers: %v", err)
	}
	defer rows.Close()

	var emails []string
	for rows.Next() {
		var email string
		if err := rows.Scan(&email); err != nil {
			return nil, fmt.Errorf("error scanning row: %v", err)
		}
		emails = append(emails, email)
	}
	return emails, nil
}

func saveCampaign(name string, emails []string) error {
	_, err := database.Exec(`INSERT INTO campaigns (name) VALUES (?)`, name)
	if err != nil {
		return fmt.Errorf("error creating campaign: %v", err)
	}
	//save the subscribers if there are any

	fmt.Println("Campaign created")
	return nil
}

func updateTemplate(id int, title, content string) error {
	updateSQL := `UPDATE email_templates SET title = ?, content = ? WHERE id = ?`
	_, err := database.Exec(updateSQL, title, content, id)
	if err != nil {
		return fmt.Errorf("error updating template: %v", err)
	}
	return nil
}

func deleteTemplate(id int) error {
	deleteSQL := `DELETE FROM email_templates WHERE id = ?`
	_, err := database.Exec(deleteSQL, id)
	if err != nil {
		return fmt.Errorf("error deleting template: %v", err)
	}
	return nil
}

func deleteMailingList(name string) error {
	tx, err := database.Begin()
	if err != nil {
		return fmt.Errorf("error beginning transaction: %v", err)
	}

	// Delete subscribers associated with the mailing list
	deleteSubscribersSQL := `DELETE FROM campaign_subscribers WHERE campaign_id = (SELECT id FROM campaigns WHERE name = ?)`
	_, err = tx.Exec(deleteSubscribersSQL, name)
	if err != nil {
		tx.Rollback()
		return fmt.Errorf("error deleting subscribers: %v", err)
	}

	// Delete the mailing list
	deleteMailingListSQL := `DELETE FROM mailing_lists WHERE name = ?`
	_, err = tx.Exec(deleteMailingListSQL, name)
	if err != nil {
		tx.Rollback()
		return fmt.Errorf("error deleting mailing list: %v", err)
	}

	err = tx.Commit()
	if err != nil {
		return fmt.Errorf("error committing transaction: %v", err)
	}
	return nil
}

func removeSubscriber(listName, email string) error {
	tx, err := database.Begin()
	if err != nil {
		return fmt.Errorf("error starting transaction: %v", err)
	}

	deleteSQL := `DELETE FROM campaign_subscribers WHERE subscriber_id = (SELECT id FROM users WHERE email = ?) AND campaign_id = (SELECT id FROM campaigns WHERE name = ?)`

	_, err = tx.Exec(deleteSQL, email, listName)
	if err != nil {
		tx.Rollback()
		return fmt.Errorf("error removing subscriber: %v", err)
	}

	err = tx.Commit()
	if err != nil {
		return fmt.Errorf("error committing transaction: %v", err)
	}

	return nil
}

func addUser(email string) error {
	_, err := database.Exec(`INSERT INTO users (name) VALUES (?)`, email)
	if err != nil {
		return fmt.Errorf("error creating campaign: %v", err)
	}
	return nil
}

func getAllCampaigns() ([]string, error) {
	rows, err := database.Query(`SELECT name FROM campaigns`)
	if err != nil {
		return nil, fmt.Errorf("error querying campaigns: %v", err)
	}
	defer rows.Close()

	var campaigns []string
	for rows.Next() {
		var name string
		if err := rows.Scan(&name); err != nil {
			return nil, fmt.Errorf("error scanning row: %v", err)
		}
		campaigns = append(campaigns, name)
	}
	return campaigns, nil
}
