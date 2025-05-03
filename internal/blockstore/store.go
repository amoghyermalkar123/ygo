package blockstore

import (
	"fmt"
	"math/rand"
	"sort"

	"ygo/internal/block"
	markers "ygo/internal/marker"
	"ygo/internal/utils"
	"ygo/logger"

	"go.uber.org/zap"
)

type BlockStore struct {
	Start  *block.Block
	Length int
	Clock  int64
	Blocks map[int64][]*block.Block
	// SV always stores the next expected clock for a client
	StateVector     map[int64]int64
	MarkerSystem    *markers.MarkerSystem
	CurrentClientID int64
	// lists all deletions performed by the CurentClientID
	DeleteSet map[int64][]block.DeleteRange
}

// NewStore initializes a new BlockStore.
func NewStore() *BlockStore {
	b := &BlockStore{
		Blocks:          make(map[int64][]*block.Block),
		StateVector:     make(map[int64]int64),
		MarkerSystem:    markers.NewSystem(),
		CurrentClientID: rand.Int63(),
		DeleteSet:       make(map[int64][]block.DeleteRange),
	}

	return b
}

func (s *BlockStore) adjustLength(content string) {
	s.Length += len(content)
}

func (s *BlockStore) updateState(block *block.Block) {
	if current, ok := s.StateVector[block.ID.Client]; ok {
		if block.ID.Clock > current {
			s.StateVector[block.ID.Client] = block.ID.Clock + int64(len(block.Content))
		}
	} else {
		s.StateVector[block.ID.Client] = block.ID.Clock + int64(len(block.Content))
	}
}

func (s *BlockStore) GetState(client int64) int64 {
	if clock, ok := s.StateVector[client]; ok {
		return clock
	}
	return 0
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
func (s *BlockStore) Insert(pos int64, content string) error {
	logger.Info("insert text", zap.Int64("pos", pos), zap.String("content", content))

	// find the correct position
	blockPos, err := s.findPositionForNewBlock(pos)
	if err != nil {
		return fmt.Errorf("find position for new block: %w", err)
	}

	// create a brand new block
	newBlk := &block.Block{
		ID:        block.ID{Client: s.CurrentClientID, Clock: s.GetState(s.CurrentClientID)},
		Content:   content,
		IsDeleted: false,
	}

	// if we found a viable left neighbor from findPositionForNewBlock
	// attach it
	if blockPos.Left != nil {
		newBlk.Left = blockPos.Left
		newBlk.LeftOrigin = blockPos.Left.ID
	}

	// if we found a viable right neighbor from findPositionForNewBlock
	// attach it
	if blockPos.Right != nil {
		newBlk.RightOrigin = blockPos.Right.ID
		newBlk.Right = blockPos.Right
	}

	// start integration
	s.Integrate(newBlk, 0)

	blockPos.Right = newBlk

	s.adjustLength(content)

	s.MarkerSystem.Add(newBlk, pos)

	return nil
}

// DeleteText marks text as deleted starting from `pos`, over `length` characters.
func (s *BlockStore) Delete(pos, length int64) error {
	if length > int64(s.Length) {
		return fmt.Errorf("delete length %d exceeds block store length %d", length, s.Length)
	}
	// find the correct position
	blockPos, err := s.findPositionForNewBlock(pos)
	if err != nil {
		return fmt.Errorf("find position for new block: %w", err)
	}

	// traverse and delete the blocks until `length` is deleted from blockstore
	for length > 0 && blockPos.Right != nil {
		if length < int64(len(blockPos.Right.Content)) {
			s.refinePreciseBlock(block.ID{
				Client: blockPos.Right.ID.Client,
				Clock:  blockPos.Right.ID.Clock + int64(length),
			})
		}

		s.addToDeleteSet(s.CurrentClientID, blockPos.Right.ID.Clock, int64(len(blockPos.Right.Content)))

		length -= int64(len(blockPos.Right.Content))

		blockPos.Right.MarkDeleted()

		blockPos.Forward()
	}

	return nil
}

func (s *BlockStore) addToDeleteSet(client int64, startClock, length int64) {
	s.DeleteSet[client] = append(s.DeleteSet[client], block.DeleteRange{
		StartClock:   startClock,
		DeleteLength: length,
	})
}

// getItemCleanEnd retrieves or creates a block that ends exactly at the specified ID.
// This is similar to refinePreciseBlock but focuses on the end position.
func (s *BlockStore) getItemCleanEnd(id block.ID) *block.Block {
	structs := s.Blocks[id.Client]
	if len(structs) == 0 {
		return nil
	}

	// Find the block that contains the ID
	index := s.FindIndexInBlockArrayByID(structs, id)
	blk := structs[index]

	// If the ID is not exactly at the end of the block, we need to split
	if id.Clock != blk.ID.Clock+int64(len(blk.Content))-1 && !blk.IsDeleted {
		// Calculate the position to split: difference between target ID and block start + 1
		// here id.Clock and blk.ID.Clock belong to same block
		// so when we do id.Clock - blk.ID.Clock + 1
		// we get the exact position where the existing left block ends
		// and the new block starts
		splitPosition := int(id.Clock - blk.ID.Clock + 1)

		// Create a new block by splitting the existing one
		s.PreciseBlockCut(blk, splitPosition)

		// After splitting, the original block 'blk' now ends exactly at id.Clock
	}

	return blk
}

// Integrate integrates a remote block into the local block store.
// Core logic for CRDT convergence and conflict resolution.
func (s *BlockStore) Integrate(newBlk *block.Block, offset int64) {
	// offset is localClock - remoteClock
	// if its greater than 0 and less than the length of the block
	// it means the new blk needs to be added somewhere in between
	// the current block and hence that block needs a split
	if offset > 0 {
		// Adjust the clock
		newBlk.ID.Clock = newBlk.ID.Clock + offset

		// Find or create the left block that ends exactly where this block should start
		// here newBlk.ID.Clock - 1 indicates the exact end of the left block we want
		leftBlock := s.getItemCleanEnd(block.ID{Client: newBlk.ID.Client, Clock: newBlk.ID.Clock - 1})
		newBlk.Left = leftBlock

		// Update the origin to point to the end of the left block
		newBlk.LeftOrigin = leftBlock.ID

		// Trim the content to remove the already integrated part
		if len(newBlk.Content) > int(offset) {
			newBlk.Content = newBlk.Content[offset:]
		} else {
			newBlk.Content = ""
		}
	}

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

	// add the new block to the block store
	s.addBlock(newBlk)
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
func (s *BlockStore) findPositionForNewBlock(index int64) (*block.BlockTextListPosition, error) {
	textListPosition := &block.BlockTextListPosition{}

	// find marker
	marker, _ := s.MarkerSystem.FindMarker(index)

	if (marker == markers.Marker{}) {
		textListPosition = &block.BlockTextListPosition{
			Right: s.Start,
			Index: 0,
		}
	} else {

		textListPosition = &block.BlockTextListPosition{
			Left:  marker.Block.Left,
			Right: marker.Block,
			Index: marker.Pos,
		}

	}

	// marker.Pos always point to the start of the block
	// so index-marker.Pos is the offset from the start of the block
	// to the position where the user wants to insert
	updatedTLP := s.refineTextListPosition(textListPosition, index-marker.Pos)

	return updatedTLP, nil
}

func (s *BlockStore) refineTextListPosition(pos *block.BlockTextListPosition, blockOffset int64) *block.BlockTextListPosition {
	// no block split at index required
	// no need to refine the text list position
	if blockOffset <= 0 {
		return pos
	}
	// find the next position
	// if necessary, split the block
	// and then return the proper position
	if !pos.Right.IsDeleted {
		// we deal with the right block
		// so check if the offset is within the block
		// if yes, we need a clean start so split the block
		if blockOffset < int64(len(pos.Right.Content)) {
			_ = s.refinePreciseBlock(block.ID{
				Client: pos.Right.ID.Client,
				Clock:  pos.Right.ID.Clock + int64(blockOffset),
			})
		}
		// if the block didn't required to be split, then the below ops
		// simple would mean left becomes the marker block we found
		// right becomes the right to the marker block
		// since `pos` passed to this function is generally by `findPositionForNewBlock`
		// where pos.Left is the left of the marker and pos.Right is the marker block itself
		// while insertion, you will see this works out for us, check `Insert`
		pos.Index += int64(len(pos.Right.Content))
		blockOffset -= int64(len(pos.Right.Content))
		// move `pos` to the right
		pos.Left = pos.Right
		pos.Right = pos.Right.Right
	}
	return pos
}

// refinePreciseBlock refines the block at the precise start position.
// This is the position where the block is split
// based on the clock provided in the id.
func (s *BlockStore) refinePreciseBlock(id block.ID) *block.Block {

	index := s.FindIndexInBlockArrayByID(s.Blocks[id.Client], id)

	blk := s.Blocks[id.Client][index]

	if !blk.IsDeleted && blk.ID.Clock <= id.Clock {
		s.PreciseBlockCut(blk, int(id.Clock)-int(blk.ID.Clock))
		// because we split the block, we deal with the right of the blk
		// which is the new block and hence we return index + 1
		return s.Blocks[id.Client][index+1]
	}

	return blk
}

// equivalent to findIndexSS from yjs
// this can be optimized to a hasmpa with O(1) lookup
// but for now we will use a linear search
func (s *BlockStore) FindIndexInBlockArrayByID(blocks []*block.Block, id block.ID) int {
	for i, blk := range blocks {
		if blk.ID.Clock == id.Clock || id.Clock < blk.ID.Clock+int64(len(blk.Content)) {
			return i
		}
	}
	panic(fmt.Sprintf("findIndexInBlockArrayByID: no exact match for ID %v", id))
}

// PreciseBlockCut splits a block at the precise position of the diff provided to it
func (s *BlockStore) PreciseBlockCut(left *block.Block, diff int) *block.Block {
	if diff <= 0 || diff >= len(left.Content) {
		panic(fmt.Sprintf("PreciseBlockCut: invalid split position %d in block with length %d", diff, len(left.Content)))
	}

	// Create the right block
	right := &block.Block{
		ID:          block.ID{Client: left.ID.Client, Clock: left.ID.Clock + int64(diff)},
		Content:     left.Content[diff:],
		IsDeleted:   left.IsDeleted,
		LeftOrigin:  block.ID{Client: left.ID.Client, Clock: left.ID.Clock + int64(diff-1)},
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
	state := s.GetState(id.Client)
	if state == 0 || id.Clock >= state {
		return false
	}

	return true
}

// GetPreciseBlockByID retrieves a block by its ID, splitting a block if necessary
// at the precise position of the ID.
func (s *BlockStore) ResolveNeighborByPreciseBlockID(originID block.ID) *block.Block {
	blocks, ok := s.Blocks[originID.Client]
	if !ok {
		return nil
	}

	for _, b := range blocks {
		if b.ID.Clock == originID.Clock {
			return b
		}
	}

	// If no exact match, look for a block that contains this ID
	for _, b := range blocks {
		blockStart := b.ID.Clock
		blockEnd := blockStart + int64(len(b.Content)) - 1

		// basically check if the block we want as a neighbor
		// as suggested by the clock in the `originID` falls somewhere
		// between an existing block's clock range, if yes that means
		// that the new block is placed in between the existing block we have
		// and we need to split this existing block
		if originID.Clock >= blockStart && originID.Clock <= blockEnd {
			// We found a block that contains this ID
			// We need to split it to create the exact block we're looking for

			// If we're looking for the end of a block
			if originID.Clock == blockEnd {
				return b
			}

			// If we're looking for the start of a block
			if originID.Clock == blockStart {
				return b
			}

			// If we're looking for a position in the middle of a block
			// Split the block at the appropriate position
			if originID.Clock > blockStart {
				// Split at id.Clock - blockStart
				splitPos := int(originID.Clock - blockStart)
				if splitPos > 0 && splitPos < len(b.Content) {
					right := s.PreciseBlockCut(b, splitPos)
					// If we want the start of the right part
					return right
				}
			}
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
		if b.ID.Clock >= int64(startClock) && b.ID.Clock < int64(endClock) {
			result = append(result, b)
		}
	}
	return result
}
