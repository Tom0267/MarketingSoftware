package main

import (
	"crypto/tls"
	"fmt"
	"net"
	"net/smtp"
	"os"
	"strings"
)

// MailingListInterface defines the behavior of a mailing list
type MailingListInterface interface {
	AddSubscriber(email string)
	RemoveSubscriber(email string)
	sendEmailTLS(from, password, smtpServer string, to []string, subject, body string) error
	GetAllSubscribers() []string
}

// MailingList is a concrete implementation of MailingListInterface
type MailingList struct {
	Emails map[string]bool // map to store unique email addresses
}

// NewMailingList initializes and returns a new MailingList
func NewMailingList() *MailingList {
	return &MailingList{
		Emails: make(map[string]bool),
	}
}

// AddSubscriber adds an email to the mailing list
func (ml *MailingList) AddSubscriber(email string) {
	if ml.Emails[email] {
		fmt.Println("Email already subscribed:", email)
		return
	}
	ml.Emails[email] = true
	fmt.Println("Successfully subscribed:", email)
}

// RemoveSubscriber removes an email from the mailing list
func (ml *MailingList) RemoveSubscriber(email string) {
	if !ml.Emails[email] {
		fmt.Println("Email not found:", email)
		return
	}
	delete(ml.Emails, email)
	fmt.Println("Successfully unsubscribed:", email)
}

// GetAllSubscribers returns a list of all subscribers
func (ml *MailingList) GetAllSubscribers() []string {
	subscribers := make([]string, 0, len(ml.Emails))
	for email := range ml.Emails {
		subscribers = append(subscribers, email)
	}
	return subscribers
}

func (ml *MailingList) sendEmailTLS(from, password, smtpServer string, to []string, subject, body string) error {
	host, _, _ := net.SplitHostPort(smtpServer)
	auth := smtp.PlainAuth("", from, password, host)

	// Establish a TLS connection
	conn, err := tls.Dial("tcp", smtpServer, &tls.Config{ServerName: host})
	if err != nil {
		return fmt.Errorf("failed to establish TLS connection: %v", err)
	}
	client, err := smtp.NewClient(conn, host)
	if err != nil {
		return fmt.Errorf("failed to create SMTP client: %v", err)
	}
	defer client.Quit()

	// Authenticate
	if err = client.Auth(auth); err != nil {
		return fmt.Errorf("authentication failed: %v", err)
	}

	// Set sender and recipients
	if err = client.Mail(from); err != nil {
		return fmt.Errorf("failed to set sender: %v", err)
	}
	for _, addr := range to {
		if err = client.Rcpt(addr); err != nil {
			return fmt.Errorf("failed to set recipient %s: %v", addr, err)
		}
	}

	// Write the email body
	w, err := client.Data()
	if err != nil {
		return fmt.Errorf("failed to get writer: %v", err)
	}
	msg := []byte(fmt.Sprintf("Subject: %s\r\n\r\n%s", subject, body))
	_, err = w.Write(msg)
	if err != nil {
		return fmt.Errorf("failed to write message: %v", err)
	}
	if err = w.Close(); err != nil {
		return fmt.Errorf("failed to close writer: %v", err)
	}

	fmt.Println("Email sent successfully!")
	return nil
}

func getEmail() string {
	data, err := os.ReadFile("C:/Email.txt") // Replace with your file name
	if err != nil {
		fmt.Printf("Failed to read file: %v\n", err)
		return ""
	}

	// Convert the byte slice to a string
	content := string(data)

	//split the content by new line
	lines := strings.Split(content, "\n")
	email := lines[0]
	email = strings.TrimSuffix(email, "\r")
	return email
}

func getPassword() string {
	data, err := os.ReadFile("C:/Email.txt") // Replace with your file name
	if err != nil {
		fmt.Printf("Failed to read file: %v\n", err)
		return ""
	}

	// Convert the byte slice to a string
	content := string(data)

	//split the content by new line
	lines := strings.Split(content, "\n")
	pass := lines[1]
	pass = strings.TrimSuffix(pass, "\r")
	return pass
}

// Main function demonstrating the use of the MailingListInterface
func main() {
	var mailingList MailingListInterface = NewMailingList()

	// Add subscribers
	mailingList.AddSubscriber("121year@gmail.com")
	mailingList.AddSubscriber("user2@example.com")
	mailingList.AddSubscriber("user2@example.com") // duplicate test

	// Remove a subscriber
	mailingList.RemoveSubscriber("user2@example.com")

	// List all subscribers
	fmt.Println("Current Subscribers:", mailingList.GetAllSubscribers())

	email := getEmail()
	password := getPassword()

	// Send email to all remaining subscribers
	subject := "Welcome to our mailing list!"
	message := "Thank you for subscribing. We are excited to have you with us!"
	mailingList.sendEmailTLS(email, password, "smtp.gmail.com:587", mailingList.GetAllSubscribers(), subject, message)
}
