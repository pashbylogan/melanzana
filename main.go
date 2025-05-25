package main

import (
	"fmt"
	"log"
	"strings"
)

func runScrapingCycle(config AppConfig) {
	log.Println("--- Starting scraping cycle ---")

	// Load seen appointments
	seenAppointments, err := loadSeenAppointments(config.DataFile)
	if err != nil {
		log.Printf("Error loading seen appointments: %v", err)
		seenAppointments = []Appointment{}
	} else {
		log.Printf("Loaded %d seen appointments", len(seenAppointments))
	}

	// Scrape current appointments
	log.Printf("Scraping appointments for %d months ahead...", config.MonthsLookahead)
	scrapedAppointments, err := scrapeAppointments(config.MonthsLookahead)
	if err != nil {
		log.Printf("Error scraping appointments: %v", err)
		return
	}

	log.Printf("Found %d available appointment slots", len(scrapedAppointments))

	// Filter for new appointments
	newAppointments := filterNewAppointments(scrapedAppointments, seenAppointments)

	if len(newAppointments) > 0 {
		log.Printf("Found %d NEW appointments:", len(newAppointments))

		logNewAppointments(newAppointments)

		// Email sending is commented out by default
		// Uncomment and configure the following lines to enable email notifications:
		//
		// emailBody := buildEmailBody(newAppointments)
		// if err := sendEmailNotification(config, emailBody); err != nil {
		// 	log.Printf("Error sending email: %v", err)
		// } else {
		// 	log.Println("Email notification sent successfully")
		// }

		log.Println("Email notifications are disabled. See main.go to enable.")

		// Update seen appointments
		seenAppointments = append(seenAppointments, newAppointments...)
	} else {
		log.Println("No new appointments found")
	}

	// Save seen appointments
	if err := saveSeenAppointments(seenAppointments, config.DataFile); err != nil {
		log.Printf("Error saving appointments: %v", err)
	} else {
		log.Printf("Saved %d appointments to %s", len(seenAppointments), config.DataFile)
	}

	log.Println("--- Scraping cycle complete ---")
}

func buildEmailBody(appointments []Appointment) string {
	var body strings.Builder
	body.WriteString("New Melanzana appointments found:\n\n")

	for _, appt := range appointments {
		fmt.Fprintf(&body, "- %s at %s (%d spaces available)\n",
			appt.Date, appt.Time, appt.Spaces)
	}

	body.WriteString("\nBook at: https://melanzana.com/book-an-appointment")
	return body.String()
}

func logNewAppointments(appointments []Appointment) {
	for _, appt := range appointments {
		log.Printf("- %s at %s (%d spaces)", appt.Date, appt.Time, appt.Spaces)
	}
}

func sendEmailNotification(config AppConfig, body string) error {
	emailConf := EmailConfig{
		SMTPHost:     config.SMTPServer,
		SMTPPort:     config.SMTPPort,
		SMTPUsername: config.SMTPUsername,
		SMTPPassword: config.SMTPPassword,
		FromEmail:    config.FromEmail,
		ToEmails:     config.ToEmails,
	}

	return sendEmail(emailConf, "New Melanzana Appointments Available!", body)
}

func main() {
	config, err := loadConfig()
	if err != nil {
		log.Fatalf("Failed to load configuration: %v", err)
	}

	log.Printf("Melanzana Scraper - Checking %d months ahead", config.MonthsLookahead)
	runScrapingCycle(config)
}
