package marker

import (
	"testing"
	"time"

	"github.com/amoghyermalkar123/ygo/internal/block"
)

func TestAddAndFindMarker(t *testing.T) {
	ms := NewSystem()

	b := block.NewBlock(block.ID{Clock: 1, Client: 1}, "Hello")
	ms.Add(b, 5)

	m, err := ms.FindMarker(5)
	if err != nil {
		t.Fatalf("expected to find marker, got error: %v", err)
	}
	if m.Block != b {
		t.Fatalf("expected block %v, got %v", b, m.Block)
	}
	if m.Pos != 5 {
		t.Fatalf("expected pos 5, got %d", m.Pos)
	}
}

func TestUpdateMarkers_Add(t *testing.T) {
	ms := NewSystem()

	b := block.NewBlock(block.ID{Clock: 1, Client: 1}, "Hello")
	ms.Add(b, 5)

	time.Sleep(1 * time.Millisecond)
	ms.UpdateMarkers(4, 2, OpAdd)

	if ms.Markers[0].Pos != 7 {
		t.Errorf("expected pos 7, got %d", ms.Markers[0].Pos)
	}
}

func TestUpdateMarkers_Del(t *testing.T) {
	ms := NewSystem()

	b := block.NewBlock(block.ID{Clock: 1, Client: 1}, "Hello")
	ms.Add(b, 5)

	ms.UpdateMarkers(4, 2, OpDel)
	if ms.Markers[0].Pos != 3 {
		t.Errorf("expected pos 3, got %d", ms.Markers[0].Pos)
	}
}

func TestDeleteMarkerAt(t *testing.T) {
	ms := NewSystem()
	b := block.NewBlock(block.ID{Clock: 1, Client: 1}, "Hello")
	ms.Add(b, 5)

	ms.DeleteMarkerAt(5)

	if len(ms.Markers) != 0 {
		t.Fatalf("expected marker list to be empty")
	}
}

func TestGetBlockPositionByID(t *testing.T) {
	ms := NewSystem()

	b := block.NewBlock(block.ID{Clock: 1, Client: 1}, "hello")
	ms.Add(b, 10)

	pos, err := ms.GetBlockPositionByID(&b.ID)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if pos != 10 {
		t.Fatalf("expected pos 10, got %d", pos)
	}
}

func TestGetBlockPositionByID_Error(t *testing.T) {
	ms := NewSystem()
	id := block.ID{Clock: 99, Client: 42}

	_, err := ms.GetBlockPositionByID(&id)
	if err == nil {
		t.Fatalf("expected error for unknown ID")
	}
}

func TestGetBlockPositionByClock(t *testing.T) {
	ms := NewSystem()
	b := block.NewBlock(block.ID{Clock: 7, Client: 99}, "yo")
	ms.Add(b, 3)

	pos, err := ms.GetBlockPositionByClock(b.ID)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if pos != 3 {
		t.Fatalf("expected position 3, got %d", pos)
	}
}

func TestDestroyMarkers(t *testing.T) {
	ms := NewSystem()
	b := block.NewBlock(block.ID{Clock: 1, Client: 1}, "data")
	ms.Add(b, 10)

	ms.DestroyMarkers()

	if len(ms.Markers) != 0 {
		t.Fatalf("expected all markers to be removed")
	}
}
