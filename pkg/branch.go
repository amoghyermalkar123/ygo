package pkg

// Branch describes a content of a complex Ygo data structure
type Branch struct {
	// pointer to a block type
	Start any
	// A map component of this branch node, used by some of the specialized complex types likes Maps or Arrays
	Map map[string]any

	// Item is a unique identifier of the current branch node
	// if its a string - this branch is a root level data structure
	// if its a block identifier - this block is a complex type i.e map/ array
	// TODO: figure out what block id is, whats a root level DS and whats a complex type
	Item any

	// For root-level types, this is a name of a branch.
	Name string

	// A length of an indexed sequence component of a current branch node.
	BlockLength int64

	ContentLength int64

	// An identifier of an underlying complex data type (eg. is it an Array or a Map).
	TypeRef any

	// TODO: this
	Observers any

	// TODO: this
	DeepObservers any
}
