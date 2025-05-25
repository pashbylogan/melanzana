package main

import (
	"encoding/json"
	"os"
	"path/filepath"
	"reflect"
	"testing"
)

func TestLoadAndSaveSeenAppointments(t *testing.T) {
	// Create a temporary directory for test files
	tempDir, err := os.MkdirTemp("", "storage_test_")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir) // Clean up the temp directory

	testFilePath := filepath.Join(tempDir, "test_seen_appointments.json")

	t.Run("SaveAndLoadSuccessfully", func(t *testing.T) {
		originalAppointments := []Appointment{
			{Date: "2024-08-10", Time: "10:00 am – 11:00 am", Spaces: 2, IsAvailable: true},
			{Date: "2024-09-22", Time: "3:00 pm – 4:00 pm", Spaces: 1, IsAvailable: true},
		}

		err := saveSeenAppointments(originalAppointments, testFilePath)
		if err != nil {
			t.Fatalf("saveSeenAppointments() failed: %v", err)
		}

		loadedAppointments, err := loadSeenAppointments(testFilePath)
		if err != nil {
			t.Fatalf("loadSeenAppointments() failed: %v", err)
		}

		if !reflect.DeepEqual(originalAppointments, loadedAppointments) {
			t.Errorf("loadSeenAppointments() got = %v, want %v", loadedAppointments, originalAppointments)
		}
	})

	t.Run("LoadNonExistentFile", func(t *testing.T) {
		nonExistentFilePath := filepath.Join(tempDir, "non_existent.json")
		loaded, err := loadSeenAppointments(nonExistentFilePath)
		if err != nil {
			t.Errorf("loadSeenAppointments() with non-existent file error = %v, want nil", err)
		}
		if len(loaded) != 0 {
			t.Errorf("loadSeenAppointments() with non-existent file got %d appointments, want 0", len(loaded))
		}
	})

	t.Run("LoadEmptyFile", func(t *testing.T) {
		emptyFilePath := filepath.Join(tempDir, "empty.json")
		f, err := os.Create(emptyFilePath)
		if err != nil {
			t.Fatalf("Failed to create empty file: %v", err)
		}
		f.Close() // Ensure file is created and empty

		loaded, err := loadSeenAppointments(emptyFilePath)
		if err != nil {
			t.Errorf("loadSeenAppointments() with empty file error = %v, want nil", err)
		}
		if len(loaded) != 0 {
			t.Errorf("loadSeenAppointments() with empty file got %d appointments, want 0", len(loaded))
		}
	})

	t.Run("LoadMalformedJSON", func(t *testing.T) {
		malformedFilePath := filepath.Join(tempDir, "malformed.json")
		err := os.WriteFile(malformedFilePath, []byte("[{malformed json}}"), 0644)
		if err != nil {
			t.Fatalf("Failed to write malformed JSON file: %v", err)
		}

		_, err = loadSeenAppointments(malformedFilePath)
		if err == nil {
			t.Errorf("loadSeenAppointments() with malformed JSON error = nil, want error")
		}
	})

	t.Run("SaveEmptySlice", func(t *testing.T) {
		emptyAppointments := []Appointment{}
		emptySliceFilePath := filepath.Join(tempDir, "empty_slice_saved.json")

		err := saveSeenAppointments(emptyAppointments, emptySliceFilePath)
		if err != nil {
			t.Fatalf("saveSeenAppointments() with empty slice failed: %v", err)
		}

		loaded, err := loadSeenAppointments(emptySliceFilePath)
		if err != nil {
			t.Fatalf("loadSeenAppointments() after saving empty slice failed: %v", err)
		}
		if len(loaded) != 0 {
			t.Errorf("loadSeenAppointments() after saving empty slice got %d, want 0", len(loaded))
		}

		// Verify content is valid empty JSON array
		content, readErr := os.ReadFile(emptySliceFilePath)
		if readErr != nil {
			t.Fatalf("Failed to read file after saving empty slice: %v", readErr)
		}

		var checkEmpty []Appointment
		if unmarshalErr := json.Unmarshal(content, &checkEmpty); unmarshalErr != nil || len(checkEmpty) != 0 {
			t.Errorf("File content after saving empty slice is not a valid empty JSON array. Got: %s", string(content))
		}
	})

	t.Run("SaveAndLoadLargeDataset", func(t *testing.T) {
		// Test with a larger dataset to ensure the system handles it well
		var largeAppointments []Appointment
		for i := 0; i < 100; i++ {
			largeAppointments = append(largeAppointments, Appointment{
				Date:        "2024-05-15",
				Time:        "10:00 am – 11:00 am",
				Spaces:      i%5 + 1, // Varies from 1-5
				IsAvailable: true,
			})
		}

		largeFilePath := filepath.Join(tempDir, "large_dataset.json")

		err := saveSeenAppointments(largeAppointments, largeFilePath)
		if err != nil {
			t.Fatalf("saveSeenAppointments() with large dataset failed: %v", err)
		}

		loaded, err := loadSeenAppointments(largeFilePath)
		if err != nil {
			t.Fatalf("loadSeenAppointments() with large dataset failed: %v", err)
		}

		if len(loaded) != len(largeAppointments) {
			t.Errorf("loadSeenAppointments() large dataset length = %d, want %d", len(loaded), len(largeAppointments))
		}

		if !reflect.DeepEqual(loaded, largeAppointments) {
			t.Errorf("Large dataset not preserved through save/load cycle")
		}
	})
}
