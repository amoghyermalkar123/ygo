package blockstore

import (
	"fmt"
	"ygo/internal/block"
	"ygo/internal/marker"
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
	MarkerSystem *marker.MarkerSystem
}

// NewStore initializes a new BlockStore.
func NewStore() *BlockStore {
	return &BlockStore{
		Blocks:       make(map[block.ID]*block.Block),
		StateVector:  make(map[uint64]uint64),
		MarkerSystem: marker.NewSystem(),
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
func (s *BlockStore) Insert(pos uint64, content string, index uint64) error {
	marker, err := s.MarkerSystem.FindMarker(s.Start, pos)
	if err != nil {
		return fmt.Errorf("find marker: %w", err)
	}

	if err := s.insertText(marker, content, index); err != nil {
		return fmt.Errorf("insert text: %w", err)
	}

	return nil
}

func (s *BlockStore) insertText(marker marker.Marker, content string, index uint64) error {
	right := &block.Block{
		ID:          block.ID{Client: CLIENT_ID, Clock: s.getNextClock()},
		Content:     content,
		Left:        marker.Block,
		LeftOrigin:  marker.Block.ID,
		RightOrigin: marker.Block.RightOrigin,
		IsDeleted:   false,
	}

	if index > marker.Pos && index < uint64(len(marker.Block.Content)) {
		splitPoint := uint64(len(marker.Block.Content)) - index - 1

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

		s.Length += len(marker.Block.Content)
	}

	s.Integrate(right)
	return nil
}

// DeleteText marks text as deleted starting from `pos`, over `length` characters.
func (s *BlockStore) DeleteText(pos, length int) error {
	return nil
}

func (s *BlockStore) Integrate(remoteBlock *block.Block) {
	if remoteBlock.Left != nil || remoteBlock.Right != nil {
		// Already linked – assume resolved
		return
	}

	var left *block.Block
	if (remoteBlock.LeftOrigin != block.ID{}) {
		left = s.Blocks[remoteBlock.LeftOrigin]
	}

	// Find the first conflict candidate
	o := s.Start
	if left != nil {
		o = left.Right
	}

	// Sets for conflict detection
	conflicts := map[block.ID]bool{}
	seenBefore := map[block.ID]bool{}

	for o != nil && !utils.EqualID(o.ID, remoteBlock.RightOrigin) {
		conflicts[o.ID] = true
		if (o.LeftOrigin != block.ID{}) {
			seenBefore[o.LeftOrigin] = true
		}

		// Compare origins
		if utils.EqualID(o.LeftOrigin, remoteBlock.LeftOrigin) {
			// Order by client ID
			if o.ID.Client < remoteBlock.ID.Client {
				left = o
				conflicts = map[block.ID]bool{}
			} else if utils.EqualID(o.RightOrigin, remoteBlock.RightOrigin) {
				break
			}
		} else if (o.LeftOrigin != block.ID{}) {
			// Origin crossing logic
			if seenBefore[o.LeftOrigin] && !conflicts[o.LeftOrigin] {
				left = o
				conflicts = map[block.ID]bool{}
			}
		}
		o = o.Right
	}

	// Reconnect neighbors
	remoteBlock.Left = left
	if left != nil {
		remoteBlock.Right = left.Right
		left.Right = remoteBlock
	} else {
		remoteBlock.Right = s.Start
		s.Start = remoteBlock
	}

	if remoteBlock.Right != nil {
		remoteBlock.Right.Left = remoteBlock
	}

	s.Blocks[remoteBlock.ID] = remoteBlock
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
