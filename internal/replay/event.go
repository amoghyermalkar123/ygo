package replay

import (
	"encoding/json"
	"fmt"
	"os"
	"sync"
	"time"
	"ygo/internal/block"
)

type Logger struct {
	mu        sync.Mutex
	debugMode bool
	events    []block.Event
}

func NewLogger(debugMode bool) *Logger {
	return &Logger{
		debugMode: debugMode,
		events:    make([]block.Event, 0),
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

	l.events = append(l.events, event)
}

func (l *Logger) Flush() error {
	l.mu.Lock()
	defer l.mu.Unlock()

	file, err := os.Create(fmt.Sprintf("../../tmp/replay_%s", time.Now().Format(time.RFC3339)))
	if err != nil {
		return err
	}
	defer file.Close()

	encoder := json.NewEncoder(file)
	encoder.SetIndent("", "  ")
	return encoder.Encode(l.events)
}
