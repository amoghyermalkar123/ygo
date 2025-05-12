# YGo: A Text-based Conflict-Free Replicated Data Type in Go

YGo is a lightweight, efficient implementation of a text-based Conflict-Free Replicated Data Type (CRDT) in Go. Based on the Yata algorithm ("Yet Another Transformation Approach"), YGo enables real-time collaborative editing with strong eventual consistency guarantees without requiring a central coordinator.

🚀 Features
- **Collaborative Text Editing:** Multiple users can edit the same document concurrently
- *Conflict-Free Resolution:* Automatic handling of conflicting edits
- *Offline-First Support:* Work offline and synchronize changes when reconnected
- *Network Agnostic:* Use with any transport layer (WebSockets, HTTP, QUIC, etc.)
- *Lightweight:* Minimal dependencies and efficient memory usage
- *Fully Tested:* Comprehensive test suite ensuring reliability

📚 What are CRDTs?
CRDTs (Conflict-Free Replicated Data Types) are a family of data structures that enable multiple processes to independently update shared data without coordination, while ensuring that all replicas eventually converge to the same state.
YGo implements a text CRDT with three key mathematical properties:

- *Commutativity:* The order of operations doesn't affect the final result
- *Associativity:* Grouping of operations doesn't affect the final result
- *Idempotence:* Applying the same operation multiple times doesn't change the result

This makes YGo ideal for collaborative applications where:

- Network connections may be unreliable
- Users need to work offline
- Real-time collaboration is required without strict coordination

🔧 Installation
```go get github.com/amoghyermalkar123/ygo```

📝 Usage
Basic Example
```go

// Create a new collaborative document
doc := ygo.NewYDoc()

// Insert text
err := doc.InsertText(0, "Hello, World!")
if err != nil {
    // Handle error
}

// Get current document content
fmt.Println(doc.Content()) // Output: Hello, World!

// Delete some text
err = doc.DeleteText(7, 5) // Delete "World"
if err != nil {
    // Handle error
}

fmt.Println(doc.Content()) // Output: Hello, !
```

Synchronizing Documents
```go
// Document 1
docA := ygo.NewYDoc()
docA.InsertText(0, "Hello, collaborative world!")

// Encode state as update
update, err := docA.EncodeStateAsUpdate()
if err != nil {
    // Handle error
}

// Document 2 (could be on a different machine)
docB := ygo.NewYDoc()

// Apply update from Document 1
err = docB.ApplyUpdate(update)
if err != nil {
    // Handle error
}

// Both documents now have the same content
fmt.Println(docB.Content()) // Output: Hello, collaborative world!
```

Handling Concurrent Edits
```go
// Initial document
source := ygo.NewYDoc()
source.InsertText(0, "Hello World")

// Create update to share
update1, _ := source.EncodeStateAsUpdate()

// Two clients receive the initial state
client1 := ygo.NewYDoc()
client2 := ygo.NewYDoc()
client1.ApplyUpdate(update1)
client2.ApplyUpdate(update1)

// Client 1 makes an edit
client1.InsertText(6, "beautiful ")

// Client 2 makes a different edit at the same position
client2.InsertText(6, "amazing ")

// Generate updates from both clients
update2, _ := client1.EncodeStateAsUpdate()
update3, _ := client2.EncodeStateAsUpdate()

// Apply both updates to both clients
client1.ApplyUpdate(update3)
client2.ApplyUpdate(update2)

// Both clients now have identical content
// The exact order depends on client IDs for conflict resolution
fmt.Println(client1.Content()) // Both clients have the same content
fmt.Println(client2.Content()) // with both edits integrated
```

🏗️ Architecture
YGo consists of several core components:

- YDoc: The main document interface that users interact with
- BlockStore: The underlying data structure that maintains blocks of text
- Block: The basic unit of text storage with metadata for CRDT operations
- MarkerSystem: Manages insertion positions throughout the document

🛣️ Roadmap

- Performance optimizations for large documents
- Additional CRDT data types (arrays, maps, counters)
- Network integration examples
- Developer tools and visualizations
- Interoperability with other CRDT implementations

📄 License
This project is licensed under the MIT License - see the LICENSE file for details.
📚 Learn More About CRDTs

A Comprehensive Study of CRDTs
- Paper this library is based on: https://www.researchgate.net/publication/310212186_Near_Real-Time_Peer-to-Peer_Shared_Editing_on_Extensible_Data_Types
- [CRDT Website](https://crdt.tech/): to learn about CRDT's in depth
- I recommend reading this entire [blog series](https://www.bartoszsypytkowski.com/the-state-of-a-state-based-crdts/) by Bartosz (maintainer of the y-crdt project) 

🙏 Acknowledgements
This implementation draws inspiration from:
- ([Yjs](https://github.com/yjs/yjs))
