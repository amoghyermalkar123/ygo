package pkg

// BlockStore collection of blocks keyed by client id
type BlockStore struct {
	clientBlocks map[int64]*ClientBlockList
}

func NewBlockStore(clientID ID) *BlockStore {
	b := &BlockStore{
		clientBlocks: make(map[int64]*ClientBlockList),
	}
	b.clientBlocks[clientID.ClientID] = &ClientBlockList{clientBlockList: []*Block{}}

	return b
}

type ClientBlockList struct {
	clientBlockList []*Block
}

func (c *ClientBlockList) First() *Block {
	return c.clientBlockList[0]
}

type Block struct{}
