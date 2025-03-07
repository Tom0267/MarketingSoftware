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

	for _, sql := range []string{createTableSQL, createUsersSQL, createCampaignsSQL, createCampaignSubscribersSQL} {
		_, err := db.Exec(sql)
		if err != nil {
			return nil, fmt.Errorf("error creating table: %v", err)
		}
	}

	return database, nil
}

// saveTemplate inserts a new email template into the database
func saveTemplate(title, content string) error {
	var exists int
	err := database.QueryRow(`SELECT COUNT(*) FROM email_templates WHERE title = ?`, title).Scan(&exists)
	if err != nil {
		return fmt.Errorf("error checking template existence: %v", err)
	}
	if exists > 0 {
		return fmt.Errorf("template title already exists")
	}

	// insert new template
	_, err = database.Exec(`INSERT INTO email_templates (title, content) VALUES (?, ?)`, title, content)
	if err != nil {
		return fmt.Errorf("error saving template: %v", err)
	}
	log.Println("Template saved successfully:", title)
	return nil
}

func checkTemplatesExists() bool {
	var id int
	err := database.QueryRow(`SELECT id FROM email_templates`).Scan(&id)
	if err == sql.ErrNoRows {
		log.Println("No templates found in the database")
		return false
	} else if err != nil {
		log.Println("Error checking templates:", err)
		return false
	}
	log.Println("Template found with ID:", id)
	return true
}

// getTemplates retrieves all email templates from the database
func getTemplates() ([]EmailTemplate, error) {
	checkTemplatesExists()
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
	//verify if there are any templates
	if len(templates) == 0 {
		return nil, fmt.Errorf("no templates found")
	}
	return templates, nil
}

func createMailingList(name string) error {
	var id int
	err := database.QueryRow(`SELECT id FROM campaigns WHERE name = ?`, name).Scan(&id)

	// check if the campaign already exists
	if err == nil {
		return fmt.Errorf("mailing list name already exists")
	} else if err != sql.ErrNoRows {
		return fmt.Errorf("error checking campaign existence: %v", err)
	}

	// insert new campaign
	_, err = database.Exec(`INSERT INTO campaigns (name) VALUES (?)`, name)
	if err != nil {
		return fmt.Errorf("error creating campaign: %v", err)
	}

	fmt.Println("Campaign created successfully:", name)
	return nil
}

func addSubscriber(listName, email string) error {
	// check if the subscriber already exists
	var id int
	err := database.QueryRow(`SELECT id FROM users WHERE email = ?`, email).Scan(&id)
	if err != nil {
		return fmt.Errorf("error adding subscriber: %v", err)
	}
	// check campaign exists
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
	// prepare the query
	query := `
		SELECT u.email 
		FROM users u 
		JOIN campaign_subscribers cs ON u.id = cs.subscriber_id 
		JOIN campaigns c ON cs.campaign_id = c.id 
		WHERE c.name = ?`

	rows, err := database.Query(query, campaignName)
	if err != nil {
		return nil, fmt.Errorf("database query error: %w", err) // wrap error with context
	}
	defer rows.Close()

	var emails []string
	// use an explicit loop to scan rows efficiently
	for rows.Next() {
		var email string
		if err := rows.Scan(&email); err != nil {
			return nil, fmt.Errorf("database scan error: %w", err)
		}
		emails = append(emails, email)
	}

	// ensure we check for any iteration errors
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("database row iteration error: %w", err)
	}

	return emails, nil
}

func saveCampaign(name string, emails []string) error {
	// start a transaction for atomic operations
	tx, err := database.Begin()
	if err != nil {
		return fmt.Errorf("error starting transaction: %v", err)
	}

	// check if the campaign already exists
	var exists int
	err = tx.QueryRow(`SELECT COUNT(*) FROM campaigns WHERE name = ?`, name).Scan(&exists)
	if err != nil {
		tx.Rollback()
		return fmt.Errorf("error checking campaign existence: %v", err)
	}
	if exists > 0 {
		tx.Rollback()
		return fmt.Errorf("campaign name '%s' already exists", name)
	}

	// insert the new campaign
	result, err := tx.Exec(`INSERT INTO campaigns (name) VALUES (?)`, name)
	if err != nil {
		tx.Rollback()
		return fmt.Errorf("error creating campaign: %v", err)
	}

	// get new campaign ID
	campaignID64, err := result.LastInsertId()
	if err != nil {
		tx.Rollback()
		return fmt.Errorf("error getting new campaign ID: %v", err)
	}
	campaignID := int(campaignID64)

	log.Printf("Campaign '%s' created with ID: %d", name, campaignID)

	// add subscribers to the campaign
	for _, email := range emails {
		var userID int
		err := tx.QueryRow(`SELECT id FROM users WHERE email = ?`, email).Scan(&userID)

		// if user does not exist, create the user
		if err == sql.ErrNoRows {
			result, err := tx.Exec(`INSERT INTO users (email) VALUES (?)`, email)
			if err != nil {
				tx.Rollback()
				return fmt.Errorf("error adding user '%s': %v", email, err)
			}

			// get new user ID
			userID64, err := result.LastInsertId()
			if err != nil {
				tx.Rollback()
				return fmt.Errorf("error getting new user ID: %v", err)
			}
			userID = int(userID64)

			log.Printf("New user '%s' added with ID: %d", email, userID)
		} else if err != nil {
			tx.Rollback()
			return fmt.Errorf("error checking user existence: %v", err)
		}

		// add the user as a subscriber
		_, err = tx.Exec(`INSERT INTO campaign_subscribers (campaign_id, subscriber_id) VALUES (?, ?)`, campaignID, userID)
		if err != nil {
			tx.Rollback()
			return fmt.Errorf("error adding subscriber '%s' to campaign '%s': %v", email, name, err)
		}

		log.Printf("User '%s' subscribed to campaign '%s'", email, name)
	}

	// commit transaction if everything was successful
	err = tx.Commit()
	if err != nil {
		return fmt.Errorf("error committing transaction: %v", err)
	}

	log.Printf("Campaign '%s' saved successfully with %d subscribers.", name, len(emails))
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
		return fmt.Errorf("error starting transaction: %v", err)
	}

	// delete subscribers first
	if _, err := tx.Exec(`DELETE FROM campaign_subscribers WHERE campaign_id = (SELECT id FROM campaigns WHERE name = ?)`, name); err != nil {
		tx.Rollback()
		return fmt.Errorf("error deleting subscribers: %v", err)
	}

	// delete the campaign itself
	if _, err := tx.Exec(`DELETE FROM campaigns WHERE name = ?`, name); err != nil {
		tx.Rollback()
		return fmt.Errorf("error deleting mailing list: %v", err)
	}

	// commit transaction if all operations succeed
	if err := tx.Commit(); err != nil {
		return fmt.Errorf("error committing transaction: %v", err)
	}
	return nil
}

func removeSubscriber(listName, email string) error {
	tx, err := database.Begin()
	if err != nil {
		return fmt.Errorf("error starting transaction: %v", err)
	}

	deleteSQL := `
		DELETE FROM campaign_subscribers 
		WHERE subscriber_id = (SELECT id FROM users WHERE email = ?) 
		AND campaign_id = (SELECT id FROM campaigns WHERE name = ?)`

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
	_, err := database.Exec(`INSERT INTO users (email) VALUES (?)`, email)
	if err != nil {
		return fmt.Errorf("error adding user: %v", err)
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

	// ensure an empty array is returned instead of `nil`
	if campaigns == nil {
		campaigns = []string{}
	}

	return campaigns, nil
}

func clearCampaigns() {
	_, err := database.Exec(`DELETE FROM campaigns`)
	if err != nil {
		fmt.Errorf("error querying campaigns: %v", err)
	}
	_, err = database.Exec(`DELETE FROM campaign_subscribers`)
	if err != nil {
		fmt.Errorf("error querying campaign_subscribers: %v", err)
	}
}

func clearDatabase() {
	tables := []string{"email_templates", "users", "campaigns", "campaign_subscribers"}
	for _, table := range tables {
		_, err := database.Exec(fmt.Sprintf("DELETE FROM %s", table))
		if err != nil {
			log.Printf("Error clearing table %s: %v\n", table, err)
		}
	}
}
