package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"os"
	"strings"
	// Note: "time" is not directly used by loadConfig, but AppConfig might have time-related fields in a future version.
	// For now, it's not strictly needed here.
)

// AppConfig holds all application configuration parameters.
type AppConfig struct {
	MonthsLookahead  int      `json:"monthsLookahead"`
	SMTPServer       string   `json:"smtpServer"`
	SMTPPort         int      `json:"smtpPort"`
	SMTPUsername     string   `json:"smtpUsername"`
	SMTPPassword     string   `json:"smtpPassword"` // WARNING: Storing passwords in config is insecure. Use env vars or a proper secrets management system for production.
	FromEmail        string   `json:"fromEmail"`
	ToEmails         []string `json:"toEmails"`
	DataFile         string   `json:"dataFile"`
	// IntervalMinutes  int      `json:"intervalMinutes"` // Removed for single-run execution
	ConfigFile       string   // Not part of JSON, used to store path to config file loaded
}

// loadConfig loads configuration from file and command-line flags.
// Flags override file values, which override defaults.
func loadConfig() (AppConfig, error) {
	// Default configuration
	config := AppConfig{
		MonthsLookahead: 3, // Default value for MonthsLookahead
		SMTPServer:      "smtp.example.com",
		SMTPPort:        587,
		SMTPUsername:    "user",
		SMTPPassword:    "pass",
		FromEmail:       "scraper@example.com",
		ToEmails:        []string{"recipient@example.com"},
		DataFile:        "seen_appointments.json", // Default value for DataFile
		// IntervalMinutes: 60, // Removed
	}

	// Define command-line flag for config file path
	configFile := flag.String("configFile", "", "Path to JSON configuration file")
	
	// Define flags for each config option
	monthsLookaheadFlag := flag.Int("months", config.MonthsLookahead, "Number of months to look ahead for appointments")
	smtpServerFlag := flag.String("smtpServer", config.SMTPServer, "SMTP server address")
	smtpPortFlag := flag.Int("smtpPort", config.SMTPPort, "SMTP server port")
	smtpUserFlag := flag.String("smtpUser", config.SMTPUsername, "SMTP username")
	smtpPassFlag := flag.String("smtpPass", "", "SMTP password (leave empty to use config file value if set; for production, prefer environment variables or other secure means over direct command-line flags or config files for passwords)")
	fromEmailFlag := flag.String("fromEmail", config.FromEmail, "Email address to send notifications from")
	toEmailsFlag := flag.String("toEmails", strings.Join(config.ToEmails, ","), "Comma-separated list of email addresses to send notifications to")
	dataFileFlag := flag.String("dataFile", config.DataFile, "Path to the data file for seen appointments")

	flag.Parse() // Parse all command-line flags

	// Load from config file if path is provided

	// Load from config file if path is provided
	if *configFile != "" {
		config.ConfigFile = *configFile
		data, err := os.ReadFile(*configFile)
		if err != nil {
			// If config file is specified but not found, it's an error
			return AppConfig{}, fmt.Errorf("failed to read config file %s: %w", *configFile, err)
		}
		err = json.Unmarshal(data, &config)
		if err != nil {
			return AppConfig{}, fmt.Errorf("failed to unmarshal config file %s: %w", *configFile, err)
		}
		log.Printf("Loaded configuration from %s", *configFile)
	}
	
	// Update config with values from flags
	config.MonthsLookahead = *monthsLookaheadFlag
	config.SMTPServer = *smtpServerFlag
	config.SMTPPort = *smtpPortFlag
	config.SMTPUsername = *smtpUserFlag
	if *smtpPassFlag != "" { // Only update password if flag is explicitly set
		config.SMTPPassword = *smtpPassFlag
	}
	config.FromEmail = *fromEmailFlag
	if *toEmailsFlag != "" { // Process comma-separated string for ToEmails
		config.ToEmails = strings.Split(*toEmailsFlag, ",")
	}
	config.DataFile = *dataFileFlag
	
	return config, nil
}
