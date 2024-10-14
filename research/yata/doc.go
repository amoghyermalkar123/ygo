package yata

import (
	"fmt"
)

// Document represents the shared document.
type Document struct {
	Blocks      []*Block
	StateVector StateVector
}

// NewDocument creates a new empty document.
func NewDocument() *Document {
	return &Document{
		Blocks:      []*Block{},
		StateVector: NewStateVector(),
	}
}

// Integrate inserts a block into the document, handling conflict resolution.
func (doc *Document) Integrate(block *Block) {
	fmt.Println("Integrating block:", block)

	// Find position based on origin
	var left *Block
	var right *Block

	for _, b := range doc.Blocks {
		if block.Origin != nil && b.ID == *block.Origin {
			left = b
			right = b.Right
		}
	}

	// Conflict resolution by clientID
	if right != nil && right.ID.Client < block.ID.Client {
		left = right
		right = right.Right
	}

	// Insert block
	block.Left = left
	block.Right = right

	if left != nil {
		left.Right = block
	}
	if right != nil {
		right.Left = block
	}

	doc.Blocks = append(doc.Blocks, block)

	// Update state vector
	doc.StateVector.UpdateState(block.ClientID, block.ID.Clock)
}

// PrintDocument prints the entire document content.
func (doc *Document) PrintDocument() {
	for _, block := range doc.Blocks {
		if !block.Deleted {
			fmt.Println(block)
		}
	}
}
