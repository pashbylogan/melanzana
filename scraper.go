package main

import (
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/PuerkitoBio/goquery"
)

const (
	bookingURL   = "https://melanzana.com/book-an-appointment"
	ajaxURL      = "https://melanzana.com/wp-admin/admin-ajax.php"
	requestDelay = 100 * time.Millisecond
	daysPerMonth = 30
	calendarID   = "0"
)

var (
	noncePatterns = []string{
		`booked_js_vars.*?"nonce":\s*"([^"]+)"`,
		`booked_wc_variables.*?"nonce":\s*"([^"]+)"`,
		`"nonce":\s*"([^"]+)"`,
	}
	timePattern  = regexp.MustCompile(`(\d{1,2}:\d{2}\s+[ap]m)\s*[–-]\s*(\d{1,2}:\d{2}\s+[ap]m)`)
	spacePattern = regexp.MustCompile(`(\d+)\s+spaces?\s+available`)
)

// Appointment holds information about a single appointment slot.
type Appointment struct {
	Date        string `json:"date"`        // YYYY-MM-DD format
	Time        string `json:"time"`        // e.g., "10:30 am – 11:00 am"
	Spaces      int    `json:"spaces"`      // number of available spaces
	IsAvailable bool   `json:"isAvailable"` // whether any appointments are available
}

// extractNonce fetches the booking page and extracts the current nonce value.
func extractNonce() (string, error) {
	resp, err := http.Get(bookingURL)
	if err != nil {
		return "", fmt.Errorf("failed to fetch booking page: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("booking page returned status %d", resp.StatusCode)
	}

	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("failed to read booking page: %w", err)
	}

	content := string(bodyBytes)
	for _, pattern := range noncePatterns {
		if re := regexp.MustCompile(pattern); re != nil {
			if matches := re.FindStringSubmatch(content); len(matches) >= 2 {
				return matches[1], nil
			}
		}
	}

	return "", fmt.Errorf("nonce not found in booking page")
}

// checkDateAvailability makes a POST request to check appointment availability for a specific date.
func checkDateAvailability(date, nonce string) ([]Appointment, error) {
	formData := url.Values{
		"action":      {"booked_calendar_date"},
		"nonce":       {nonce},
		"date":        {date},
		"calendar_id": {calendarID},
	}

	resp, err := http.PostForm(ajaxURL, formData)
	if err != nil {
		return nil, fmt.Errorf("API request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("API returned status %d", resp.StatusCode)
	}

	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read API response: %w", err)
	}

	responseText := string(bodyBytes)

	// Check for known error conditions
	if strings.Contains(responseText, "Required \"nonce\" value is not here") {
		return nil, fmt.Errorf("invalid nonce")
	}

	if strings.Contains(responseText, "There are no appointment time slots available") {
		return []Appointment{}, nil
	}

	return parseAppointmentSlots(responseText, date)
}

// parseAppointmentSlots extracts appointment time slots from the HTML response.
func parseAppointmentSlots(htmlContent, date string) ([]Appointment, error) {
	doc, err := goquery.NewDocumentFromReader(strings.NewReader(htmlContent))
	if err != nil {
		return nil, fmt.Errorf("failed to parse HTML: %w", err)
	}

	var appointments []Appointment

	// Target specific timeslot containers instead of searching all elements
	doc.Find(".timeslot").Each(func(i int, s *goquery.Selection) {
		// Extract time from the timeslot-range span
		timeText := strings.TrimSpace(s.Find(".timeslot-range").Text())

		if !timePattern.MatchString(timeText) {
			return
		}

		timeMatch := timePattern.FindStringSubmatch(timeText)
		if len(timeMatch) < 3 {
			return
		}

		timeSlot := fmt.Sprintf("%s – %s", timeMatch[1], timeMatch[2])

		// Extract spaces from the spots-available span within timeslot-time
		spacesText := strings.TrimSpace(s.Find(".timeslot-time .spots-available").Text())
		spaces := extractSpaces(spacesText)

		appointments = append(appointments, Appointment{
			Date:        date,
			Time:        timeSlot,
			Spaces:      spaces,
			IsAvailable: spaces > 0,
		})
	})

	return appointments, nil
}

// extractSpaces extracts the number of available spaces from text.
func extractSpaces(text string) int {
	if spaceMatch := spacePattern.FindStringSubmatch(text); len(spaceMatch) >= 2 {
		if spaces, err := strconv.Atoi(spaceMatch[1]); err == nil {
			return spaces
		}
	}
	return 0
}

// generateDateRange creates a slice of dates from start date for the specified number of days.
func generateDateRange(startDate time.Time, days int) []string {
	dates := make([]string, days)
	for i := 0; i < days; i++ {
		dates[i] = startDate.AddDate(0, 0, i).Format("2006-01-02")
	}
	return dates
}

// scrapeAppointments checks appointment availability across a range of dates.
func scrapeAppointments(monthsAhead int) ([]Appointment, error) {
	log.Println("Extracting nonce from booking page...")
	nonce, err := extractNonce()
	if err != nil {
		return nil, fmt.Errorf("failed to extract nonce: %w", err)
	}

	daysToCheck := monthsAhead * daysPerMonth
	dates := generateDateRange(time.Now(), daysToCheck)

	log.Printf("Checking appointments for %d dates (%d months ahead)", len(dates), monthsAhead)

	var allAppointments []Appointment

	for i, date := range dates {
		if i > 0 {
			time.Sleep(requestDelay)
		}

		appointments, err := checkDateAvailability(date, nonce)
		if err != nil {
			if strings.Contains(err.Error(), "invalid nonce") {
				log.Println("Nonce expired, refreshing...")
				if nonce, err = extractNonce(); err != nil {
					log.Printf("Failed to refresh nonce: %v", err)
					continue
				}
				if appointments, err = checkDateAvailability(date, nonce); err != nil {
					log.Printf("Error checking date %s after nonce refresh: %v", date, err)
					continue
				}
			} else {
				log.Printf("Error checking date %s: %v", date, err)
				continue
			}
		}

		// Only add available appointments
		for _, appt := range appointments {
			if appt.IsAvailable {
				allAppointments = append(allAppointments, appt)
			}
		}

		if len(appointments) > 0 {
			log.Printf("Found %d appointment slots for %s", len(appointments), date)
		}
	}

	log.Printf("Total available appointments found: %d", len(allAppointments))
	return allAppointments, nil
}
