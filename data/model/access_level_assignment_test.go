package model

import "testing"

func TestNewAccessLevelAssignment_shouldLinkAccessLevelAndBadge(t *testing.T) {
	al := &AccessLevel{ID: 10, Name: "Main Entrance"}
	b := &Badge{ID: 20, Key: 200}
	a, err := NewAccessLevelAssignment(al, b)
	if err != nil {
		t.Fatal(err)
	}
	if a.AccessLevel.ID != 10 {
		t.Errorf("expected AccessLevel.ID 10, got %d", a.AccessLevel.ID)
	}
	if a.Badge.ID != 20 {
		t.Errorf("expected Badge.ID 20, got %d", a.Badge.ID)
	}
}

func TestNewAccessLevelAssignment_shouldErrorWhenAccessLevelNil(t *testing.T) {
	_, err := NewAccessLevelAssignment(nil, &Badge{ID: 20, Key: 200})
	if err != ErrAssignmentNilAccessLevel {
		t.Errorf("expected ErrAssignmentNilAccessLevel, got %v", err)
	}
}

func TestNewAccessLevelAssignment_shouldErrorWhenBadgeNil(t *testing.T) {
	_, err := NewAccessLevelAssignment(&AccessLevel{ID: 10, Name: "Main"}, nil)
	if err != ErrAssignmentNilBadge {
		t.Errorf("expected ErrAssignmentNilBadge, got %v", err)
	}
}

func TestNewAccessLevelAssignmentFromJSON_shouldErrorWhenCacheNil(t *testing.T) {
	_, err := NewAccessLevelAssignmentFromJSON(map[string]any{}, nil)
	if err != ErrAssignmentNilCache {
		t.Errorf("expected ErrAssignmentNilCache, got %v", err)
	}
}
