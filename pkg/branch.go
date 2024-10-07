package pkg

// Branch describes a content of a complex Ygo data structure like array or map
type Branch struct {
	// A pointer to a first block of a indexed sequence component of this branch node.
	// If nil, sequence is empty
	Start *Block

	// For root-level types, this is a name of a branch.
	Name string

	// A length of an indexed sequence component of a current branch node.
	BlockLength int64

	// length of the content that the underlying block store holds for the complex data type
	ContentLength int64
}

func NewBranch(firstBlock *Block, name string) *Branch {
	return &Branch{
		Start:         firstBlock,
		Name:          name,
		BlockLength:   0,
		ContentLength: 0,
	}
}
