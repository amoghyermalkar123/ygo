package replay

import (
	"sync"
	"ygo/internal/block"
)

type Logger struct {
	mu        sync.Mutex
	Events    []block.Event
	debugMode bool
}

func NewLogger(debugMode bool) *Logger {
	return &Logger{
		Events:    make([]block.Event, 0),
		debugMode: debugMode,
	}
}

// Capture stores the event in the logger if debug mode is enabled.
// It's a no-op if debug mode is disabled.
func (l *Logger) Capture(store map[int64][]block.BlockSnapshot, sv map[int64]uint64, typ block.EventType) {
	if !l.debugMode {
		return
	}

	l.mu.Lock()
	defer l.mu.Unlock()

	event := block.Event{
		Type:           typ,
		StateVector:    sv,
		BlocksByClient: store,
	}

	l.Events = append(l.Events, event)
}
