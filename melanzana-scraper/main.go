package main

import (
	"fmt"
	"log"
	"strings"
	"time" // Keep time for the main loop if it were periodic, but it's single run. Still used for logging.
	// Other imports are now in their respective files.
)

// Note: Global constants like seenAppointmentsFile and defaultMonthsAhead have been removed.
// Their values are now managed as defaults within AppConfig in config.go or directly via AppConfig fields.

func runScrapingCycle(config AppConfig) {
	log.Println("--- Starting new scraping cycle ---")

	pageURL := "https://melanzana.com/book-an-appointment" // This could also be a config option if it ever changes
	log.Printf("Fetching content from %s", pageURL)

	content, err := fetchPageContent(pageURL)
	if err != nil {
		log.Printf("Error fetching page content during cycle: %v", err)
		return
	}
	log.Println("Successfully fetched page content.")

	// Load seen appointments using config.DataFile
	currentSeenAppointments, err := loadSeenAppointments(config.DataFile)
	if err != nil {
		log.Printf("Error loading seen appointments during cycle: %v", err)
		currentSeenAppointments = []Appointment{}
	} else {
		log.Printf("Loaded %d seen appointments from %s", len(currentSeenAppointments), config.DataFile)
	}

	// Parse HTML content
	log.Println("Parsing HTML content for appointments...")
	scrapedAppointments, err := parseAppointments(content)
	if err != nil {
		log.Printf("Error parsing appointments during cycle: %v", err)
		return
	}

	if len(scrapedAppointments) > 0 {
		log.Printf("Scraped %d potential appointment days/slots from current fetch.", len(scrapedAppointments))
	} else {
		log.Println("No new appointment information found or parsed from current fetch.")
	}

	// Filter new and available appointments using config.MonthsLookahead
	log.Printf("Filtering appointments within %d months ahead...", config.MonthsLookahead)
	newAvailableAppointments := filterAppointments(scrapedAppointments, currentSeenAppointments, config.MonthsLookahead)

	if len(newAvailableAppointments) > 0 {
		log.Printf("Found %d NEW and AVAILABLE appointments:", len(newAvailableAppointments))
		emailBody := &strings.Builder{}
		fmt.Fprintln(emailBody, "New Melanzana appointments found:")
		for _, appt := range newAvailableAppointments {
			logMsg := fmt.Sprintf("- Month: %s, Day: %s, Time: %s, Status: Available", appt.Month, appt.Day, appt.Time)
			log.Println(logMsg) 
			fmt.Fprintf(emailBody, "- %s %s. More details: https://melanzana.com/book-an-appointment\n", appt.Month, appt.Day)
		}

		currentSeenAppointments = append(currentSeenAppointments, newAvailableAppointments...)

		emailConf := EmailConfig{
			SMTPHost:     config.SMTPServer,
			SMTPPort:     config.SMTPPort,
			SMTPUsername: config.SMTPUsername,
			SMTPPassword: config.SMTPPassword,
			FromEmail:    config.FromEmail,
			ToEmails:     config.ToEmails,
		}
		emailSubject := "New Melanzana Appointments Available!"
		
		// --- IMPORTANT: Email Sending Configuration ---
		// The following email sending logic is COMMENTED OUT BY DEFAULT.
		// To enable email notifications:
		// 1. Ensure your `config.json` (or command-line flags) provide real SMTP server details,
		//    username, password, from-email, and to-emails.
		// 2. Uncomment the call to `sendEmail` below.
		//
		// WARNING: Avoid hardcoding sensitive credentials directly in the source code for production.
		// Prefer using a configuration file (with appropriate permissions) or environment variables.
		// The `AppConfig.SMTPPassword` field should be handled securely.

		// err = sendEmail(emailConf, emailSubject, emailBody.String())
		// if err != nil {
		// 	log.Printf("Error sending email notification: %v", err)
		// } else {
		// 	log.Printf("Successfully sent email notification to %s.", strings.Join(emailConf.ToEmails, ", "))
		// }
		log.Println("Email sending is currently COMMENTED OUT. See comments in main.go to enable and configure.")

	} else {
		log.Println("No new available appointments found meeting the criteria.")
	}

	// Save all seen appointments (old + newly found) using config.DataFile
	err = saveSeenAppointments(currentSeenAppointments, config.DataFile)
	if err != nil {
		log.Printf("Error saving appointments to %s during cycle: %v", config.DataFile, err)
	} else {
		log.Printf("Successfully saved %d appointments to %s", len(currentSeenAppointments), config.DataFile)
	}
	log.Println("--- Scraping cycle finished ---")
}

func main() {
	appCfg, err := loadConfig()
	if err != nil {
		log.Fatalf("Error loading configuration: %v", err)
	}

	log.Printf("Melanzana Scraper Initialized. Effective Config: DataFile='%s', MonthsLookahead=%d",
		appCfg.DataFile, appCfg.MonthsLookahead)

	runScrapingCycle(appCfg)

	log.Println("Scraping cycle complete. Application will now exit.")
}
