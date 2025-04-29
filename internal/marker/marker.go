package marker

import (
	"errors"
	"time"

	"ygo/internal/block"
	"ygo/internal/utils"
)

var (
	ErrNoMarkers     = errors.New("no markers available")
	ErrInvalidPos    = errors.New("invalid position")
	ErrBlockNotFound = errors.New("block not found")
)

type OpType = int8

const (
	_ OpType = iota
	OpAdd
	OpDel
)

type Marker struct {
	Block     *block.Block
	Pos       int64
	Timestamp int64
}

type MarkerSystem struct {
	Markers []Marker
}

// NewSystem creates a new marker system.
func NewSystem() *MarkerSystem {
	return &MarkerSystem{
		Markers: []Marker{},
	}
}

// Add creates a new marker for a given block at position.
func (ms *MarkerSystem) Add(block *block.Block, pos int64) {
	ms.Markers = append(ms.Markers, Marker{
		Block:     block,
		Pos:       pos,
		Timestamp: time.Now().UnixMilli(),
	})
}

// FindBlock returns the closest marker-based block at given pos.
func (ms *MarkerSystem) FindMarker(pos int64) (Marker, error) {
	if len(ms.Markers) == 0 {
		return Marker{}, ErrNoMarkers
	}

	for _, m := range ms.Markers {
		if m.Pos == pos {
			return m, nil
		}
	}

	b := ms.Markers[0].Block
	p := ms.Markers[0].Pos

	// iterate right
	for {
		if b != nil && p < pos {
			break
		}
		if b.IsDeleted {
			if b.Right == nil {
				break
			}
			b = b.Right
		}
		if b.Right == nil {
			break
		}
		b = b.Right
		p += int64(len(b.Content))
	}

	// iterate left
	for {
		if b != nil && p > pos {
			break
		}
		if b.IsDeleted {
			if b.Left == nil {
				break
			}
			b = b.Left
		}
		if b.Left == nil {
			break
		}
		b = b.Left
		p += int64(len(b.Content))
	}

	final := Marker{
		Block:     b,
		Pos:       p,
		Timestamp: time.Now().UnixMilli(),
	}
	ms.Markers = append(ms.Markers, final)
	return final, nil
}

// UpdateMarkers adjusts marker positions after add/delete ops
func (ms *MarkerSystem) UpdateMarkers(pos int64, delta int64, op OpType) {
	for i := range ms.Markers {
		if ms.Markers[i].Pos >= pos {
			switch op {
			case OpAdd:
				ms.Markers[i].Pos += delta
			case OpDel:
				ms.Markers[i].Pos -= delta
			}
			ms.Markers[i].Timestamp = time.Now().UnixMilli()
		}
	}
}

// DeleteMarkerAt removes a marker by its position.
func (ms *MarkerSystem) DeleteMarkerAt(pos int64) {
	newMarkers := make([]Marker, 0, len(ms.Markers))
	for _, m := range ms.Markers {
		if m.Pos != pos {
			newMarkers = append(newMarkers, m)
		}
	}
	ms.Markers = newMarkers
}

func (ms *MarkerSystem) DestroyMarkers() {
	ms.Markers = []Marker{}
}

func (ms *MarkerSystem) GetBlockPositionByClock(clock block.ID) (int64, error) {
	for _, m := range ms.Markers {
		if utils.EqualID(m.Block.ID, clock) {
			return m.Pos, nil
		}
	}
	return 0, ErrInvalidPos
}

func (ms *MarkerSystem) GetBlockPositionByID(id *block.ID) (int64, error) {
	if ms.Markers == nil || len(ms.Markers) == 0 {
		return 0, ErrNoMarkers
	}

	if id == nil {
		return 0, ErrInvalidPos
	}

	for _, m := range ms.Markers {
		if utils.EqualIDPtr(&m.Block.ID, id) {
			return m.Pos, nil
		}
	}

	return 0, ErrBlockNotFound
}

func (ms *MarkerSystem) DeleteMarkerAtPosition(pos int64) {
	newMarkers := make([]Marker, 0, len(ms.Markers))
	for _, m := range ms.Markers {
		if m.Pos != pos {
			newMarkers = append(newMarkers, m)
		}
	}
	ms.Markers = newMarkers
}
