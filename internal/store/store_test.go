package store

import (
	"path/filepath"
	"testing"
)

func TestAddTaskAndSession(t *testing.T) {
	s, err := New(filepath.Join(t.TempDir(), "studyflow.json"))
	if err != nil {
		t.Fatal(err)
	}
	if _, err := s.AddTask("Изучить HTTP в Go", "Go"); err != nil {
		t.Fatal(err)
	}
	if _, err := s.AddSession(45, "Написал первый handler"); err != nil {
		t.Fatal(err)
	}
	snapshot, err := s.Snapshot()
	if err != nil {
		t.Fatal(err)
	}
	if len(snapshot.Tasks) != 1 || snapshot.TotalMinutes != 45 {
		t.Fatalf("unexpected snapshot: %+v", snapshot)
	}
}

func TestRejectsInvalidInput(t *testing.T) {
	s, _ := New(filepath.Join(t.TempDir(), "studyflow.json"))
	if _, err := s.AddTask(" ", ""); err != ErrEmptyTitle {
		t.Fatalf("expected empty title error, got %v", err)
	}
	if _, err := s.AddSession(0, ""); err != ErrInvalidMins {
		t.Fatalf("expected minutes error, got %v", err)
	}
}
