package blockstore

import (
	"fmt"
	"math/rand"
	"sort"
	"ygo/internal/block"
	markers "ygo/internal/marker"
	"ygo/internal/utils"
)

type StoreOptions func(*BlockStore)

func WithDebugModeEnabled() StoreOptions {
	return func(s *BlockStore) {}
}

type BlockStore struct {
	Start           *block.Block
	Length          int
	Clock           uint64
	Blocks          map[int64][]*block.Block
	StateVector     map[int64]uint64
	MarkerSystem    *markers.MarkerSystem
	CurrentClientID int64
	pendingUpdates  []*block.Update
	pendingDeletes  []*block.DeleteUpdate
}

// NewStore initializes a new BlockStore.
func NewStore(options ...StoreOptions) *BlockStore {
	b := &BlockStore{
		Blocks:          make(map[int64][]*block.Block),
		StateVector:     make(map[int64]uint64),
		MarkerSystem:    markers.NewSystem(),
		CurrentClientID: rand.Int63(),
	}

	for _, option := range options {
		option(b)
	}

	return b
}

func (s *BlockStore) getNextClock() uint64 {
	s.Clock++
	return s.Clock
}

func (s *BlockStore) adjustLength(content string) {
	s.Length += len(content)

}

func (s *BlockStore) updateState(block *block.Block) {
	current := s.StateVector[block.ID.Client]
	if block.ID.Clock > current {
		s.StateVector[block.ID.Client] = block.ID.Clock
	}
}

func (s *BlockStore) GetState(client int64) uint64 {
	return s.StateVector[client]
}

func (s *BlockStore) GetCurrentClient() int64 {
	return s.CurrentClientID
}

func (s *BlockStore) GetMissing(blk *block.Block) *int64 {
	check := func(origin block.ID) *int64 {
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
			ID:          block.ID{Client: s.CurrentClientID, Clock: s.getNextClock()},
			Content:     content,
			LeftOrigin:  block.ID{},
			RightOrigin: block.ID{},
			IsDeleted:   false,
		}
		s.MarkerSystem.Add(s.Start, 0)
		s.addBlock(s.Start)
		s.adjustLength(content)
		return nil
	}

	// find the correct position
	blockPos, err := s.findPositionForNewBlock(pos)
	if err != nil {
		return fmt.Errorf("find position for new block: %w", err)
	}

	right := &block.Block{
		ID:         block.ID{Client: s.CurrentClientID, Clock: s.getNextClock()},
		Content:    content,
		Left:       blockPos.Left,
		LeftOrigin: blockPos.Left.ID,
		IsDeleted:  false,
	}

	if blockPos.Right != nil {
		right.RightOrigin = blockPos.Right.ID
		right.Right = blockPos.Right
	}

	s.Integrate(right)

	blockPos.Right = right

	s.adjustLength(content)

	return nil
}

// DeleteText marks text as deleted starting from `pos`, over `length` characters.
func (s *BlockStore) Delete(pos, length uint64) error {
	if length > uint64(s.Length) {
		return fmt.Errorf("delete length %d exceeds block store length %d", length, s.Length)
	}
	// find the correct position
	blockPos, err := s.findPositionForNewBlock(pos)
	if err != nil {
		return fmt.Errorf("find position for new block: %w", err)
	}

	// traverse and delete the blocks until `length` is deleted from blockstore
	for length > 0 && blockPos.Right != nil {
		if length < uint64(len(blockPos.Right.Content)) {
			s.getItemCleanStart(block.ID{
				Client: blockPos.Right.ID.Client,
				Clock:  blockPos.Right.ID.Clock + uint64(length),
			})
		}
		length -= uint64(len(blockPos.Right.Content))
		blockPos.Right.MarkDeleted()
		blockPos.Forward()
	}
	return nil
}

// Integrate integrates a remote block into the local block store.
// Core logic for CRDT convergence and conflict resolution.
func (s *BlockStore) Integrate(newBlk *block.Block) {
	// the whole purpose of this branch
	// is to detect conflict and find the perfect left neighbor for `newBlk`
	// this means what we find is a perfect conflict-free position
	// for `newBlk` post this branch logic is reconnection of
	// the left and right neighbors of `newBlk` for nailing
	// the final position of `newBlk` in our BlockStore
	if (newBlk.Left == nil &&
		(newBlk.Right == nil || newBlk.Right.Left == nil)) ||
		(newBlk.Left != nil && newBlk.Left.Right != newBlk.Right) {

		// this is the left pointer. We will find the best
		// left neighbor for the remote block by traversing
		// the block store from the start to the right
		// until we find the perfect left neighbor
		// which does not conflict with the remote block
		var left *block.Block
		left = newBlk.Left

		// Find the first conflict candidate
		o := s.Start
		if left != nil {
			o = left.Right
		}

		// Sets for conflict detection
		conflicts := map[block.ID]bool{}
		seenBefore := map[block.ID]bool{}
		// conflict resolution logic
		for o != nil && o != newBlk.Right {

			// very first thing, add this to the conflicting block set
			// and the seenBefore block set. They are cleared once
			// conflicts are resolved and the appropriate `left` is found
			conflicts[o.ID] = true
			if (o.LeftOrigin != block.ID{}) {
				seenBefore[o.LeftOrigin] = true
			}
			// do the conflicting block and remote block
			// derive from the same origin?
			if utils.EqualID(o.LeftOrigin, newBlk.LeftOrigin) {
				// if yes, order by client id
				if o.ID.Client < newBlk.ID.Client {
					left = o
					// and clear the conflicting items, since we have found new left
					conflicts = map[block.ID]bool{}
				} else if utils.EqualID(o.RightOrigin, newBlk.RightOrigin) {
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
		newBlk.Left = left
	}

	// Reconnect neighbors

	// handles right neighbor when either we have a left from post-conflict resolution
	// OR we have a left from the original block
	// handles left neighbor when it's nil then attaches the block at start
	if newBlk.Left != nil {
		newBlk.Right = newBlk.Left.Right
		newBlk.Left.Right = newBlk
	} else {
		newBlk.Right = s.Start
		s.Start = newBlk
	}
	// final reconnect, handles the right neighbor of the
	// block to the left, which is the `newBlk` itself
	if newBlk.Right != nil {
		newBlk.Right.Left = newBlk
	}

	// add the block to the block store now
	s.Blocks[newBlk.ID.Client] = append(s.Blocks[newBlk.ID.Client], newBlk)
	// update our state vector
	s.updateState(newBlk)
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

func (s *BlockStore) addBlock(blk *block.Block) {
	blocks := s.Blocks[blk.ID.Client]

	n := len(blocks)
	i := sort.Search(n, func(i int) bool {
		return blocks[i].ID.Clock > blk.ID.Clock
	})

	blocks = append(blocks, nil)
	copy(blocks[i+1:], blocks[i:])

	blocks[i] = blk

	s.Blocks[blk.ID.Client] = blocks
}

// find the next appropriate position for integrating a new block
func (s *BlockStore) findPositionForNewBlock(index uint64) (*block.BlockTextListPosition, error) {
	// find marker
	marker, err := s.MarkerSystem.FindMarker(index)
	if err != nil {
		return nil, fmt.Errorf("find marker: %w", err)
	}
	// find next position
	// return the next position
	nextBlock := s.findNextPosition(&block.BlockTextListPosition{
		Left:  marker.Block.Left,
		Right: marker.Block,
		Index: marker.Pos,
	}, index-marker.Pos)

	return nextBlock, nil
}

func (s *BlockStore) findNextPosition(pos *block.BlockTextListPosition, count uint64) *block.BlockTextListPosition {
	if count <= 0 {
		return pos
	}
	// find the next position
	// if necessary, split the block
	// and then return the proper position
	if !pos.Right.IsDeleted {
		if count < uint64(len(pos.Right.Content)) {
			_ = s.getItemCleanStart(block.ID{
				Client: pos.Right.ID.Client,
				Clock:  pos.Right.ID.Clock + uint64(count),
			})
		}
		pos.Index += uint64(len(pos.Right.Content))
		count -= uint64(len(pos.Right.Content))
		// move `pos` to the right
		pos.Left = pos.Right
		pos.Right = pos.Right.Right
	}
	return pos
}

// getItemCleanStart retrieves the block at the clean start position.
// This is the position where the block can be split
// the clock is adjusted accordingly
func (s *BlockStore) getItemCleanStart(id block.ID) *block.Block {
	structs := s.Blocks[id.Client]
	index := s.findIndexCleanStart(structs, id)
	structs = s.Blocks[id.Client]
	return structs[index]
}

func (s *BlockStore) findIndexCleanStart(blocks []*block.Block, id block.ID) int {
	index := s.FindIndexInBlockArrayByID(blocks, id)
	blk := blocks[index]

	if blk.ID.Clock <= id.Clock {
		s.SplitBlock(blk, int(id.Clock)-int(blk.ID.Clock))
		// because we split the block, we deal with the right of the blk
		// which is the new block and hence we return index + 1
		return index + 1
	}

	panic(fmt.Sprintf("getItemCleanStart: no block found for ID %v", id))
}

// equivalent to findIndexSS from yjs
func (s *BlockStore) FindIndexInBlockArrayByID(blocks []*block.Block, id block.ID) int {
	for i, blk := range blocks {
		if blk.ID.Clock == id.Clock || id.Clock < blk.ID.Clock+uint64(len(blk.Content)) {
			return i
		}
	}
	panic(fmt.Sprintf("findIndexInBlockArrayByID: no exact match for ID %v", id))
}

func (s *BlockStore) SplitBlock(left *block.Block, diff int) *block.Block {
	if diff <= 0 || diff >= len(left.Content) {
		panic(fmt.Sprintf("SplitBlock: invalid split position %d in block with length %d", diff, len(left.Content)))
	}

	// Create the right block
	right := &block.Block{
		ID:          block.ID{Client: left.ID.Client, Clock: left.ID.Clock + uint64(diff)},
		Content:     left.Content[diff:],
		IsDeleted:   left.IsDeleted,
		LeftOrigin:  block.ID{Client: left.ID.Client, Clock: left.ID.Clock + uint64(diff-1)},
		RightOrigin: left.RightOrigin,
		Left:        left,
		Right:       left.Right,
	}

	// Adjust left block
	left.Content = left.Content[:diff]
	left.Right = right

	// Fix neighbor pointer if right was non-nil
	if right.Right != nil {
		right.Right.Left = right
	}

	// Insert new block into BlockStore
	s.addBlock(right)
	return right
}

// internal/blockstore/store.go

// HasBlock checks if a block with the given ID exists in the store
func (s *BlockStore) HasBlock(id block.ID) bool {
	blocks, ok := s.Blocks[id.Client]
	if !ok {
		return false
	}

	for _, b := range blocks {
		if b.ID.Clock == id.Clock {
			return true
		}
	}
	return false
}

// GetBlockByID retrieves a block by its ID
func (s *BlockStore) GetBlockByID(id block.ID) *block.Block {
	blocks, ok := s.Blocks[id.Client]
	if !ok {
		return nil
	}

	for _, b := range blocks {
		if b.ID.Clock == id.Clock {
			return b
		}
	}
	return nil
}

// GetBlocksInRange returns blocks from a specific client within a clock range
func (s *BlockStore) GetBlocksInRange(client int64, startClock int64, length int64) []*block.Block {
	var result []*block.Block
	blocks, ok := s.Blocks[client]
	if !ok {
		return result
	}

	endClock := startClock + length

	for _, b := range blocks {
		if b.ID.Clock >= uint64(startClock) && b.ID.Clock < uint64(endClock) {
			result = append(result, b)
		}
	}
	return result
}

// AddPendingUpdate adds an update to the pending queue
func (s *BlockStore) AddPendingUpdate(update *block.Update) {
	s.pendingUpdates = append(s.pendingUpdates, update)
}

// GetPendingUpdates returns the current pending updates
func (s *BlockStore) GetPendingUpdates() []*block.Update {
	return s.pendingUpdates
}

// SetPendingUpdates replaces the pending updates list
func (s *BlockStore) SetPendingUpdates(updates []*block.Update) {
	s.pendingUpdates = updates
}

// AddPendingDelete adds a delete operation to the pending queue
func (s *BlockStore) AddPendingDelete(del *block.DeleteUpdate) {
	s.pendingDeletes = append(s.pendingDeletes, del)
}

// GetPendingDeletes returns the current pending deletes
func (s *BlockStore) GetPendingDeletes() []*block.DeleteUpdate {
	return s.pendingDeletes
}

// SetPendingDeletes replaces the pending deletes list
func (s *BlockStore) SetPendingDeletes(deletes []*block.DeleteUpdate) {
	s.pendingDeletes = deletes
}
