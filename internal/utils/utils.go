package utils

import "github.com/amoghyermalkar123/ygo/internal/block"

func EqualID(a, b block.ID) bool {
	return a.Client == b.Client && a.Clock == b.Clock
}

func EqualIDPtr(a, b *block.ID) bool {
	if a == nil || b == nil {
		return false
	}
	return EqualID(*a, *b)
}
