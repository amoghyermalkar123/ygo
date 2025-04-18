package replay

import "ygo/internal/block"

type ReplayEngine struct {
	doc   []block.Event
	index int
}

func NewReplayEngine(events []block.Event) *ReplayEngine {
	return &ReplayEngine{doc: events}
}

func (r *ReplayEngine) PlayNext() block.Event {
	if r.index >= len(r.doc) {
		return block.Event{Type: "end"}
	}
	evt := r.doc[r.index]
	r.index++
	return evt
}

func (r *ReplayEngine) Reset() {
	r.index = 0
}
