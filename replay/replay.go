package replay

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"
	"ygo/internal/block"
	replayui "ygo/replay/ui"
)

func StartReplayEngine(mux *http.ServeMux) {
	events, err := LoadEvents()
	if err != nil {
		panic(fmt.Errorf("load events: %w", err))
	}

	playback := NewPlayback(events)

	mux.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.Dir("replay/web/static"))))

	mux.HandleFunc("/replay", func(w http.ResponseWriter, r *http.Request) {
		replayui.ReplayPage().Render(r.Context(), w)
	})

	mux.HandleFunc("/replay/next", func(w http.ResponseWriter, r *http.Request) {
		evt := playback.PlayNext()
		replayui.FullReplayPage(evt).Render(r.Context(), w)
	})

	mux.HandleFunc("/replay/reset", func(w http.ResponseWriter, r *http.Request) {
		playback.Reset()
		replayui.Empty().Render(r.Context(), w)
	})
}

func LoadEvents() ([]block.Event, error) {
	dir := "tmp/" // Change to your directory

	files, err := os.ReadDir(dir)
	if err != nil {
		panic(err)
	}

	var latestTime time.Time
	var latestFile string

	for _, file := range files {
		if file.IsDir() {
			continue
		}

		name := file.Name()

		if !strings.HasPrefix(name, "replay_") {
			continue
		}

		timePart := strings.TrimPrefix(name, "replay_")
		timePart = strings.TrimSuffix(timePart, filepath.Ext(timePart))

		t, err := time.Parse(time.RFC3339, timePart)
		if err != nil {
			fmt.Printf("Skipping %s: %v\n", name, err)
			continue
		}

		if t.After(latestTime) {
			latestTime = t
			latestFile = name
		}
	}

	if latestFile == "" {
		return nil, fmt.Errorf("no valid replay files found")
	}

	data, err := os.ReadFile(fmt.Sprintf("%s%s", dir, latestFile))
	if err != nil {
		return nil, fmt.Errorf("read file: %w", err)
	}

	var events []block.Event
	if err := json.Unmarshal(data, &events); err != nil {
		return nil, fmt.Errorf("unmarshal: %w", err)
	}

	return events, nil
}
