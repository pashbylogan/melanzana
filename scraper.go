package main

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"time"
)

const (
	cowlendarURL = "https://app.cowlendar.com/extapi/calendar/685b42f202405a8372cd6b78/availability"
	requestDelay = 100 * time.Millisecond
)

// CowlendarResponse represents the API response structure
type CowlendarResponse struct {
	Short                  []string       `json:"short"`
	Long                   []DetailedSlot `json:"long"`
	MaxDate                string         `json:"max_date"`
	NextAvailability       string         `json:"next_availability"`
	NoAvailabilityInFuture bool           `json:"no_availability_in_futur"`
	TargetTimezone         string         `json:"target_timezone"`
	NextUnix               *int64         `json:"next_unix"`
	JumpToNextAvs          bool           `json:"jump_to_next_avs"`
}

// DetailedSlot represents a detailed time slot from the "long" array
type DetailedSlot struct {
	Slot         string `json:"slot"`
	SlotStart    string `json:"slot_start"`
	SlotEnd      string `json:"slot_end"`
	SlotDuration int    `json:"slot_duration"`
	IsBookable   bool   `json:"is_bookable"`
	QtyBooked    int    `json:"qty_booked"`
	QtyLeft      int    `json:"qty_left"`
	MaxQty       int    `json:"max_qty"`
}

// Appointment holds information about a single appointment slot.
type Appointment struct {
	Date        string `json:"date"`        // YYYY-MM-DD format
	Time        string `json:"time"`        // e.g., "10:30 am – 11:00 am"
	Spaces      int    `json:"spaces"`      // number of available spaces
	IsAvailable bool   `json:"isAvailable"` // whether any appointments are available
}

// fetchAvailability fetches appointment availability for a specific month from Cowlendar API
func fetchAvailability(year, month int) (*CowlendarResponse, error) {
	url := fmt.Sprintf("%s?year=%d&month=%d&timezone=America/Denver&quantity_details[0][type]=default&quantity_details[0][quantity]=1&quantity_details[0][name]=Default&teammate_id=all&duration=30&is_manual=false&variant_id=41855678382123",
		cowlendarURL, year, month)

	resp, err := http.Get(url)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch availability: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("API returned status %d", resp.StatusCode)
	}

	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	var response CowlendarResponse
	if err := json.Unmarshal(bodyBytes, &response); err != nil {
		return nil, fmt.Errorf("failed to parse JSON response: %w", err)
	}

	return &response, nil
}

// convertCowlendarToAppointments converts Cowlendar response to our Appointment format
func convertCowlendarToAppointments(response *CowlendarResponse) []Appointment {
	var appointments []Appointment

	// Process detailed slots from "long" array
	for _, slot := range response.Long {
		if !slot.IsBookable || slot.QtyLeft <= 0 {
			continue
		}

		// Parse date and time from slot_start and slot_end
		startTime, err := time.Parse("2006-01-02 15:04", slot.SlotStart)
		if err != nil {
			log.Printf("Error parsing start time %s: %v", slot.SlotStart, err)
			continue
		}

		endTime, err := time.Parse("2006-01-02 15:04", slot.SlotEnd)
		if err != nil {
			log.Printf("Error parsing end time %s: %v", slot.SlotEnd, err)
			continue
		}

		// Format times for display
		timeSlot := fmt.Sprintf("%s – %s",
			startTime.Format("3:04 pm"),
			endTime.Format("3:04 pm"))

		appointments = append(appointments, Appointment{
			Date:        startTime.Format("2006-01-02"),
			Time:        timeSlot,
			Spaces:      slot.QtyLeft,
			IsAvailable: slot.QtyLeft > 0,
		})
	}

	return appointments
}

// scrapeAppointments checks appointment availability using the Cowlendar API
func scrapeAppointments(monthsAhead int) ([]Appointment, error) {
	var allAppointments []Appointment
	currentTime := time.Now()
	thresholdDate := currentTime.AddDate(0, monthsAhead, 0)

	// Check each month ahead
	for i := 0; i < monthsAhead; i++ {
		targetDate := currentTime.AddDate(0, i, 0)
		year := targetDate.Year()
		month := int(targetDate.Month())

		log.Printf("Checking availability for %d-%02d", year, month)

		response, err := fetchAvailability(year, month)
		if err != nil {
			log.Printf("Error fetching availability for %d-%02d: %v", year, month, err)
			continue
		}

		// Check if next availability is beyond our search threshold
		if response.NextAvailability != "" {
			nextAvailable, err := time.Parse("2006-01-02", response.NextAvailability)
			if err == nil && nextAvailable.After(thresholdDate) {
				log.Printf("Next availability %s is beyond threshold %s - stopping search",
					response.NextAvailability, thresholdDate.Format("2006-01-02"))
				break
			}
		}

		appointments := convertCowlendarToAppointments(response)
		if len(appointments) > 0 {
			log.Printf("Found %d appointment slots for %d-%02d", len(appointments), year, month)
			allAppointments = append(allAppointments, appointments...)
		} else {
			log.Printf("No appointments available for %d-%02d", year, month)
			if response.NextAvailability != "" {
				log.Printf("Next availability: %s", response.NextAvailability)
			}
		}

		if i < monthsAhead-1 {
			time.Sleep(requestDelay)
		}
	}

	log.Printf("Total available appointments found: %d", len(allAppointments))
	return allAppointments, nil
}
