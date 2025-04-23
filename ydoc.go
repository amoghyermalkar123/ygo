package ygo

import (
	"fmt"
	"ygo/internal/block"
	"ygo/internal/blockstore"
	"ygo/internal/decoder"
)

type YDoc struct {
	blockStore *blockstore.BlockStore
}

func NewYDoc() *YDoc {
	return &YDoc{
		blockStore: blockstore.NewStore(),
	}
}

func (yd *YDoc) InsertText(pos uint64, text string) error {
	return yd.blockStore.Insert(pos, text)
}

func (yd *YDoc) DeleteText(pos, length uint64) error {
	return yd.blockStore.Delete(pos, length)
}

func (yd *YDoc) Content() string {
	return yd.blockStore.Content()
}

// refer readUpdateV2 from yjs in encoding.js
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
	for _, blocks := range update.Updates {
		for _, remoteBlock := range blocks {
			// Check if block already exists in store
			if yd.blockStore.HasBlock(remoteBlock.ID) {
				continue
			}

			// Check for missing dependencies
			missingClientID := yd.blockStore.GetMissing(remoteBlock)
			if missingClientID != nil {
				// We have missing dependencies, add to pending updates
				yd.blockStore.AddPendingUpdate(update)
			}

			// Resolve left and right references
			if remoteBlock.LeftOrigin != (block.ID{}) {
				remoteBlock.Left = yd.blockStore.GetBlockByID(remoteBlock.LeftOrigin)
			}

			if remoteBlock.RightOrigin != (block.ID{}) {
				remoteBlock.Right = yd.blockStore.GetBlockByID(remoteBlock.RightOrigin)
			}

			// Integrate the block
			yd.blockStore.Integrate(remoteBlock)
		}
	}
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
		clientBlocks := yd.blockStore.Blocks[clientID]
		state := yd.blockStore.GetState(clientID)

		// Create a slice for unapplied deletions for this client
		clientUnappliedDeletes := []block.DeleteRange{}

		for _, deletedRange := range deletion.DeletedRanges {
			startClock := deletedRange.StartClock
			endClock := startClock + deletedRange.DeleteLength

			// Check if we have already integrated the blocks to be deleted
			if startClock < int64(state) {
				// We can proceed only when the clock we are deleting is already previously integrated
				// If some part of the range is beyond our state, record it as unapplied
				if int64(state) < endClock {
					clientUnappliedDeletes = append(clientUnappliedDeletes, block.DeleteRange{
						StartClock:   int64(state),
						DeleteLength: endClock - int64(state),
					})
				}

				// Find the starting block for deletion
				index := yd.blockStore.FindIndexInBlockArrayByID(clientBlocks, block.ID{
					Clock:  uint64(startClock),
					Client: clientID,
				})

				// Get the first block and potentially split it
				blk := clientBlocks[index]

				// If the block starts before our deletion point, we need to split it
				if !blk.IsDeleted && blk.ID.Clock < uint64(startClock) {
					diff := int(uint64(startClock) - blk.ID.Clock)
					// Split the block at exactly the deletion point
					_ = yd.blockStore.SplitBlock(blk, diff)

					index++ // We want to start deleting from the right part of the split
				}

				// Now process all blocks in the deletion range
				for index < len(clientBlocks) {
					blk = clientBlocks[index]

					if blk.ID.Clock < uint64(endClock) {
						if !blk.IsDeleted {
							// check if the endClock sits between the clock range of `blk`
							// if it does, we need to split the block
							if uint64(endClock) < blk.ID.Clock+uint64(len(blk.Content)) {
								splitPoint := int(uint64(endClock) - blk.ID.Clock)
								yd.blockStore.SplitBlock(blk, splitPoint)
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
		yd.blockStore.AddPendingDelete(unappliedDeletes)
	}

}

// Process any pending updates that can now be integrated
func (yd *YDoc) processPendingUpdates() {
	pendingUpdates := yd.blockStore.GetPendingUpdates()
	if len(pendingUpdates) == 0 {
		return
	}

	// This is a simplified approach to handle pending updates
	// In a real implementation, you should check for dependency resolution
	// and only process updates whose dependencies are met
	newPendingUpdates := make([]*block.Update, 0)

	for _, update := range pendingUpdates {
		stillPending := false

		// Try to apply the update
		for _, blocks := range update.Updates {
			for _, remoteBlock := range blocks {
				if yd.blockStore.HasBlock(remoteBlock.ID) {
					continue // Already integrated
				}

				missingClientID := yd.blockStore.GetMissing(remoteBlock)
				if missingClientID != nil {
					stillPending = true
					break // Missing dependencies, can't integrate yet
				}

				// Resolve left and right references
				if remoteBlock.LeftOrigin != (block.ID{}) {
					remoteBlock.Left = yd.blockStore.GetBlockByID(remoteBlock.LeftOrigin)
				}
				if remoteBlock.RightOrigin != (block.ID{}) {
					remoteBlock.Right = yd.blockStore.GetBlockByID(remoteBlock.RightOrigin)
				}

				// Integrate the block
				yd.blockStore.Integrate(remoteBlock)
			}
			if stillPending {
				break
			}
		}

		if stillPending {
			newPendingUpdates = append(newPendingUpdates, update)
		}
	}

	pendingDeletes := yd.blockStore.GetPendingDeletes()

	for _, deleteUpdate := range pendingDeletes {
		yd.processDeletes(deleteUpdate)
	}

	// Update the pending updates list
	yd.blockStore.SetPendingUpdates(newPendingUpdates)
}
