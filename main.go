package main

import "github.com/amoghyermalkar123/ygo/research/yata"

func main() {
	crdt := yata.NewCRDT()

	// User 1 inserts Block A
	crdt.InsertBlock(1, 1, "Block A", nil)

	// User 2 inserts Block B after Block A
	crdt.InsertBlock(2, 1, "Block B", &yata.ID{Client: 1, Clock: 1})

	// User 1 inserts Block C after Block B
	crdt.InsertBlock(1, 2, "Block C", &yata.ID{Client: 2, Clock: 1})

	// Print final document
	crdt.PrintDocument()
}
