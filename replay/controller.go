package replay

import (
	"net/http"
	"sync"
	"ygo/internal/block"
	"ygo/internal/replay"
	replayui "ygo/replay/ui"
)

var (
	events       []block.Event // This should be populated from logger.Events
	replayEngine *ReplayEngine
	once         sync.Once
)

func loadEventsFromLogger(logger *replay.Logger) {
	once.Do(func() {
		events = logger.Events
		replayEngine = NewReplayEngine(events)
	})
}

func RegisterHandlers(mux *http.ServeMux, logger *replay.Logger) {
	loadEventsFromLogger(logger)

	mux.HandleFunc("/replay", func(w http.ResponseWriter, r *http.Request) {
		replayui.ReplayPage().Render(r.Context(), w)
	})

	mux.HandleFunc("/replay/next", func(w http.ResponseWriter, r *http.Request) {
		evt := replayEngine.PlayNext()
		replayui.EventItem(evt).Render(r.Context(), w)
	})

	mux.HandleFunc("/replay/reset", func(w http.ResponseWriter, r *http.Request) {
		replayEngine.Reset()
		replayui.Empty().Render(r.Context(), w)
	})
}
