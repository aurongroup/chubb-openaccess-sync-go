package model

import "testing"

func TestNewCardholder_shouldSetAllFields(t *testing.T) {
	ch, err := NewCardholder(10, "1234", "Bob", "Brown")
	if err != nil {
		t.Fatal(err)
	}
	if ch.ID != 10 {
		t.Errorf("ID: expected 10, got %d", ch.ID)
	}
	if ch.SSNO != "1234" {
		t.Errorf("SSNO: expected %q, got %q", "1234", ch.SSNO)
	}
	if ch.FirstName != "Bob" {
		t.Errorf("FirstName: expected %q, got %q", "Bob", ch.FirstName)
	}
	if ch.LastName != "Brown" {
		t.Errorf("LastName: expected %q, got %q", "Brown", ch.LastName)
	}
}

func TestNewCardholder_shouldAllowZeroIDAndEmptySsno(t *testing.T) {
	ch, err := NewCardholder(0, "", "Jane", "Doe")
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if ch.ID != 0 {
		t.Errorf("ID: expected 0, got %d", ch.ID)
	}
	if ch.SSNO != "" {
		t.Errorf("SSNO: expected empty, got %q", ch.SSNO)
	}
}

func TestNewCardholder_shouldErrorWhenLastNameMissing(t *testing.T) {
	_, err := NewCardholder(5, "9999", "Alice", "")
	if err != ErrCardholderMissingLastName {
		t.Errorf("expected ErrCardholderMissingLastName, got %v", err)
	}
}

func TestNewCardholderFromJSON_shouldParseAllFields(t *testing.T) {
	ch, err := NewCardholderFromJSON(map[string]any{
		"ID":        float64(10),
		"SSNO":      "1234",
		"FIRSTNAME": "Bob",
		"LASTNAME":  "Brown",
	})
	if err != nil {
		t.Fatal(err)
	}
	if ch.ID != 10 {
		t.Errorf("ID: expected 10, got %d", ch.ID)
	}
	if ch.SSNO != "1234" {
		t.Errorf("SSNO: expected %q, got %q", "1234", ch.SSNO)
	}
	if ch.FirstName != "Bob" {
		t.Errorf("FirstName: expected %q, got %q", "Bob", ch.FirstName)
	}
	if ch.LastName != "Brown" {
		t.Errorf("LastName: expected %q, got %q", "Brown", ch.LastName)
	}
}

func TestNewCardholderFromJSON_shouldTolerateMissingOptionalFields(t *testing.T) {
	ch, err := NewCardholderFromJSON(map[string]any{"LASTNAME": "Smith"})
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if ch.ID != 0 {
		t.Errorf("ID: expected 0, got %d", ch.ID)
	}
	if ch.SSNO != "" {
		t.Errorf("SSNO: expected empty, got %q", ch.SSNO)
	}
	if ch.FirstName != "" {
		t.Errorf("FirstName: expected empty, got %q", ch.FirstName)
	}
}

func TestNewCardholderFromJSON_shouldErrorWhenLastNameMissing(t *testing.T) {
	_, err := NewCardholderFromJSON(map[string]any{"ID": float64(5), "SSNO": "9999"})
	if err != ErrCardholderMissingLastName {
		t.Errorf("expected ErrCardholderMissingLastName, got %v", err)
	}
}
