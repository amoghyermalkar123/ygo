package ygo

import "ygo/internal/blockstore"

type YDoc struct {
	// The document's content.
	blockStore *blockstore.BlockStore
}

func NewYDoc() *YDoc {
	return &YDoc{
		blockStore: blockstore.NewStore(),
	}
}

func (yd *YDoc) InsertText(pos uint64, text string) error {
	return yd.blockStore.Insert(pos, text)
}

func (yd *YDoc) Content() string {
	return yd.blockStore.Content()
}
