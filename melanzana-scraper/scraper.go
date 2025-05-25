package main

import (
	"fmt"
	"io"
	"log"
	"net/http"
	"strings"

	"github.com/PuerkitoBio/goquery"
)

// Appointment holds information about a single appointment slot or day.
type Appointment struct {
	Month       string `json:"month"`
	Day         string `json:"day"`
	Time        string `json:"time"` // Placeholder for now
	IsAvailable bool   `json:"isAvailable"`
}

// fetchPageContent fetches the content of a given URL and returns it as a string.
func fetchPageContent(url string) (string, error) {
	resp, err := http.Get(url)
	if err != nil {
		return "", fmt.Errorf("failed to get URL: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("failed to fetch page, status code: %d", resp.StatusCode)
	}

	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("failed to read response body: %w", err)
	}

	return string(bodyBytes), nil
}

// parseAppointments attempts to extract appointment information from HTML content.
func parseAppointments(htmlContent string) ([]Appointment, error) {
	doc, err := goquery.NewDocumentFromReader(strings.NewReader(htmlContent))
	if err != nil {
		return nil, fmt.Errorf("failed to create goquery document: %w", err)
	}

	var appointments []Appointment
	var currentMonth string

	// Attempt 1: Find month and year from prominent headers (h1-h5, .entry-title, .page-title)
	// This assumes the month/year is displayed in a common heading element near the calendar.
	// Example: "Appointments for May 2025" or "May 2025"
	doc.Find("h1, h2, h3, h4, h5, .entry-title, .page-title").Each(func(i int, s *goquery.Selection) {
		text := s.Text()
		// Basic check for a year-like number (e.g., 2024, 2025, 2026)
		if strings.Contains(text, "202") { 
			currentMonth = strings.TrimSpace(text)
			// If the text is like "Appointments for May 2025", extract "May 2025".
			parts := strings.Split(currentMonth, " for ")
			if len(parts) > 1 {
				currentMonth = strings.TrimSpace(parts[len(parts)-1]) // Take the part after " for "
			}
			// Assuming the first such header found is the primary one for the current calendar view.
			return // Stop search after finding the first potential month heading
		}
	})

	// Attempt 2: Fallback to searching paragraph tags if no month found in headers.
	// This is less reliable but provides a backup if the site structure is unexpected.
	if currentMonth == "" {
		doc.Find("p").EachWithBreak(func(i int, s *goquery.Selection) bool {
			text := s.Text()
			hasYear := strings.Contains(text, "202") // Check for a year string like "2024", "2025"
			hasMonthName := false
			lowerText := strings.ToLower(text)
			// Check for any common month name substring.
			// This is a broad check and might lead to false positives if month names appear in other contexts.
			commonMonths := []string{"january", "february", "march", "april", "may", "june", "july", "august", "september", "october", "november", "december"}
			for _, m := range commonMonths {
				if strings.Contains(lowerText, m) {
					hasMonthName = true
					break
				}
			}
			if hasYear && hasMonthName {
				currentMonth = strings.TrimSpace(text)
				return false // Stop search
			}
			return true // Continue search
		})
	}

	if currentMonth == "" {
		log.Println("Warning: Could not determine current month from page headers or paragraphs. Appointments will use 'Unknown Month'.")
		currentMonth = "Unknown Month" // Default if no month information could be parsed
	}

	// Find potential day cells within common content containers.
	// This targets <td> (table cells) or <div> elements that might represent days.
	doc.Find(".entry-content, article, .post-content, .page-content").Find("td, div").Each(func(i int, s *goquery.Selection) {
		dayText := strings.TrimSpace(s.Text())

		// Check if the text is a number, likely representing a day of the month (1-31).
		if _, err := strconv.Atoi(dayText); err == nil && len(dayText) > 0 && len(dayText) <= 2 {
			// Determine availability based on CSS classes or presence of a link.
			// This logic is based on common calendar patterns and previous observations.
			var isAvailable bool // Default to false unless specific conditions met
			
			// Explicit "available" or "active" classes are strong indicators.
			if s.HasClass("available") || s.HasClass("active") {
				isAvailable = true
			// Explicit "booked", "unavailable", etc., classes indicate unavailability.
			} else if s.HasClass("booked") || s.HasClass("unavailable") || s.HasClass("disabled") || s.HasClass("past-day") || s.HasClass("grey") || s.HasClass("gray") || s.HasClass("red") {
				isAvailable = false
			// If no specific availability/unavailability classes, presence of an <a> tag often means it's clickable/available.
			} else if s.Find("a").Length() > 0 {
				isAvailable = true
			// Otherwise, assume unavailable.
			} else {
				isAvailable = false
			}

			appointments = append(appointments, Appointment{
				Month:       currentMonth,
				Day:         dayText,
				IsAvailable: isAvailable,
				IsAvailable: finalAvailable,
			})
		}
	})
	
	if len(appointments) == 0 {
		log.Println("No day numbers found with the current selectors. The HTML structure might be different than assumed.")
	}

	return appointments, nil
}
