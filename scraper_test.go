package main

import (
	"strings"
	"testing"
	"time"
)

func TestExtractSpaces(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected int
	}{
		{
			name:     "Single space available",
			input:    "1 space available",
			expected: 1,
		},
		{
			name:     "Multiple spaces available",
			input:    "5 spaces available",
			expected: 5,
		},
		{
			name:     "No number in text",
			input:    "No spaces available",
			expected: 0,
		},
		{
			name:     "Empty string",
			input:    "",
			expected: 0,
		},
		{
			name:     "Text with extra whitespace",
			input:    "  3 spaces available  ",
			expected: 3,
		},
		{
			name:     "Singular vs plural",
			input:    "1 spaces available",
			expected: 1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := extractSpaces(tt.input)
			if result != tt.expected {
				t.Errorf("extractSpaces(%q) = %d, want %d", tt.input, result, tt.expected)
			}
		})
	}
}

func TestGenerateDateRange(t *testing.T) {
	// Fixed start date for consistent testing
	startDate := time.Date(2024, 5, 15, 0, 0, 0, 0, time.UTC)

	tests := []struct {
		name     string
		days     int
		expected []string
	}{
		{
			name:     "Single day",
			days:     1,
			expected: []string{"2024-05-15"},
		},
		{
			name:     "Three days",
			days:     3,
			expected: []string{"2024-05-15", "2024-05-16", "2024-05-17"},
		},
		{
			name:     "Zero days",
			days:     0,
			expected: []string{},
		},
		{
			name: "Week range",
			days: 7,
			expected: []string{
				"2024-05-15", "2024-05-16", "2024-05-17", "2024-05-18",
				"2024-05-19", "2024-05-20", "2024-05-21",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := generateDateRange(startDate, tt.days)

			if len(result) != len(tt.expected) {
				t.Errorf("generateDateRange() length = %d, want %d", len(result), len(tt.expected))
				return
			}

			for i, date := range result {
				if date != tt.expected[i] {
					t.Errorf("generateDateRange()[%d] = %q, want %q", i, date, tt.expected[i])
				}
			}
		})
	}
}

func TestParseAppointmentSlots(t *testing.T) {
	tests := []struct {
		name         string
		htmlContent  string
		date         string
		expectedLen  int
		expectedAppt *Appointment // Check first appointment if provided
	}{
		{
			name: "Single appointment slot",
			htmlContent: `<div class="timeslot">
				<div class="timeslot-range">10:00 am - 11:00 am</div>
				<div class="timeslot-time">
					<span class="spots-available">2 spaces available</span>
				</div>
			</div>`,
			date:        "2024-05-15",
			expectedLen: 1,
			expectedAppt: &Appointment{
				Date:        "2024-05-15",
				Time:        "10:00 am – 11:00 am",
				Spaces:      2,
				IsAvailable: true,
			},
		},
		{
			name: "Multiple appointment slots",
			htmlContent: `<div class="timeslot">
				<div class="timeslot-range">10:00 am - 11:00 am</div>
				<div class="timeslot-time">
					<span class="spots-available">1 space available</span>
				</div>
			</div>
			<div class="timeslot">
				<div class="timeslot-range">2:00 pm - 3:00 pm</div>
				<div class="timeslot-time">
					<span class="spots-available">3 spaces available</span>
				</div>
			</div>`,
			date:        "2024-05-15",
			expectedLen: 2,
		},
		{
			name:        "No appointment slots",
			htmlContent: `<div>No appointments available</div>`,
			date:        "2024-05-15",
			expectedLen: 0,
		},
		{
			name: "Invalid time format",
			htmlContent: `<div class="timeslot">
				<div class="timeslot-range">Invalid time format</div>
				<div class="timeslot-time">
					<span class="spots-available">2 spaces available</span>
				</div>
			</div>`,
			date:        "2024-05-15",
			expectedLen: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := parseAppointmentSlots(tt.htmlContent, tt.date)

			if err != nil {
				t.Errorf("parseAppointmentSlots() error = %v, want nil", err)
				return
			}

			if len(result) != tt.expectedLen {
				t.Errorf("parseAppointmentSlots() length = %d, want %d", len(result), tt.expectedLen)
				return
			}

			if tt.expectedAppt != nil && len(result) > 0 {
				appt := result[0]
				if appt.Date != tt.expectedAppt.Date ||
					appt.Time != tt.expectedAppt.Time ||
					appt.Spaces != tt.expectedAppt.Spaces ||
					appt.IsAvailable != tt.expectedAppt.IsAvailable {
					t.Errorf("parseAppointmentSlots()[0] = %+v, want %+v", appt, *tt.expectedAppt)
				}
			}
		})
	}
}

func TestBuildEmailBody(t *testing.T) {
	tests := []struct {
		name               string
		appointments       []Appointment
		expectedSubstrings []string
	}{
		{
			name: "Single appointment",
			appointments: []Appointment{
				{Date: "2024-05-15", Time: "10:00 am – 11:00 am", Spaces: 2, IsAvailable: true},
			},
			expectedSubstrings: []string{
				"New Melanzana appointments found:",
				"2024-05-15 at 10:00 am – 11:00 am (2 spaces available)",
				"Book at: https://melanzana.com/book-an-appointment",
			},
		},
		{
			name: "Multiple appointments",
			appointments: []Appointment{
				{Date: "2024-05-15", Time: "10:00 am – 11:00 am", Spaces: 2, IsAvailable: true},
				{Date: "2024-05-16", Time: "2:00 pm – 3:00 pm", Spaces: 1, IsAvailable: true},
			},
			expectedSubstrings: []string{
				"New Melanzana appointments found:",
				"2024-05-15 at 10:00 am – 11:00 am (2 spaces available)",
				"2024-05-16 at 2:00 pm – 3:00 pm (1 spaces available)",
				"Book at: https://melanzana.com/book-an-appointment",
			},
		},
		{
			name:         "Empty appointments",
			appointments: []Appointment{},
			expectedSubstrings: []string{
				"New Melanzana appointments found:",
				"Book at: https://melanzana.com/book-an-appointment",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := buildEmailBody(tt.appointments)

			for _, substring := range tt.expectedSubstrings {
				if !strings.Contains(result, substring) {
					t.Errorf("buildEmailBody() missing expected substring: %q\nFull result: %q", substring, result)
				}
			}
		})
	}
}
