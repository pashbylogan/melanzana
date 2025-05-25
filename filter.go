package main

import "log"

// filterNewAppointments returns appointments that haven't been seen before.
func filterNewAppointments(appointments, seenAppointments []Appointment) []Appointment {
	if len(seenAppointments) == 0 {
		log.Printf("No previous appointments found, all %d appointments are new", len(appointments))
		return appointments
	}

	// Create a set of seen appointments for O(1) lookup
	seenSet := make(map[string]bool)
	for _, seen := range seenAppointments {
		key := seen.Date + "|" + seen.Time
		seenSet[key] = true
	}

	var newAppointments []Appointment
	for _, appt := range appointments {
		key := appt.Date + "|" + appt.Time
		if !seenSet[key] {
			newAppointments = append(newAppointments, appt)
		}
	}

	log.Printf("Filtered %d new appointments from %d total", len(newAppointments), len(appointments))
	return newAppointments
}
