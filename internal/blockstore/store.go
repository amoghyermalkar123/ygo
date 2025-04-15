package blockstore

import (
	"fmt"
	"ygo/internal/block"
	markers "ygo/internal/marker"
	"ygo/internal/utils"
)

// persist in db later and gen randomly
const CLIENT_ID = 1

type BlockStore struct {
	Start        *block.Block
	Length       int
	Clock        uint64
	Blocks       map[block.ID]*block.Block
	StateVector  map[uint64]uint64
	MarkerSystem *markers.MarkerSystem
}

// NewStore initializes a new BlockStore.
func NewStore() *BlockStore {
	return &BlockStore{
		Blocks:       make(map[block.ID]*block.Block),
		StateVector:  make(map[uint64]uint64),
		MarkerSystem: markers.NewSystem(),
	}
}

func (s *BlockStore) getNextClock() uint64 {
	s.Clock++
	return s.Clock
}

func (s *BlockStore) updateState(block *block.Block) {
	current := s.StateVector[block.ID.Client]
	if block.ID.Clock > current {
		s.StateVector[block.ID.Client] = block.ID.Clock
	}
}

func (s *BlockStore) GetState(client uint64) uint64 {
	return s.StateVector[client]
}

func (s *BlockStore) GetMissing(blk *block.Block) *uint64 {
	check := func(origin block.ID) *uint64 {
		if (origin != block.ID{} && origin.Client != blk.ID.Client && origin.Clock > s.GetState(origin.Client)) {
			return &origin.Client
		}
		return nil
	}
	if m := check(blk.LeftOrigin); m != nil {
		return m
	}
	if m := check(blk.RightOrigin); m != nil {
		return m
	}
	return nil
}

// InsertText inserts content at a given position (supports split).
func (s *BlockStore) Insert(pos uint64, content string) error {
	if s.Start == nil {
		s.Start = &block.Block{
			ID:          block.ID{Client: CLIENT_ID, Clock: s.getNextClock()},
			Content:     content,
			LeftOrigin:  block.ID{},
			RightOrigin: block.ID{},
			IsDeleted:   false,
		}
		s.MarkerSystem.Add(s.Start, 0)
		return nil
	}

	marker, err := s.MarkerSystem.FindMarker(pos)
	if err != nil {
		return fmt.Errorf("find marker: %w", err)
	}

	right := &block.Block{
		ID:          block.ID{Client: CLIENT_ID, Clock: s.getNextClock()},
		Content:     content,
		Left:        marker.Block,
		LeftOrigin:  marker.Block.ID,
		RightOrigin: marker.Block.RightOrigin,
		IsDeleted:   false,
	}

	if pos > marker.Pos && pos < uint64(len(marker.Block.Content)) {
		splitPoint := uint64(len(marker.Block.Content)) - pos - 1

		right := &block.Block{
			ID:          block.ID{Client: CLIENT_ID, Clock: s.getNextClock()},
			Content:     marker.Block.Content[splitPoint:],
			Left:        marker.Block,
			LeftOrigin:  marker.Block.ID,
			RightOrigin: marker.Block.RightOrigin,
			Right:       marker.Block.Right,
			IsDeleted:   false,
		}

		marker.Block.Content = marker.Block.Content[:splitPoint]
		marker.Block.Right = right

	}

	s.Length += len(marker.Block.Content)
	s.Integrate(right)
	return nil
}

// DeleteText marks text as deleted starting from `pos`, over `length` characters.
// :)
func (s *BlockStore) DeleteText(pos, length uint64) error {
	if s.Start == nil {
		return fmt.Errorf("block store is empty")
	}

	marker, err := s.MarkerSystem.FindMarker(uint64(pos))
	if err != nil {
		return fmt.Errorf("find marker: %w", err)
	}

	s.Length -= int(length)

	for length > 0 && marker.Block != nil {
		if length < uint64(len(marker.Block.Content)) {
			splitPoint := uint64(len(marker.Block.Content)) - length - 1

			left := &block.Block{
				ID:          block.ID{Client: CLIENT_ID, Clock: s.getNextClock()},
				Content:     "",
				Left:        marker.Block.Left,
				LeftOrigin:  marker.Block.LeftOrigin,
				RightOrigin: marker.Block.ID,
				Right:       marker.Block,
				IsDeleted:   true,
			}

			marker.Block.Content = marker.Block.Content[:splitPoint] + marker.Block.Content[length+1:]
			marker.Block.Left = left
			length = 0
			continue
		}

		length -= uint64(len(marker.Block.Content))
		marker.Block.IsDeleted = true
		marker.Block.Content = ""
	}
	return nil
}

// Integrate integrates a remote block into the local block store.
// Core logic for CRDT convergence and conflict resolution.
func (s *BlockStore) Integrate(remoteBlock *block.Block) {
	// the whole purpose of this branch
	// is to detect conflict and find the perfect left neighbor for `remoteBlock`
	// this means what we find is a perfect conflict-free position
	// for `remoteBlock` post this branch logic is reconnection of
	// the left and right neighbors of `remoteBlock` for nailing
	// the final position of `remoteBlock` in our BlockStore
	if (remoteBlock.Left == nil &&
		(remoteBlock.Right == nil || remoteBlock.Right.Left == nil)) ||
		(remoteBlock.Left != nil && remoteBlock.Left.Right != remoteBlock.Right) {

		// this is the left pointer. We will find the best
		// left neighbor for the remote block by traversing
		// the block store from the start to the right
		// until we find the perfect left neighbor
		// which does not conflict with the remote block
		var left *block.Block
		left = remoteBlock.Left

		// Find the first conflict candidate
		o := s.Start
		if left != nil {
			o = left.Right
		}

		// Sets for conflict detection
		conflicts := map[block.ID]bool{}
		seenBefore := map[block.ID]bool{}
		// conflict resolution logic
		for o != nil && o != remoteBlock.Right {
			// very first thing, add this to the conflicting block set
			// and the seenBefore block set. They are cleared once
			// conflicts are resolved and the appropriate `left` is found
			conflicts[o.ID] = true
			if (o.LeftOrigin != block.ID{}) {
				seenBefore[o.LeftOrigin] = true
			}
			// do the conflicting block and remote block
			// derive from the same origin?
			if utils.EqualID(o.LeftOrigin, remoteBlock.LeftOrigin) {
				// if yes, order by client id
				if o.ID.Client < remoteBlock.ID.Client {
					left = o
					// and clear the conflicting items, since we have found new left
					conflicts = map[block.ID]bool{}
				} else if utils.EqualID(o.RightOrigin, remoteBlock.RightOrigin) {
					// if remote block client is greater and we have same right origins
					// break here since they will naturalyl be in correct order
					break
				}
			} else if (o.LeftOrigin != block.ID{}) {
				// if no, check if we have seen the conflicting blocks left origin before
				// if yes, check if it also conflicts
				// if no, we have found new left, clear the conflicting items
				if seenBefore[o.LeftOrigin] && !conflicts[o.LeftOrigin] {
					left = o
					conflicts = map[block.ID]bool{}
				}
			}
			// move ahead one block to the right
			// since we process one block at a time, left -> right
			o = o.Right
		}
		// set the remote block left to `left` since it's the best
		// one found by the above conflict resolution logic
		remoteBlock.Left = left
	}

	// Reconnect neighbors

	// handles right neighbor when either we have a left from post-conflict resolution
	// OR we have a left from the original block
	// handles left neighbor when it's nil then attaches the block at start
	if remoteBlock.Left != nil {
		remoteBlock.Right = remoteBlock.Left.Right
		remoteBlock.Left.Right = remoteBlock
	} else {
		remoteBlock.Right = s.Start
		s.Start = remoteBlock
	}
	// final reconnect, handles the right neighbor of the
	// block to the left, which is the `remoteBlock` itself
	if remoteBlock.Right != nil {
		remoteBlock.Right.Left = remoteBlock
	}
	// add the block to the block store now
	s.Blocks[remoteBlock.ID] = remoteBlock
	// update our state vector
	s.updateState(remoteBlock)
}

func (s *BlockStore) Content() string {
	curr := s.Start
	content := ""
	for curr != nil {
		if !curr.IsDeleted {
			content += curr.Content
		}
		curr = curr.Right
	}
	return content
}
