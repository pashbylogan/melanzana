# Melanzana Appointment Scraper

## Description

This tool checks the Melanzana website for available shopping appointments using their Cowlendar booking API. If new appointments are found within a configured lookahead period, it can send an email notification.

The scraper uses Cowlendar's external API endpoint to efficiently check appointment availability across multiple months, with intelligent optimization to minimize unnecessary requests.

It is designed to be run periodically, for example, using a cron job, to monitor for new appointment openings.

## Features

* **Modern API integration**: Uses Cowlendar's external API for direct, reliable data retrieval
* **Intelligent optimization**: Automatically stops searching when next available date exceeds the configured timeframe
* **Month-based scanning**: Efficiently checks appointment availability by month rather than individual dates
* **Smart filtering**: Identifies newly available appointments by comparing against previously seen appointments
* **Configurable timeframe**: Filters appointments by a configurable lookahead period (e.g., next 3 months)
* **Email notifications**: Sends detailed email alerts for newly available appointments
* **Persistent storage**: Stores seen appointments in JSON file to prevent duplicate notifications
* **Highly configurable**: Manage settings via JSON configuration file and/or command-line flags

## Prerequisites

* Go (developed with Go 1.22+, should work with Go 1.18+ due to standard library usage).

## Installation / Setup

1. Clone the repository (if you haven't already).
2. Navigate to the `melanzana` directory.
3. Build the executable:

    ```bash
    go build .
    ```

    This will create a `melanzana` executable in the current directory.

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

* `monthsLookahead` (integer): Number of months to look ahead for appointments from the current date.
* `smtpServer` (string): SMTP server address for email notifications.
* `smtpPort` (integer): SMTP server port (e.g., 587 for TLS, 465 for SSL).
* `smtpUsername` (string): Username for SMTP authentication.
* `smtpPassword` (string): Password for SMTP authentication.
  * **WARNING: SECURITY ADVISORY FOR `smtpPassword`**
        Storing plaintext passwords in configuration files is a security risk, especially in automated environments like cron jobs. For production or sensitive use:
    * **Prefer Environment Variables:** Do not put the actual password in `config.json`. Instead, modify the scraper to read the password from an environment variable (e.g., `SMTP_PASSWORD`). This is a common and more secure practice.
    * **Secrets Management Tools:** For more robust security, use a dedicated secrets management tool (e.g., HashiCorp Vault, AWS Secrets Manager, GCP Secret Manager).
    * The tool, as provided, loads the `smtpPassword` directly from the configuration. It does **not** implement advanced secret protection mechanisms itself. Ensure the configuration file has appropriate file permissions if you must store the password there temporarily.
* `fromEmail` (string): Email address to send notifications from.
* `toEmails` (array of strings): List of email addresses to send notifications to.
* `dataFile` (string): Path to the JSON file used for storing seen appointments (e.g., `seen_appointments.json`). The file stores appointment data including date, time, and number of available spaces.

### Command-Line Flags

* `-configFile <path>`: Path to the JSON configuration file.
* `-months <int>`: Number of months to look ahead (overrides `monthsLookahead` in config file). (Default: 3)
* `-smtpServer <string>`: SMTP server address.
* `-smtpPort <int>`: SMTP server port. (Default: 587)
* `-smtpUser <string>`: SMTP username.
* `-smtpPass <string>`: SMTP password. **(Strongly discouraged for production use; see security advisory above)**.
* `-fromEmail <string>`: Email address for sending notifications.
* `-toEmails <string>`: Comma-separated list of recipient email addresses.
* `-dataFile <string>`: Path to the data file for seen appointments. (Default: `seen_appointments.json`)

## Usage

Run the scraper from the command line:

* **Basic (using default `config.json` if present, or default settings):**

    ```bash
    ./melanzana
    ```

* **With a custom config file path:**

    ```bash
    ./melanzana -configFile /path/to/your/custom_config.json
    ```

* **Overriding specific settings with flags:**

    ```bash
    ./melanzana -months 6 -toEmails "me@example.com,you@example.com"
    ```

The scraper will log its activities to standard output, showing:

* Monthly availability checking progress
* API optimization decisions (early termination when appropriate)
* Number of appointment slots found per month
* Next availability dates from the API
* New appointments discovered
* Email notification status

**Example output:**

```
2025/09/23 21:38:16 Melanzana Scraper - Checking 3 months ahead
2025/09/23 21:38:16 --- Starting scraping cycle ---
2025/09/23 21:38:16 Checking availability for 2025-09
2025/09/23 21:38:16 Next availability 2026-04-16 is beyond threshold 2025-12-23 - stopping search
2025/09/23 21:38:16 Total available appointments found: 0
2025/09/23 21:38:16 No new appointments found
2025/09/23 21:38:16 --- Scraping cycle complete ---
```

## Email Notifications

For email notifications to function correctly:

1. You **must** provide valid SMTP server details (`smtpServer`, `smtpPort`, `smtpUsername`, `smtpPassword`, `fromEmail`) and at least one recipient in `toEmails` via the configuration file or command-line flags.
2. **Important:** The actual call to `sendEmail()` in `main.go` (within the `runScrapingCycle` function) is **commented out by default** for safety. To enable email sending, you need to:
    * Uncomment the line: `// err = sendEmail(emailConf, emailSubject, emailBody.String())`
    * And the associated error handling block.
    * Then, recompile the application: `go build .`

    This is a safety measure to prevent accidental email sending without explicit configuration and acknowledgment.

## Running as a Cron Job

To run the scraper periodically, you can set up a cron job.

**Example Cron Job Line (runs every hour at the 5-minute mark):**

```cron
# Run Melanzana scraper every hour at the 5-minute mark
5 * * * * /path/to/your/melanzana -configFile /path/to/your/config.json >> /path/to/your/scraper.log 2>&1
```

**Explanation:**

* `5 * * * *`: Cron schedule (minute, hour, day of month, month, day of week). This means "at minute 5 of every hour, every day".
* `/path/to/your/melanzana`: **Absolute path** to your compiled scraper executable.
* `-configFile /path/to/your/config.json`: **Absolute path** to your configuration file. Using absolute paths in cron jobs is highly recommended.
* `>> /path/to/your/scraper.log 2>&1`: Redirects standard output (`stdout`) and standard error (`stderr`) to a log file. This is useful for monitoring and debugging.

**Important for Cron:**

* Always use absolute paths for the executable and any referenced files (like `config.json`, `dataFile`).
* Ensure the user under which the cron job runs has the necessary permissions to execute the scraper and read/write the `dataFile` and log file.
* If using environment variables for `smtpPassword` (recommended), ensure those variables are available in the cron execution environment.

## How It Works

The scraper operates by:

1. **Month Iteration**: Iterates through the configured number of months ahead from the current date
2. **API Requests**: Makes GET requests to `https://app.cowlendar.com/extapi/calendar/685b42f202405a8372cd6b78/availability` with:
   * `year=YYYY`
   * `month=MM`
   * `timezone=America/Denver`
   * Various booking configuration parameters
3. **Optimization Check**: Examines the `next_availability` field in the API response - if it's beyond the configured threshold, stops searching to save unnecessary API calls
4. **Response Parsing**: Processes the JSON response to extract detailed appointment slot information including times and availability counts
5. **Change Detection**: Compares found appointments against previously seen ones
6. **Notifications**: Sends email alerts for any new available appointments

## API Limitations

This scraper relies on Cowlendar's external API that Melanzana uses for their booking system. If they:

* Change booking platforms
* Update the Cowlendar calendar ID or variant ID
* Modify API parameters or response format
* Implement rate limiting or access restrictions

The scraper may require updates to continue functioning. However, this API-based approach using a dedicated booking service is very stable and robust.

**Current API endpoint:** `https://app.cowlendar.com/extapi/calendar/685b42f202405a8372cd6b78/availability`

**Key identifiers:**
* Calendar ID: `685b42f202405a8372cd6b78`
* Variant ID: `41855678382123` (appears to be optional/flexible)

## Development

### Running Tests

This project includes comprehensive test coverage for all core functionality. To run the tests:

**Run all tests:**
```bash
go test -v
```

**Run tests with coverage report:**
```bash
go test -cover
```

**Run specific test file:**
```bash
go test -run TestFilterNewAppointments -v
```

### Test Coverage

The test suite covers:

- **Filter functionality** (`filter_test.go`): Tests appointment filtering logic, including handling of new vs. seen appointments
- **Storage functionality** (`storage_test.go`): Tests JSON file operations for loading and saving appointment data, including edge cases like malformed files and large datasets
- **Scraper functionality** (`scraper_test.go`): Tests HTML parsing, date range generation, email body building, and space extraction from text

**Key test scenarios:**
- Empty appointment lists and files
- Malformed JSON handling
- Large dataset processing
- HTML parsing edge cases
- Date range generation across different periods
- Email notification content generation

All tests use temporary files and mock data to avoid external dependencies.
