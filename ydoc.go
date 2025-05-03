package ygo

import (
	"encoding/json"
	"fmt"
	"sort"
	"ygo/internal/block"
	"ygo/internal/blockstore"
	"ygo/internal/decoder"
)

type YDoc struct {
	blockStore     *blockstore.BlockStore
	pendingUpdates []*block.Update
	pendingDeletes []*block.DeleteUpdate
}

func NewYDoc() *YDoc {
	return &YDoc{
		blockStore: blockstore.NewStore(),
	}
}

func (yd *YDoc) Client() int64 {
	return yd.blockStore.GetCurrentClient()
}

func (yd *YDoc) InsertText(pos int64, text string) error {
	yd.blockStore.Insert(pos, text)
	return nil
}

func (yd *YDoc) DeleteText(pos, length int64) error {
	return yd.blockStore.Delete(pos, length)
}

func (yd *YDoc) Content() string {
	return yd.blockStore.Content()
}

// refer readUpdateV2 from yjs in encoding.js
// this should not be the first thing you call
// its important to call InsertText atleast once before calling this
// :) figure out how we can add a marker in this flow
func (yd *YDoc) ApplyUpdate(data []byte) error {
	// Step 1: Decode the binary update into a DecodedUpdate struct
	update, err := decoder.DecodeUpdate(data)
	if err != nil {
		return fmt.Errorf("decode update: %w", err)
	}

	// Step 2: Integrate the blocks from remote clients
	yd.processUpdates(&update.Updates)

	// Step 3: Process deletions
	yd.processDeletes(&update.Deletes)

	// Check if there are any pending updates that can now be processed
	yd.processPendingUpdates()

	return nil
}

func (yd *YDoc) processUpdates(update *block.Update) {
	// Collect all blocks from all clients
	var allBlocks []*block.Block

	for _, blocks := range update.Updates {
		for _, remoteBlock := range blocks {
			// Skip blocks that already exist in store
			if !yd.blockStore.HasBlock(remoteBlock.ID) {
				allBlocks = append(allBlocks, remoteBlock)
			}
		}
	}

	// Sort blocks by clock in ascending order
	sort.Slice(allBlocks, func(i, j int) bool {
		return allBlocks[i].ID.Clock < allBlocks[j].ID.Clock
	})

	restOfTheUpdates := make(map[int64][]*block.Block)

	// Process blocks in sorted order
	for _, remoteBlock := range allBlocks {
		state := yd.blockStore.GetState(remoteBlock.ID.Client)
		offset := state - remoteBlock.ID.Clock
		if offset < 0 {
			// This block is ahead of our current state, so we need to wait
			restOfTheUpdates[remoteBlock.ID.Client] = append(restOfTheUpdates[remoteBlock.ID.Client], remoteBlock)
		} else {
			// Check for missing dependencies
			missingClientID := yd.blockStore.GetMissing(remoteBlock)
			if missingClientID != nil {
				restOfTheUpdates[remoteBlock.ID.Client] = append(restOfTheUpdates[remoteBlock.ID.Client], remoteBlock)
				continue
			}

			// Resolve left and right references
			if remoteBlock.LeftOrigin != (block.ID{}) {
				remoteBlock.Left = yd.blockStore.ResolveNeighborByPreciseBlockID(remoteBlock.LeftOrigin)
			}

			if remoteBlock.RightOrigin != (block.ID{}) {
				remoteBlock.Right = yd.blockStore.ResolveNeighborByPreciseBlockID(remoteBlock.RightOrigin)
			}

			// Integrate the block
			yd.blockStore.Integrate(remoteBlock, int64(offset))
		}
	}

	yd.AddPendingUpdate(&block.Update{
		Updates: restOfTheUpdates,
	})
}

func (yd *YDoc) processDeletes(deletes *block.DeleteUpdate) {
	// Create a structure to track unapplied deletions
	unappliedDeletes := &block.DeleteUpdate{
		NumClients:    0,
		ClientDeletes: []block.ClientDeletes{},
	}

	// Process each client's delete operations
	for _, deletion := range deletes.ClientDeletes {
		clientID := deletion.Client
		state := yd.blockStore.GetState(clientID)

		// Create a slice for unapplied deletions for this client
		clientUnappliedDeletes := []block.DeleteRange{}

		for _, deletedRange := range deletion.DeletedRanges {
			startClock := deletedRange.StartClock
			endClock := startClock + deletedRange.DeleteLength

			// Check if we have already integrated the blocks to be deleted
			if startClock < int64(state) {
				// We can proceed only when the clock we are deleting is already previously integrated
				// If some part of the deletion range is beyond our state, record it as unapplied
				if int64(state) < endClock {
					clientUnappliedDeletes = append(clientUnappliedDeletes, block.DeleteRange{
						StartClock:   int64(state),
						DeleteLength: endClock - int64(state),
					})
				}

				// NOTE: this is a common routine in the code base
				// if you see, the next three code blocks is basically the `refinePreciseBlock` function
				// but we dont use it here since we need to know if a block needed a precise cut
				// and based on that need to increment the index to start the deletion from.

				// Find the starting block for deletion
				index := yd.blockStore.FindIndexInBlockArrayByID(yd.blockStore.Blocks[clientID], block.ID{
					Clock:  int64(startClock),
					Client: clientID,
				})

				// Get the first block and potentially split it
				blk := yd.blockStore.Blocks[clientID][index]

				// If the block starts before our deletion point, we need to split it
				if !blk.IsDeleted && blk.ID.Clock < int64(startClock) {
					diff := int(int64(startClock) - blk.ID.Clock)
					// Split the block at exactly the deletion point
					_ = yd.blockStore.PreciseBlockCut(blk, diff)

					index++ // We want to start deleting from the right part of the split
				}

				// Now process all blocks in the deletion range
				for index < len(yd.blockStore.Blocks[clientID]) {
					blk = yd.blockStore.Blocks[clientID][index]

					if blk.ID.Clock < int64(endClock) {
						if !blk.IsDeleted {
							// check if the endClock sits between the clock range of `blk`
							// if it does, we need to split the block
							if int64(endClock) < blk.ID.Clock+int64(len(blk.Content)) {
								splitPoint := int(int64(endClock) - blk.ID.Clock)
								yd.blockStore.PreciseBlockCut(blk, splitPoint)
							}

							// Mark the block as deleted
							blk.MarkDeleted()
						}
					} else {
						break
					}

					index++
				}
			} else {
				// the start clock itself is ahead of what we've seen for the client so far.
				// which means we havent yet received an update message for it yet
				// so add this deletion range to the unapplied delete set
				clientUnappliedDeletes = append(clientUnappliedDeletes, block.DeleteRange{
					StartClock:   startClock,
					DeleteLength: deletedRange.DeleteLength,
				})
			}
		}

		// If we have unapplied deletions for this client, add them to our tracking
		if len(clientUnappliedDeletes) > 0 {
			unappliedDeletes.ClientDeletes = append(unappliedDeletes.ClientDeletes, block.ClientDeletes{
				Client:        clientID,
				DeletedRanges: clientUnappliedDeletes,
			})
			unappliedDeletes.NumClients++
		}
	}

	// If we have unapplied deletions, store them for later processing
	if unappliedDeletes.NumClients > 0 {
		yd.AddPendingDelete(unappliedDeletes)
	}

}

// Process any pending updates that can now be integrated
func (yd *YDoc) processPendingUpdates() {
	pendingUpdates := yd.GetPendingUpdates()

	for _, pendingUpdate := range pendingUpdates {
		yd.processUpdates(pendingUpdate)
	}

	pendingDeletes := yd.GetPendingDeletes()

	for _, deleteUpdate := range pendingDeletes {
		yd.processDeletes(deleteUpdate)
	}
}

// EncodeStateAsUpdate encodes the current document state as an update message
// that can be applied to other YDoc instances
func (yd *YDoc) EncodeStateAsUpdate() ([]byte, error) {
	// Create an update message containing all blocks in the store
	updates := make(map[int64][]*block.Block)

	// Process each client's blocks
	for clientID, blocks := range yd.blockStore.Blocks {
		// Deep copy blocks to avoid modifying the original store
		clientBlocks := make([]*block.Block, len(blocks))
		for i, b := range blocks {
			// Copy only the block data, not the references
			clientBlocks[i] = &block.Block{
				ID:          b.ID,
				Content:     b.Content,
				IsDeleted:   b.IsDeleted,
				LeftOrigin:  b.LeftOrigin,
				RightOrigin: b.RightOrigin,
				// Leave Left and Right as nil since they're references
			}
		}
		updates[clientID] = clientBlocks
	}

	// Convert the internal deleteSet to a DeleteUpdate structure
	deleteUpdate := createDeleteUpdateFromDeleteSet(yd.blockStore.DeleteSet)

	// Create the update message
	updateMsg := block.Updates{
		Updates: block.Update{
			Updates: updates,
		},
		Deletes: deleteUpdate,
	}

	// Marshal the update message
	return json.Marshal(updateMsg)
}

// EncodeStateVector returns the current state vector as a map of client IDs to clocks
func (yd *YDoc) EncodeStateVector() map[int64]int64 {
	// Return a copy of the state vector
	stateVector := make(map[int64]int64)
	for clientID, clock := range yd.blockStore.StateVector {
		stateVector[clientID] = clock
	}

	return stateVector
}

// AddPendingUpdate adds an update to the pending queue
func (yd *YDoc) AddPendingUpdate(update *block.Update) {
	yd.pendingUpdates = append(yd.pendingUpdates, update)
}

// GetPendingUpdates returns the current pending updates
func (yd *YDoc) GetPendingUpdates() []*block.Update {
	return yd.pendingUpdates
}

// SetPendingUpdates replaces the pending updates list
func (yd *YDoc) SetPendingUpdates(updates []*block.Update) {
	yd.pendingUpdates = updates
}

// AddPendingDelete adds a delete operation to the pending queue
func (yd *YDoc) AddPendingDelete(del *block.DeleteUpdate) {
	yd.pendingDeletes = append(yd.pendingDeletes, del)
}

// GetPendingDeletes returns the current pending deletes
func (yd *YDoc) GetPendingDeletes() []*block.DeleteUpdate {
	return yd.pendingDeletes
}

// SetPendingDeletes replaces the pending deletes list
func (yd *YDoc) SetPendingDeletes(deletes []*block.DeleteUpdate) {
	yd.pendingDeletes = deletes
}

// createDeleteUpdateFromDeleteSet converts the internal deleteSet map to a DeleteUpdate structure
func createDeleteUpdateFromDeleteSet(deleteSet map[int64][]block.DeleteRange) block.DeleteUpdate {
	clientDeletes := make([]block.ClientDeletes, 0, len(deleteSet))

	for clientID, deleteRanges := range deleteSet {
		if len(deleteRanges) > 0 {
			clientDeletes = append(clientDeletes, block.ClientDeletes{
				Client:        clientID,
				DeletedRanges: deleteRanges,
			})
		}
	}

	return block.DeleteUpdate{
		NumClients:    int64(len(clientDeletes)),
		ClientDeletes: clientDeletes,
	}
}
