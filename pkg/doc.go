package pkg

import "github.com/amoghyermalkar123/ygo/pkg/types"

type Doc struct {
	store *Store
}

func NewDocument(complexType types.TypeRef) *Doc {
	return &Doc{
		store: NewStore(ID{
			ClientID: 1,
			Clock:    0,
		}),
	}
}
