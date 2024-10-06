package pkg

// Store is a core element of a document
type Store struct {
	// Types is a map of unique top-level type name and the actual type (the branch)
	Types map[string]Branch

	// NodeRegistry is a record of all alive nodes(branches) in the document store.
	NodeRegistry []*Branch

	// A block store of a current document. It represent all blocks (inserted or tombstoned
	// operations) integrated - and therefore visible - into a current document.
	BlockStore *BlockStore

	// A pending update. It contains blocks, which are not yet integrated into our block store due to some issues
	PendingUpdates *PendingUpdate

	// pointer to parent doc, not implemented
	parent int64
}

func New() *Store {
	return &Store{
		Types:          make(map[string]Branch),
		NodeRegistry:   []*Branch{},
		BlockStore:     &BlockStore{},
		PendingUpdates: nil,
		parent:         -1,
	}
}
