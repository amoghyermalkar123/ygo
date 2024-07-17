package pkg

type Options struct{}

type StoreEvents struct{}

// Store is a core element of a document
type Store struct {
	// TODO: this
	Options Options

	// Types is a map of unique top-level type name and the actual type (the branch)
	Types map[string]Branch

	// NodeRegistry is a record of all alive nodes(branches) in the document store.
	NodeRegistry []*Branch

	// A block store of a current document. It represent all blocks (inserted or tombstoned
	// operations) integrated - and therefore visible - into a current document.
	BlockStore *BlockStore

	// A pending update. It contains blocks, which are not yet integrated into our block store due to some issues
	// TODO: this
	PendingUpdates *PendingUpdate

	// A pending delete set. Just like PendingUpdates, it contains deleted ranges of blocks that have
	// not been yet applied and are yet to be integrated into BlockStore
	// TODO: this
	PendingDeletes any

	// sub document, not implemented
	subdocs int64

	// Event management
	Events StoreEvents

	// pointer to parent doc, not implemented
	parent int64

	// weak links, not implemented
	linkedBy int64
}

func New() *Store                              { return nil }
func (s *Store) PendingUpdate() *PendingUpdate { return s.PendingUpdates }
