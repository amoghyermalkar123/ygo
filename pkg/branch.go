package pkg

import "github.com/amoghyermalkar123/ygo/pkg/types"

// Branch describes a content of a complex Ygo data structure like array or map
type Branch struct {
	// A pointer to a first block of a indexed sequence component of this branch node.
	// If nil, sequence is empty
	Start *Block
	// A map component of this branch node, used by some of the specialized complex types likes Maps or Arrays
	Map map[string]*Block

	// Item is a unique identifier of the current branch node
	// if its a string - this branch is a root level data structure
	// if its a block identifier - this block is a complex type i.e map/ array
	// TODO: figure out how this field and name field works
	Item *Block

	// For root-level types, this is a name of a branch.
	Name string

	// A length of an indexed sequence component of a current branch node.
	BlockLength int64

	ContentLength int64

	// An identifier of an underlying complex data type (eg. is it an Array or a Map).
	TypeRef types.TypeRef

	// TODO: this
	Observers any

	// TODO: this
	DeepObservers any
}
