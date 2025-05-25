package main

import (
	"reflect"
	"testing"
)

func TestFilterNewAppointments(t *testing.T) {
	tests := []struct {
		name             string
		appointments     []Appointment
		seenAppointments []Appointment
		expectedNew      []Appointment
	}{
		{
			name: "No previous appointments - all are new",
			appointments: []Appointment{
				{Date: "2024-05-15", Time: "10:00 am – 11:00 am", Spaces: 2, IsAvailable: true},
				{Date: "2024-05-16", Time: "14:00 pm – 15:00 pm", Spaces: 1, IsAvailable: true},
			},
			seenAppointments: []Appointment{},
			expectedNew: []Appointment{
				{Date: "2024-05-15", Time: "10:00 am – 11:00 am", Spaces: 2, IsAvailable: true},
				{Date: "2024-05-16", Time: "14:00 pm – 15:00 pm", Spaces: 1, IsAvailable: true},
			},
		},
		{
			name: "Some appointments already seen",
			appointments: []Appointment{
				{Date: "2024-05-15", Time: "10:00 am – 11:00 am", Spaces: 2, IsAvailable: true},
				{Date: "2024-05-16", Time: "14:00 pm – 15:00 pm", Spaces: 1, IsAvailable: true},
				{Date: "2024-05-17", Time: "09:00 am – 10:00 am", Spaces: 3, IsAvailable: true},
			},
			seenAppointments: []Appointment{
				{Date: "2024-05-15", Time: "10:00 am – 11:00 am", Spaces: 2, IsAvailable: true},
			},
			expectedNew: []Appointment{
				{Date: "2024-05-16", Time: "14:00 pm – 15:00 pm", Spaces: 1, IsAvailable: true},
				{Date: "2024-05-17", Time: "09:00 am – 10:00 am", Spaces: 3, IsAvailable: true},
			},
		},
		{
			name: "All appointments already seen",
			appointments: []Appointment{
				{Date: "2024-05-15", Time: "10:00 am – 11:00 am", Spaces: 2, IsAvailable: true},
				{Date: "2024-05-16", Time: "14:00 pm – 15:00 pm", Spaces: 1, IsAvailable: true},
			},
			seenAppointments: []Appointment{
				{Date: "2024-05-15", Time: "10:00 am – 11:00 am", Spaces: 2, IsAvailable: true},
				{Date: "2024-05-16", Time: "14:00 pm – 15:00 pm", Spaces: 1, IsAvailable: true},
			},
			expectedNew: []Appointment{},
		},
		{
			name: "Same date and time but different spaces count should not be considered new",
			appointments: []Appointment{
				{Date: "2024-05-15", Time: "10:00 am – 11:00 am", Spaces: 3, IsAvailable: true},
			},
			seenAppointments: []Appointment{
				{Date: "2024-05-15", Time: "10:00 am – 11:00 am", Spaces: 2, IsAvailable: true},
			},
			expectedNew: []Appointment{},
		},
		{
			name:         "Empty appointments list",
			appointments: []Appointment{},
			seenAppointments: []Appointment{
				{Date: "2024-05-15", Time: "10:00 am – 11:00 am", Spaces: 2, IsAvailable: true},
			},
			expectedNew: []Appointment{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := filterNewAppointments(tt.appointments, tt.seenAppointments)

			// Handle comparison of empty slices properly
			if len(result) != len(tt.expectedNew) {
				t.Errorf("filterNewAppointments() length = %d, want %d", len(result), len(tt.expectedNew))
				return
			}

			if len(result) == 0 && len(tt.expectedNew) == 0 {
				return // Both are empty, test passes
			}

			if !reflect.DeepEqual(result, tt.expectedNew) {
				t.Errorf("filterNewAppointments() = %v, want %v", result, tt.expectedNew)
			}
		})
	}
}
