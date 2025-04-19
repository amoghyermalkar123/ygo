package blockstore

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestInsertAtBeginning(t *testing.T) {
	store := NewStore()
	err := store.Insert(0, "Hello")
	if err != nil {
		t.Fatalf("insert failed: %v", err)
	}
	if got := store.Content(); got != "Hello" {
		t.Errorf("expected Hello, got %q", got)
	}
}

func TestInsertAtEnd(t *testing.T) {
	store := NewStore()
	err := store.Insert(0, "Hi")
	assert.NoError(t, err)
	err = store.Insert(2, " there")
	assert.NoError(t, err)

	if got := store.Content(); got != "Hi there" {
		t.Errorf("expected 'Hi there', got %q", got)
	}
}

func TestInsertInMiddle(t *testing.T) {
	store := NewStore(WithDebugModeEnabled())
	defer store.logger.Flush()

	_ = store.Insert(0, "A")
	_ = store.Insert(1, "B")
	_ = store.Insert(1, "X") // Insert in middle →)

	if got := store.Content(); got != "AXB" {
		t.Errorf("expected AXB, got %q", got)
	}
}

func TestInsertTriggersSplit(t *testing.T) {
	store := NewStore()
	_ = store.Insert(0, "World")
	_ = store.Insert(2, "X") // Wo|rld → insert "X" at p)

	if got := store.Content(); got != "WoXrld" {
		t.Errorf("expected WoXrld, got %q", got)
	}
}

func TestDeleteSingleBlock(t *testing.T) {
	store := NewStore()
	_ = store.Insert(0, "A")

	err := store.Delete(0, 1)
	if err != nil {
		t.Fatalf("delete failed: %v", err)
	}
	if got := store.Content(); got != "" {
		t.Errorf("expected empty string, got %q", got)
	}
}

func TestDeleteMiddleOfBlock(t *testing.T) {
	store := NewStore()
	_ = store.Insert(0, "Hello")

	err := store.Delete(1, 3) // Remove "ell"
	if err != nil {
		t.Fatalf("delete failed: %v", err)
	}
	if got := store.Content(); got != "Ho" {
		t.Errorf("expected Ho, got %q", got)
	}
}

func TestDeleteMultipleBlocks(t *testing.T) {
	store := NewStore(WithDebugModeEnabled())
	_ = store.Insert(0, "Hi")
	_ = store.Insert(2, " there") // "Hi th)

	err := store.Delete(1, 5) // Remove "i the"
	if err != nil {
		t.Fatalf("delete failed: %v", err)
	}
	if got := store.Content(); got != "Hre" {
		t.Errorf("expected Hr, got %q", got)
	}
}

func TestDeleteOutOfBounds(t *testing.T) {
	store := NewStore()
	_ = store.Insert(0, "Yo")

	err := store.Delete(3, 3)

	if err == nil {
		t.Fatal("expected error on out-of-bounds delete, got %w", err)
	}
}
