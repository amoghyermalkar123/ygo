package pkg

// BlockStore collection of blocks keyed by client id
type BlockStore struct {
	clientBlocks map[int64]*ClientBlockList
}

type ClientBlockList struct {
	blockCell []*Block
	gcCell    []*GC
}

type GC struct{}

type Block struct{}
