package replay

import (
	"encoding/json"
	"fmt"
	"log/slog"
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

func (l *Logger) LogIntegrate(msg string, params ...slog.Attr) {
	l.logEvent(block.Integrate, params...)
}

func (l *Logger) LogInsert(msg string, params ...slog.Attr) {
	l.logEvent(block.Insert, params...)

}
func (l *Logger) LogDelete(msg string, params ...slog.Attr) {
	l.logEvent(block.Delete, params...)

}
func (l *Logger) LogSplit(msg string, params ...slog.Attr) {
	l.logEvent(block.Split, params...)

}
func (l *Logger) LogMarker(msg string, params ...slog.Attr) {
	l.logEvent(block.Marker, params...)
}

func (l *Logger) logEvent(op block.EventType, params ...slog.Attr) {
	if !l.debugMode {
		return
	}

	l.mu.Lock()
	defer l.mu.Unlock()

	event := block.Event{
		Type:   op,
		Points: params,
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
