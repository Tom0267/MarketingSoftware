package main

import (
	"crypto/tls"
	"fmt"
	"os"
	"strings"

	"github.com/joho/godotenv"
	"gopkg.in/mail.v2"
)

func getRecipients() []string {
	//get email credentials from a file elsewhere
	data, err := os.ReadFile("C:/Recipients.txt") //replace with your file name
	if err != nil {
		return nil
	}

	//convert the byte slice to a string
	content := string(data)

	//split the content by new line
	lines := strings.Split(content, "\n")
	//clean the lines by removing any whitespace
	for i := range lines {
		lines[i] = strings.TrimSuffix(lines[i], "\r")
	}

	//create an array to store the recipients
	var recipients []string
	for i := range lines {
		recipients = append(recipients, lines[i])
	}
	return recipients
}

func sendMail(recipients []string, subject string, body string) {
	// Load .env file (if present)
	err := godotenv.Load()
	if err != nil {
		fmt.Println("Error loading .env file")
		return
	}

	smtpHost := os.Getenv("SMTP_HOST")
	smtpPort := 465

	smtpUser := os.Getenv("SMTP_USER")
	smtpPassword := os.Getenv("SMTP_PASSWORD")

	if smtpHost == "" || smtpUser == "" || smtpPassword == "" {
		fmt.Println("Failed to get email credentials")
		return
	}

	dialer := mail.NewDialer(smtpHost, smtpPort, smtpUser, smtpPassword) //creates a new dialer to connect to the smtp server

	m := mail.NewMessage()
	m.SetHeader("From", "SMTP_USER") //set the sender
	m.SetHeader("To", recipients...) //set the recipients
	m.SetHeader("Subject", subject)  //set the subject
	m.SetBody("text/plain", body)    //set the body

	dialer.TLSConfig = &tls.Config{InsecureSkipVerify: true} //change to false in production to verify the server's certificate
	if err := dialer.DialAndSend(m); err != nil {
		panic(err)
	}
	fmt.Println("Email sent successfully!")
}
