package main

import (
	"reflect"
	"testing"
	"time"
)

func TestMonthNameToNumber(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		want    time.Month
		wantErr bool
	}{
		{"Full January", "January", time.January, false},
		{"Full May", "May", time.May, false},
		{"Full December", "December", time.December, false},
		{"Lowercase february", "february", time.February, false},
		{"Mixed Case MarCH", "MarCH", time.March, false},
		{"With Year April 2024", "April 2024", time.April, false},
		{"With Year and comma May, 2025", "May, 2025", time.May, false},
		{"Short month Jun", "Jun", time.June, true}, // Assuming current logic doesn't handle short names
		{"Invalid Month", "InvalidMonth", time.January, true}, // want time.January is a placeholder for error cases
		{"Empty String", "", time.January, true},
		{"Month with only year", "2023", time.January, true},
		{"Month with comma", "June,", time.June, false},
		{"Month with space", " July ", time.July, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := monthNameToNumber(tt.input)
			if (err != nil) != tt.wantErr {
				t.Errorf("monthNameToNumber() error = %v, wantErr %v for input %q", err, tt.wantErr, tt.input)
				return
			}
			if !tt.wantErr && got != tt.want {
				t.Errorf("monthNameToNumber() = %v, want %v for input %q", got, tt.want, tt.input)
			}
		})
	}
}

func TestFilterAppointments_DateFiltering(t *testing.T) {
	// Define a fixed "current date" for predictable testing
	// Let's say today is January 15, 2024 for test purposes
	fixedNow := time.Date(2024, time.January, 15, 0, 0, 0, 0, time.UTC)
	
	// Override time.Now for this test. This is a simple way for this specific case.
	// More complex scenarios might require interfaces and dependency injection for time.
	originalTimeNow := timeNow // Store original time.Now
	timeNow = func() time.Time { return fixedNow }
	defer func() { timeNow = originalTimeNow }() // Restore original time.Now after test

	allAppointments := []Appointment{
		// Past
		{Month: "December 2023", Day: "10", IsAvailable: true}, 
		// Current Month, but past day (relative to fixedNow)
		{Month: "January 2024", Day: "1", IsAvailable: true},  
		// Current Month, future day (relative to fixedNow)
		{Month: "January 2024", Day: "20", IsAvailable: true}, 
		// Next Month
		{Month: "February 2024", Day: "5", IsAvailable: true},  
		// Two Months Ahead
		{Month: "March 2024", Day: "10", IsAvailable: true},   
		// Three Months Ahead
		{Month: "April 2024", Day: "15", IsAvailable: true},    
		// Four Months Ahead (should be out of range for monthsAhead=3)
		{Month: "May 2024", Day: "20", IsAvailable: true},     
		// Explicit year in the past, different from inferred
		{Month: "November 2023", Day: "5", IsAvailable: true},
	}

	tests := []struct {
		name           string
		monthsAhead    int
		expectedCount  int
		expectedAppointments []Appointment // Specific appointments expected, if checking more than just count
	}{
		{
			name:          "0 months ahead (only current month future days)",
			monthsAhead:   0,
			expectedCount: 1, // Jan 20
			expectedAppointments: []Appointment{
				{Month: "January 2024", Day: "20", IsAvailable: true},
			},
		},
		{
			name:          "1 month ahead",
			monthsAhead:   1,
			expectedCount: 2, // Jan 20, Feb 5
			expectedAppointments: []Appointment{
				{Month: "January 2024", Day: "20", IsAvailable: true},
				{Month: "February 2024", Day: "5", IsAvailable: true},
			},
		},
		{
			name:          "3 months ahead",
			monthsAhead:   3,
			expectedCount: 4, // Jan 20, Feb 5, Mar 10, Apr 15
			expectedAppointments: []Appointment{
				{Month: "January 2024", Day: "20", IsAvailable: true},
				{Month: "February 2024", Day: "5", IsAvailable: true},
				{Month: "March 2024", Day: "10", IsAvailable: true},
				{Month: "April 2024", Day: "15", IsAvailable: true},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// For this test, seenAppointments is empty to focus on date filtering
			filtered := filterAppointments(allAppointments, []Appointment{}, tt.monthsAhead)

			if len(filtered) != tt.expectedCount {
				t.Errorf("filterAppointments() count = %d, want %d", len(filtered), tt.expectedCount)
				t.Logf("Filtered appointments: %+v", filtered)
			}
			
			// Optional: Detailed check of which appointments were returned
			if tt.expectedAppointments != nil {
				// Need to ensure order or use a map/set for comparison if order is not guaranteed
				// For now, assuming filterAppointments preserves relative order of input for simplicity
				if !reflect.DeepEqual(filtered, tt.expectedAppointments) {
					t.Errorf("filterAppointments() returned = %+v, want %+v", filtered, tt.expectedAppointments)
				}
			}
		})
	}
}

// timeNow is used to allow mocking of time.Now() in tests.
var timeNow = time.Now

func TestIsNewAppointment(t *testing.T) {
	seenAppointments := []Appointment{
		{Month: "May", Day: "15", Time: "10:00", IsAvailable: true},
		{Month: "June", Day: "20", Time: "14:00", IsAvailable: true},
	}

	tests := []struct {
		name        string
		appointment Appointment
		want        bool
	}{
		{
			name: "New appointment",
			appointment: Appointment{Month: "July", Day: "10", Time: "12:00", IsAvailable: true},
			want:        true,
		},
		{
			name: "Seen appointment (same month and day)",
			appointment: Appointment{Month: "May", Day: "15", Time: "11:00", IsAvailable: false}, // Time and IsAvailable differ, but should still be 'seen'
			want:        false,
		},
		{
			name: "Different day, same month",
			appointment: Appointment{Month: "May", Day: "16", Time: "10:00", IsAvailable: true},
			want:        true,
		},
		{
			name: "Different month, same day",
			appointment: Appointment{Month: "July", Day: "15", Time: "10:00", IsAvailable: true},
			want:        true,
		},
		{
			name: "Empty seen list",
			appointment: Appointment{Month: "May", Day: "15"},
			// want: true, // This will be tested with an empty seenAppointments list directly
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.name == "Empty seen list" { // Special case for this test
				if got := isNewAppointment(tt.appointment, []Appointment{}); got != true {
					t.Errorf("isNewAppointment() with empty seen list = %v, want %v", got, true)
				}
			} else {
				if got := isNewAppointment(tt.appointment, seenAppointments); got != tt.want {
					t.Errorf("isNewAppointment() = %v, want %v for appointment %v", got, tt.want, tt.appointment)
				}
			}
		})
	}
}
