package block

type ID struct {
	Clock  uint64
	Client uint64
}

type Block struct {
	ID        ID
	Content   string
	IsDeleted bool

	LeftOrigin  ID
	RightOrigin ID
	Left        *Block
	Right       *Block
}

// NewBlock creates a block with content and ID.
func NewBlock(id ID, content string) *Block {
	return &Block{
		ID:      id,
		Content: content,
	}
}

// AttachNeighbor links the block between left and right.
func (b *Block) AttachNeighbor(left, right *Block) {
	if left != nil {
		left.Right = b
	}
	if right != nil {
		right.Left = b
	}
	b.Left = left
	b.Right = right
}
