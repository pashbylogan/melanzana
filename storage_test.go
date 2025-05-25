package main

import (
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

	// 1. Test successful save and load
	t.Run("SaveAndLoadSuccessfully", func(t *testing.T) {
		originalAppointments := []Appointment{
			{Month: "August", Day: "10", Time: "10:00", IsAvailable: true},
			{Month: "September", Day: "22", Time: "15:00", IsAvailable: false},
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

	// 2. Test loading from a non-existent file
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

	// 3. Test loading from an empty file
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

	// 4. Test loading from a malformed JSON file
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

	// 5. Test saving an empty slice
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

		// Verify content is "[]" or "[]\n"
		content, readErr := os.ReadFile(emptySliceFilePath)
		if readErr != nil {
			t.Fatalf("Failed to read file after saving empty slice: %v", readErr)
		}
		trimmedContent := string(content)
		// Allow for potential newline character from MarshalIndent
		if trimmedContent != "[]" && trimmedContent != "[]\n" && trimmedContent != "[\n\n]" { // MarshalIndent might add newlines
			// More robust check for empty array after potential pretty printing
			var checkEmpty []Appointment
			if unmarshalErr := json.Unmarshal(content, &checkEmpty); unmarshalErr != nil || len(checkEmpty) != 0 {
				t.Errorf("File content after saving empty slice is not an empty JSON array. Got: %s", string(content))
			}
		}
	})
}
