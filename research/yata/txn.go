package yata

// Transaction represents an atomic operation on the document.
type Transaction struct {
	Document *Document
	Changes  []*Block
}

// NewTransaction creates a new transaction for a document.
func NewTransaction(doc *Document) *Transaction {
	return &Transaction{
		Document: doc,
	}
}

// AddChange adds a block change to the transaction.
func (txn *Transaction) AddChange(block *Block) {
	txn.Changes = append(txn.Changes, block)
}

// Commit applies the transaction to the document.
func (txn *Transaction) Commit() {
	for _, block := range txn.Changes {
		txn.Document.Integrate(block)
	}
}
