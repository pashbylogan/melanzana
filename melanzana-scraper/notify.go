package main

import (
	"fmt"
	// "log" // Not strictly needed if sendEmail just returns errors
	"net/smtp"
	"strings"
)

// EmailConfig holds SMTP server details and recipient information.
// This struct is populated from AppConfig in main.go when sending email.
type EmailConfig struct {
	SMTPHost     string
	SMTPPort     int
	SMTPUsername string
	SMTPPassword string
	FromEmail    string
	ToEmails     []string
}

// sendEmail constructs and sends an email.
func sendEmail(config EmailConfig, subject string, body string) error {
	auth := smtp.PlainAuth("", config.SMTPUsername, config.SMTPPassword, config.SMTPHost)

	msg := strings.Builder{}
	msg.WriteString("From: " + config.FromEmail + "\r\n")
	msg.WriteString("To: " + strings.Join(config.ToEmails, ",") + "\r\n")
	msg.WriteString("Subject: " + subject + "\r\n")
	msg.WriteString("\r\n") // Empty line separates headers from body
	msg.WriteString(body + "\r\n")

	addr := fmt.Sprintf("%s:%d", config.SMTPHost, config.SMTPPort)
	err := smtp.SendMail(addr, auth, config.FromEmail, config.ToEmails, []byte(msg.String()))
	if err != nil {
		return fmt.Errorf("failed to send email: %w", err)
	}
	return nil
}
