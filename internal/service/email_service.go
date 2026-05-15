package service

import (
	"fmt"
	"log"
	"net/smtp"
	"os"
)

type EmailService struct{}

func NewEmailService() *EmailService {
	return &EmailService{}
}

func (s *EmailService) SendEmail(to, title, body string) error {
	host := os.Getenv("SMTP_HOST")
	port := os.Getenv("SMTP_PORT")
	user := os.Getenv("SMTP_USER")
	password := os.Getenv("SMTP_PASSWORD")

	if host == "" || user == "" {
		log.Printf("[SMTP Skip] Email not configured, mock sending: To: %s, Title: %s", to, title)
		return nil
	}

	auth := smtp.PlainAuth("", user, password, host)
	msg := []byte(fmt.Sprintf("To: %s\r\nSubject: %s\r\n\r\n%s", to, title, body))

	err := smtp.SendMail(host+":"+port, auth, user, []string{to}, msg)
	if err != nil {
		log.Printf("[SMTP Error] Failed to send email: %v", err)
		return err
	}

	log.Printf("[SMTP Success] Email sent to %s", to)
	return nil
}
