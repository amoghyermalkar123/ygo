package block

type ID struct {
	Clock  int64
	Client int64
}

type Block struct {
	ID          ID
	Content     string
	IsDeleted   bool
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

func (b *Block) MarkDeleted() {
	b.IsDeleted = true
	b.Content = ""
}

type BlockTextListPosition struct {
	Left  *Block
	Right *Block
	Index int64
}

func (l *BlockTextListPosition) Forward() {
	if !l.Right.IsDeleted {
		l.Index += int64(len(l.Right.Content))
	}
	l.Left = l.Right
	l.Right = l.Right.Right
}
