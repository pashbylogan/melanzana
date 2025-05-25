package main

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
)

// loadSeenAppointments reads appointments from the JSON file specified by dataFilePath.
func loadSeenAppointments(dataFilePath string) ([]Appointment, error) {
	data, err := os.ReadFile(dataFilePath)
	if err != nil {
		if os.IsNotExist(err) {
			log.Printf("File %s does not exist. Returning empty list.", dataFilePath)
			return []Appointment{}, nil // No error if file simply doesn't exist
		}
		return nil, fmt.Errorf("failed to read %s: %w", dataFilePath, err)
	}

	if len(data) == 0 { // Handle empty file case
		log.Printf("File %s is empty. Returning empty list.", dataFilePath)
		return []Appointment{}, nil
	}

	var appointments []Appointment
	err = json.Unmarshal(data, &appointments)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal appointments from %s: %w", dataFilePath, err)
	}
	return appointments, nil
}

// saveSeenAppointments writes appointments to the JSON file specified by dataFilePath.
func saveSeenAppointments(appointments []Appointment, dataFilePath string) error {
	data, err := json.MarshalIndent(appointments, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal appointments to JSON: %w", err)
	}

	err = os.WriteFile(dataFilePath, data, 0644) // 0644 are standard file permissions
	if err != nil {
		return fmt.Errorf("failed to write appointments to %s: %w", dataFilePath, err)
	}
	return nil
}
