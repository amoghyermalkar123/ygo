package yata

// CRDT represents the core of the YATA algorithm.
type CRDT struct {
	Document *Document
}

// NewCRDT creates a new CRDT instance.
func NewCRDT() *CRDT {
	return &CRDT{
		Document: NewDocument(),
	}
}

// InsertBlock inserts a new block with conflict resolution.
func (crdt *CRDT) InsertBlock(clientID, clock uint32, content string, origin *ID) {
	block := NewBlock(clientID, clock, content, origin)
	txn := NewTransaction(crdt.Document)
	txn.AddChange(block)
	txn.Commit()
}

// PrintDocument prints the entire document content.
func (crdt *CRDT) PrintDocument() {
	crdt.Document.PrintDocument()
}
