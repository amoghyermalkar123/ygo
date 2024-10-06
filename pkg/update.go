package pkg

type PendingUpdate struct {
	Update Update
}

type Update struct {
	clientBlocks map[int64]*ClientBlockList
}
