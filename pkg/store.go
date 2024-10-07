package pkg

// Store is a core element of a document
type Store struct {
	// NodeRegistry is a record of all alive nodes(branches) in the document store.
	NodeRegistry []*Branch

	// A block store of a current document. It represent all blocks (inserted or tombstoned
	// operations) integrated - and therefore visible - into a current document.
	BlockStore *BlockStore

	// A pending update. It contains blocks, which are not yet integrated into our block store due to some issues
	PendingUpdates *PendingUpdate
}

func NewStore(clientID ID) *Store {
	b := NewBlockStore(clientID)
	br := NewBranch(b.clientBlocks[clientID.ClientID].First(), "branchname")

	return &Store{
		NodeRegistry:   []*Branch{br},
		BlockStore:     b,
		PendingUpdates: nil,
	}
}
