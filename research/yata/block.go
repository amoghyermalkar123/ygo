package yata

import (
	"fmt"
)

// ID represents the unique identifier for a block.
type ID struct {
	Client uint32
	Clock  uint32
}

// Block represents a block of content in the document.
type Block struct {
	ID       ID
	Left     *Block // Left neighbor
	Right    *Block // Right neighbor
	Content  string
	Deleted  bool
	Origin   *ID
	ClientID uint32
}

// NewBlock creates a new block with content and clientID.
func NewBlock(clientID, clock uint32, content string, origin *ID) *Block {
	return &Block{
		ID:       ID{Client: clientID, Clock: clock},
		Content:  content,
		Origin:   origin,
		ClientID: clientID,
	}
}

// String returns a string representation of the block.
func (b *Block) String() string {
	return fmt.Sprintf("Block(%d#%d): %s", b.ID.Client, b.ID.Clock, b.Content)
}
