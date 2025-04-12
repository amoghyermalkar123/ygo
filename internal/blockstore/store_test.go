package blockstore

import (
	"testing"
)

func TestInsertAtBeginning(t *testing.T) {
	store := NewStore()
	err := store.InsertText(0, "Hello", 1)
	if err != nil {
		t.Fatalf("insert failed: %v", err)
	}
	if got := store.Content(); got != "Hello" {
		t.Errorf("expected Hello, got %q", got)
	}
}

func TestInsertAtEnd(t *testing.T) {
	store := NewStore()
	_ = store.InsertText(0, "Hi", 1)
	_ = store.InsertText(2, " there", 1)

	if got := store.Content(); got != "Hi there" {
		t.Errorf("expected 'Hi there', got %q", got)
	}
}

func TestInsertInMiddle(t *testing.T) {
	store := NewStore()
	_ = store.InsertText(0, "A", 1)
	_ = store.InsertText(1, "B", 1)
	_ = store.InsertText(1, "X", 1) // Insert in middle → AXB

	if got := store.Content(); got != "AXB" {
		t.Errorf("expected AXB, got %q", got)
	}
}

func TestInsertTriggersSplit(t *testing.T) {
	store := NewStore()
	_ = store.InsertText(0, "World", 1)
	_ = store.InsertText(2, "X", 1) // Wo|rld → insert "X" at pos 2

	if got := store.Content(); got != "WoXrld" {
		t.Errorf("expected WoXrld, got %q", got)
	}
}

func TestDeleteSingleBlock(t *testing.T) {
	store := NewStore()
	_ = store.InsertText(0, "A", 1)

	err := store.DeleteText(0, 1)
	if err != nil {
		t.Fatalf("delete failed: %v", err)
	}
	if got := store.Content(); got != "" {
		t.Errorf("expected empty string, got %q", got)
	}
}

func TestDeleteMiddleOfBlock(t *testing.T) {
	store := NewStore()
	_ = store.InsertText(0, "Hello", 1)

	err := store.DeleteText(1, 3) // Remove "ell"
	if err != nil {
		t.Fatalf("delete failed: %v", err)
	}
	if got := store.Content(); got != "Ho" {
		t.Errorf("expected Ho, got %q", got)
	}
}

func TestDeleteMultipleBlocks(t *testing.T) {
	store := NewStore()
	_ = store.InsertText(0, "Hi", 1)
	_ = store.InsertText(2, " there", 1) // "Hi there"

	err := store.DeleteText(1, 5) // Remove "i the"
	if err != nil {
		t.Fatalf("delete failed: %v", err)
	}
	if got := store.Content(); got != "Hr" {
		t.Errorf("expected Hr, got %q", got)
	}
}

func TestDeleteOutOfBounds(t *testing.T) {
	store := NewStore()
	_ = store.InsertText(0, "Yo", 1)

	err := store.DeleteText(3, 1)
	if err == nil {
		t.Fatal("expected error on out-of-bounds delete, got none")
	}
}
