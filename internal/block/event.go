package block

import "time"

type EventType string

const (
	Insert    EventType = "insert"
	Delete    EventType = "delete"
	Integrate EventType = "integrate"
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
	Timestamp      time.Time                 `json:"timestamp"`
	Length         int                       `json:"length"`
	StateVector    map[int64]uint64          `json:"state_vector"`
	BlocksByClient map[int64][]BlockSnapshot `json:"blocks"`
}
