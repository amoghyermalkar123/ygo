package replay

import "ygo/internal/block"

type Playback struct {
	doc   []block.Event
	index int
}

func NewPlayback(events []block.Event) *Playback {
	return &Playback{doc: events}
}

func (r *Playback) PlayNext() block.Event {
	if r.index >= len(r.doc) {
		return block.Event{}
	}
	evt := r.doc[r.index]
	r.index++
	return evt
}

func (r *Playback) Reset() {
	r.index = 0
}
