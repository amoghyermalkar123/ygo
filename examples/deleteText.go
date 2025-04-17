package examples

import (
	"fmt"
	"ygo"
)

func deleteText() {
	// Create a new YDoc instance
	doc := ygo.NewYDoc()

	// Insert text into the document
	err := doc.InsertText(0, "Hello, World!")
	if err != nil {
		panic(err)
	}

	fmt.Println(doc.Content())
	// Hello, World!

	// Delete text from the document
	err = doc.DeleteText(0, 5)
	if err != nil {
		panic(err)
	}
	fmt.Println(doc.Content())
	// World!
}
