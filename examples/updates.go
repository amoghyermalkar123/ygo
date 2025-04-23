package examples

import (
	"ygo"
)

func update() {
	// Create a new YDoc instance
	localDoc := ygo.NewYDoc()

	// Insert text into the document
	err := localDoc.InsertText(0, "Hello, World!")
	if err != nil {
		panic(err)
	}

	localDoc.ApplyUpdate([]byte("update data")) // This should be a valid update byte array
}
