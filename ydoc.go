package ygo

import "ygo/internal/blockstore"

type YDoc struct {
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

func (yd *YDoc) DeleteText(pos, length uint64) error {
	return yd.blockStore.DeleteText(pos, length)
}

func (yd *YDoc) Content() string {
	return yd.blockStore.Content()
}
