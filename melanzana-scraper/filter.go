package main

import (
	"fmt"
	"log"
	"strconv"
	"strings"
	"time"
)

// isNewAppointment checks if an appointment is new based on Month and Day.
func isNewAppointment(appointment Appointment, seenAppointments []Appointment) bool {
	for _, seen := range seenAppointments {
		if seen.Month == appointment.Month && seen.Day == appointment.Day {
			// Could extend to check Time if it becomes relevant
			return false
		}
	}
	return true
}

// monthNameToNumber converts a month name (e.g., "January", "jan", or part of "January 2024") to its time.Month number.
// It attempts to be robust by checking common month names and recursively trying shorter versions if a direct match fails.
func monthNameToNumber(monthName string) (time.Month, error) {
	cleanedMonthName := strings.TrimSpace(strings.ToLower(monthName))
	
	// Direct match for full month names
	switch cleanedMonthName {
	case "january":
		return time.January, nil
	case "february":
		return time.February, nil
	case "march":
		return time.March, nil
	case "april":
		return time.April, nil
	case "may":
		return time.May, nil
	case "june":
		return time.June, nil
	case "july":
		return time.July, nil
	case "august":
		return time.August, nil
	case "september":
		return time.September, nil
	case "october":
		return time.October, nil
	case "november":
		return time.November, nil
	case "december":
		return time.December, nil
	}

	// If no direct match, try parsing from a multi-word string (e.g., "May 2025", "March,")
	// This handles cases where the month might be the first word of the input `monthName`.
	parts := strings.Fields(cleanedMonthName)
	if len(parts) > 1 { // Only proceed if there's more than one word and it wasn't a direct match
		firstWord := parts[0]
		// Clean common trailing punctuation from the first word before recursive call.
		firstWord = strings.TrimRight(firstWord, ",.;:") 
		// Recursive call with just the first word. If this call itself fails, the error will propagate up.
		// This check `firstWord != cleanedMonthName` prevents infinite recursion if `cleanedMonthName` was already a single, unknown word.
		if firstWord != cleanedMonthName { 
			if m, err := monthNameToNumber(firstWord); err == nil {
				return m, nil
			}
		}
	}
	
	// If all attempts fail, return an error.
	return time.January, fmt.Errorf("unknown month name: %q", monthName) // Use %q for better error message with original input
}

// filterAppointments filters appointments based on availability, newness, and date range.
func filterAppointments(appointments []Appointment, seenAppointments []Appointment, monthsAhead int) []Appointment {
	var filtered []Appointment
	now := time.Now()
	currentYear := now.Year()
	currentMonth := now.Month() // This is time.Month, e.g., time.January, time.February
	
	// Calculate the end date for the filtering window (exclusive of the day itself, effectively monthsAhead + 1 day at 00:00)
	// To make it inclusive of the last day of the `monthsAhead` month, we can go to the start of the next month.
	// For example, if monthsAhead is 3, and current is May, we want to include up to end of August.
	// So, targetDate will be September 1st.
	targetDate := time.Date(currentYear, currentMonth, 1, 0, 0, 0, 0, time.UTC).AddDate(0, monthsAhead+1, 0)


	for _, appt := range appointments {
		// Skip if not available or already seen
		if !appt.IsAvailable {
			continue
		}
		if !isNewAppointment(appt, seenAppointments) {
			continue
		}

		// Convert Day string to integer
		day, err := strconv.Atoi(appt.Day)
		if err != nil {
			log.Printf("Error converting day '%s' to int for appointment %+v: %v", appt.Day, appt, err)
			continue
		}

		// Convert parsed Month string (e.g., "May 2025") to time.Month
		monthNum, err := monthNameToNumber(appt.Month)
		if err != nil {
			log.Printf("Error converting month name '%s' for appointment %+v: %v", appt.Month, appt, err)
			continue
		}
		
		// --- Determine the Year for the appointment ---
		// Start with the current year as a baseline.
		year := currentYear 
		
		// Attempt to parse an explicit year from the appointment's month string (e.g., "May 2025").
		// This is the most reliable source for the year if provided.
		monthStrParts := strings.Fields(appt.Month)
		parsedExplicitYear := 0
		if len(monthStrParts) > 1 {
			lastPart := monthStrParts[len(monthStrParts)-1]
			if pYear, errAtoi := strconv.Atoi(lastPart); errAtoi == nil && pYear >= currentYear && pYear < currentYear+5 { // Sanity check for a reasonable future year
				parsedExplicitYear = pYear
			}
		}

		if parsedExplicitYear > 0 {
			year = parsedExplicitYear
		} else {
			// If no explicit year, and the appointment month is numerically less than the current month,
			// assume it's for the next year. (e.g., current is Dec, appt month is Jan -> next year).
			// This handles calendar rollovers when only month names are displayed without explicit years.
			if monthNum < currentMonth {
				year = currentYear + 1
			}
			// Otherwise, if monthNum >= currentMonth, assume currentYear (already set).
		}
		// --- End Year Determination ---

		apptDate := time.Date(year, monthNum, day, 0, 0, 0, 0, time.UTC)

		// Check if the appointment date is within the desired range:
		// It must be on or after today (ignoring time part for 'now')
		// and strictly before the targetDate (start of the month after the `monthsAhead` period).
		today := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, time.UTC)
		if (apptDate.Equal(today) || apptDate.After(today)) && apptDate.Before(targetDate) {
			filtered = append(filtered, appt)
		}
	}
	return filtered
}
