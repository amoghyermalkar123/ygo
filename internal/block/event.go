package block

import "log/slog"

type EventType string

const (
	Insert    EventType = "insert"
	Delete    EventType = "delete"
	Integrate EventType = "integrate"
	Split     EventType = "split"
	Marker    EventType = "marker"
)

type BlockSnapshot struct {
	ID          ID     `json:"id"`
	Content     string `json:"content"`
	Deleted     bool   `json:"is_deleted"`
	LeftOrigin  ID     `json:"left_origin"`
	RightOrigin ID     `json:"right_origin"`
}

type Event struct {
	Type           EventType                 `json:"type"`
	StateVector    map[int64]uint64          `json:"state_vector,omitempty"`
	BlocksByClient map[int64][]BlockSnapshot `json:"blocks,omitempty"`
	Points         []slog.Attr               `json:"points,omitempty"`
}
