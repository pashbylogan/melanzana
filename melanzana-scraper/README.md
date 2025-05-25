# Melanzana Appointment Scraper

## Description

This tool scrapes the Melanzana website (`https://melanzana.com/book-an-appointment`) to check for available shopping appointments. If new appointments are found within a configured lookahead period, it can send an email notification.

It is designed to be run periodically, for example, using a cron job, to monitor for new appointment openings.

## Features

*   Fetches appointment availability from `https://melanzana.com/book-an-appointment`.
*   Identifies newly available appointments by comparing against previously seen appointments.
*   Filters appointments by a configurable lookahead period (e.g., only show appointments in the next 3 months).
*   Sends email notifications for newly available and relevant appointments.
*   Stores seen appointments in a JSON file to prevent duplicate notifications.
*   Highly configurable via a JSON configuration file and/or command-line flags.

## Prerequisites

*   Go (developed with Go 1.22+, should work with Go 1.18+ due to `goquery` and standard library usage).

## Installation / Setup

1.  Clone the repository (if you haven't already).
2.  Navigate to the `melanzana-scraper` directory.
3.  Build the executable:
    ```bash
    go build .
    ```
    This will create a `melanzana-scraper` executable in the current directory.

## Configuration

Configuration can be managed via a JSON file (typically `config.json`) and overridden by command-line flags. If a configuration option is set in both the file and as a flag, the flag's value will take precedence.

### `config.json` File

Create a `config.json` file in the same directory as the executable, or provide a path to a config file using the `-configFile` flag. You can use `config.example.json` as a template.

**Example `config.json`:**
```json
{
  "monthsLookahead": 3,
  "smtpServer": "smtp.example.com",
  "smtpPort": 587,
  "smtpUsername": "your_username@example.com",
  "smtpPassword": "your_smtp_password",
  "fromEmail": "scraper@example.com",
  "toEmails": ["your_email@example.com", "another_alert_email@example.com"],
  "dataFile": "seen_appointments.json"
}
```

**Configuration Fields:**

*   `monthsLookahead` (integer): Number of months to look ahead for appointments from the current date.
*   `smtpServer` (string): SMTP server address for email notifications.
*   `smtpPort` (integer): SMTP server port (e.g., 587 for TLS, 465 for SSL).
*   `smtpUsername` (string): Username for SMTP authentication.
*   `smtpPassword` (string): Password for SMTP authentication.
    *   **WARNING: SECURITY ADVISORY FOR `smtpPassword`**
        Storing plaintext passwords in configuration files is a security risk, especially in automated environments like cron jobs. For production or sensitive use:
        *   **Prefer Environment Variables:** Do not put the actual password in `config.json`. Instead, modify the scraper to read the password from an environment variable (e.g., `SMTP_PASSWORD`). This is a common and more secure practice.
        *   **Secrets Management Tools:** For more robust security, use a dedicated secrets management tool (e.g., HashiCorp Vault, AWS Secrets Manager, GCP Secret Manager).
        *   The tool, as provided, loads the `smtpPassword` directly from the configuration. It does **not** implement advanced secret protection mechanisms itself. Ensure the configuration file has appropriate file permissions if you must store the password there temporarily.
*   `fromEmail` (string): Email address to send notifications from.
*   `toEmails` (array of strings): List of email addresses to send notifications to.
*   `dataFile` (string): Path to the JSON file used for storing seen appointments (e.g., `seen_appointments.json`).

### Command-Line Flags

*   `-configFile <path>`: Path to the JSON configuration file.
*   `-months <int>`: Number of months to look ahead (overrides `monthsLookahead` in config file). (Default: 3)
*   `-smtpServer <string>`: SMTP server address.
*   `-smtpPort <int>`: SMTP server port. (Default: 587)
*   `-smtpUser <string>`: SMTP username.
*   `-smtpPass <string>`: SMTP password. **(Strongly discouraged for production use; see security advisory above)**.
*   `-fromEmail <string>`: Email address for sending notifications.
*   `-toEmails <string>`: Comma-separated list of recipient email addresses.
*   `-dataFile <string>`: Path to the data file for seen appointments. (Default: `seen_appointments.json`)

## Usage

Run the scraper from the command line:

*   **Basic (using default `config.json` if present, or default settings):**
    ```bash
    ./melanzana-scraper
    ```

*   **With a custom config file path:**
    ```bash
    ./melanzana-scraper -configFile /path/to/your/custom_config.json
    ```

*   **Overriding specific settings with flags:**
    ```bash
    ./melanzana-scraper -months 6 -toEmails "me@example.com,you@example.com"
    ```

The scraper will log its activities to standard output.

## Email Notifications

For email notifications to function correctly:

1.  You **must** provide valid SMTP server details (`smtpServer`, `smtpPort`, `smtpUsername`, `smtpPassword`, `fromEmail`) and at least one recipient in `toEmails` via the configuration file or command-line flags.
2.  **Important:** The actual call to `sendEmail()` in `main.go` (within the `runScrapingCycle` function) is **commented out by default** for safety. To enable email sending, you need to:
    *   Uncomment the line: `// err = sendEmail(emailConf, emailSubject, emailBody.String())`
    *   And the associated error handling block.
    *   Then, recompile the application: `go build .`

    This is a safety measure to prevent accidental email sending without explicit configuration and acknowledgment.

## Running as a Cron Job

To run the scraper periodically, you can set up a cron job.

**Example Cron Job Line (runs every hour at the 5-minute mark):**
```cron
# Run Melanzana scraper every hour at the 5-minute mark
5 * * * * /path/to/your/melanzana-scraper -configFile /path/to/your/config.json >> /path/to/your/scraper.log 2>&1
```

**Explanation:**

*   `5 * * * *`: Cron schedule (minute, hour, day of month, month, day of week). This means "at minute 5 of every hour, every day".
*   `/path/to/your/melanzana-scraper`: **Absolute path** to your compiled scraper executable.
*   `-configFile /path/to/your/config.json`: **Absolute path** to your configuration file. Using absolute paths in cron jobs is highly recommended.
*   `>> /path/to/your/scraper.log 2>&1`: Redirects standard output (`stdout`) and standard error (`stderr`) to a log file. This is useful for monitoring and debugging.

**Important for Cron:**

*   Always use absolute paths for the executable and any referenced files (like `config.json`, `dataFile`).
*   Ensure the user under which the cron job runs has the necessary permissions to execute the scraper and read/write the `dataFile` and log file.
*   If using environment variables for `smtpPassword` (recommended), ensure those variables are available in the cron execution environment.

## HTML Parsing Limitations

This scraper works by parsing the HTML structure of the Melanzana appointment booking page. If Melanzana significantly changes their website's HTML layout or class names for the calendar elements, the scraper's parsing logic (primarily in `scraper.go`) may break and require updates to function correctly.

## Development

To run the unit tests for this project:
```bash
go test ./...
```
This command will execute all test files (`*_test.go`) in the current directory and any subdirectories.
