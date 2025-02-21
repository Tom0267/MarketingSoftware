package main

import (
	"fmt"
	"os"
	"strings"

	"github.com/joho/godotenv"
	"gopkg.in/mail.v2"
)

func getRecipients() []string {
	// get recipient emails from a file
	data, err := os.ReadFile("C:/Recipients.txt") // replace with your file name
	if err != nil {
		fmt.Printf("error reading recipients file: %v\n", err)
		return nil
	}
	// convert the byte slice to a string
	content := string(data)
	// split the content by new line and clean whitespace
	lines := strings.Split(content, "\n")
	var recipients []string
	for _, line := range lines {
		cleaned := strings.TrimSpace(line)
		if cleaned != "" {
			recipients = append(recipients, cleaned)
		}
	}
	return recipients
}

func sendMail(recipients []string, subject string, body string, attachments []string, images map[string]string) error {
	// load environment variables from .env file
	err := godotenv.Load()
	if err != nil {
		return fmt.Errorf("error loading .env file: %v", err)
	}

	// get smtp credentials from environment variables
	smtpHost := os.Getenv("SMTP_HOST")
	smtpPort := 465
	smtpUser := os.Getenv("SMTP_USER")
	smtpPassword := os.Getenv("SMTP_PASSWORD")

	// check if email credentials are available
	if smtpHost == "" || smtpUser == "" || smtpPassword == "" {
		return fmt.Errorf("failed to get email credentials from environment variables")
	}

	// create a new dialer to connect to the smtp server
	dialer := mail.NewDialer(smtpHost, smtpPort, smtpUser, smtpPassword)

	// create a new email message
	m := mail.NewMessage()
	m.SetHeader("From", smtpUser)
	m.SetHeader("To", recipients...)
	m.SetHeader("Subject", subject)
	// set body as html to support embedded images
	m.SetBody("text/html", body)

	// attach multiple files
	for _, attachmentPath := range attachments {
		// check if file exists and is not a directory
		info, err := os.Stat(attachmentPath)
		if err != nil {
			return fmt.Errorf("error checking attachment file: %v", err)
		}
		if info.IsDir() {
			return fmt.Errorf("attachment path is a directory, not a file")
		}
		m.Attach(attachmentPath)
	}

	// embed images in the email body
	for cid, imagePath := range images {
		info, err := os.Stat(imagePath)
		if err != nil {
			return fmt.Errorf("error checking image file: %v", err)
		}
		if info.IsDir() {
			return fmt.Errorf("image path is a directory, not a file")
		}
		// embed the image with the provided cid
		m.Embed(imagePath, mail.SetHeader(map[string][]string{
			"Content-ID": {fmt.Sprintf("<%s>", cid)},
		}))
		body = strings.Replace(body, cid, fmt.Sprintf("cid:%s", cid), -1)
	}

	// send the email
	err = dialer.DialAndSend(m)
	if err != nil {
		return fmt.Errorf("error sending email: %v", err)
	}
	return err
}

func sendMailingListEmail(listName, subject, body string) error {
	subscribers, err := getSubscribers(listName)
	if err != nil {
		return err
	}

	//ensure there are subscribers in the mailing list
	if len(subscribers) == 0 {
		return fmt.Errorf("no subscribers found in mailing list: %s", listName)
	}

	// Send email to all subscribers
	err = sendMail(subscribers, subject, body, nil, nil)
	if err != nil {
		return fmt.Errorf("error sending mailing list email: %v", err)
	}

	return nil
}
