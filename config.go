package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"os"
	"strings"
)

// AppConfig holds all application configuration parameters.
type AppConfig struct {
	MonthsLookahead int      `json:"monthsLookahead"`
	SMTPServer      string   `json:"smtpServer"`
	SMTPPort        int      `json:"smtpPort"`
	SMTPUsername    string   `json:"smtpUsername"`
	SMTPPassword    string   `json:"smtpPassword"`
	FromEmail       string   `json:"fromEmail"`
	ToEmails        []string `json:"toEmails"`
	DataFile        string   `json:"dataFile"`
	ConfigFile      string   // Not part of JSON, used to store path to config file loaded
}

// loadConfig loads configuration from file and command-line flags.
// Flags override file values, which override defaults.
func loadConfig() (AppConfig, error) {
	config := AppConfig{
		MonthsLookahead: 3,
		SMTPServer:      "smtp.example.com",
		SMTPPort:        587,
		SMTPUsername:    "user",
		SMTPPassword:    "pass",
		FromEmail:       "scraper@example.com",
		ToEmails:        []string{"recipient@example.com"},
		DataFile:        "seen_appointments.json",
	}

	// Define command-line flags
	configFile := flag.String("configFile", "", "Path to JSON configuration file")
	monthsFlag := flag.Int("months", config.MonthsLookahead, "Number of months to look ahead")
	smtpServerFlag := flag.String("smtpServer", config.SMTPServer, "SMTP server address")
	smtpPortFlag := flag.Int("smtpPort", config.SMTPPort, "SMTP server port")
	smtpUserFlag := flag.String("smtpUser", config.SMTPUsername, "SMTP username")
	smtpPassFlag := flag.String("smtpPass", "", "SMTP password")
	fromEmailFlag := flag.String("fromEmail", config.FromEmail, "From email address")
	toEmailsFlag := flag.String("toEmails", strings.Join(config.ToEmails, ","), "Comma-separated recipient emails")
	dataFileFlag := flag.String("dataFile", config.DataFile, "Path to appointments data file")

	flag.Parse()

	// Load from config file if specified
	if *configFile != "" {
		config.ConfigFile = *configFile
		if err := loadConfigFile(&config, *configFile); err != nil {
			return AppConfig{}, err
		}
	}

	// Apply command-line flag overrides only if explicitly set
	flag.Visit(func(f *flag.Flag) {
		switch f.Name {
		case "months":
			config.MonthsLookahead = *monthsFlag
		case "smtpServer":
			config.SMTPServer = *smtpServerFlag
		case "smtpPort":
			config.SMTPPort = *smtpPortFlag
		case "smtpUser":
			config.SMTPUsername = *smtpUserFlag
		case "smtpPass":
			config.SMTPPassword = *smtpPassFlag
		case "fromEmail":
			config.FromEmail = *fromEmailFlag
		case "toEmails":
			config.ToEmails = strings.Split(*toEmailsFlag, ",")
		case "dataFile":
			config.DataFile = *dataFileFlag
		}
	})

	return config, nil
}

// loadConfigFile loads configuration from a JSON file.
func loadConfigFile(config *AppConfig, filename string) error {
	data, err := os.ReadFile(filename)
	if err != nil {
		return fmt.Errorf("failed to read config file %s: %w", filename, err)
	}

	if err := json.Unmarshal(data, config); err != nil {
		return fmt.Errorf("failed to parse config file %s: %w", filename, err)
	}

	log.Printf("Loaded configuration from %s", filename)
	return nil
}
