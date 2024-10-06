package pkg

// BlockStore collection of blocks keyed by client id
type BlockStore struct {
	clientBlocks map[int64]*ClientBlockList
}

type ClientBlockList struct {
	clientBlockList []Block
}

type Block struct{}
