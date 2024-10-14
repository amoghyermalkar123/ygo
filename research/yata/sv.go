package yata

// StateVector tracks the version of each client's document.
type StateVector map[uint32]uint32

// NewStateVector creates a new state vector.
func NewStateVector() StateVector {
	return make(StateVector)
}

// UpdateState updates the state vector for a client.
func (sv StateVector) UpdateState(clientID, clock uint32) {
	sv[clientID] = clock
}

// Merge merges another state vector into this one.
func (sv StateVector) Merge(other StateVector) {
	for clientID, clock := range other {
		if sv[clientID] < clock {
			sv[clientID] = clock
		}
	}
}

// HasSeen checks if this state vector has seen the given client's clock.
func (sv StateVector) HasSeen(clientID, clock uint32) bool {
	return sv[clientID] >= clock
}
