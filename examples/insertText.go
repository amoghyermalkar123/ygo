package examples

import (
	"fmt"

	"github.com/amoghyermalkar123/ygo"
)

func insertText() {
	// Create a new YDoc instance
	doc := ygo.NewYDoc()

	// Insert text into the document
	err := doc.InsertText(0, "Hello, World!")
	if err != nil {
		panic(err)
	}

	fmt.Println(doc.Content())
	// Hello, World!
}
