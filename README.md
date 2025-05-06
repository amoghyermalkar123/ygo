### What even are CRDT's ?
CRDT's or Conflict-Free Replicated Data Types are a novel algorithm or a family of data types that have 3 mathematical properties:
- Commutativity
- Associativity
- Idempotence

It's a common practise that such data types fulfill these properties in-order to be categorized as one of the following:
- Convergent
- Commutative

Ygo stands for Yata-Go based on the original paper where the algorithm is called - Yet Another Transformation Approach

This CRDT though is a hybrid of both Convergent and Commutative CRDT's.
- Updates are Commutative
- Deletes are Convergent

### What problem do CRDT's solve?
The core idea behind these data types is that they don't require any co-ordination. Think of a typical async network scenario where a node has some data other nodes in the network might be interested in. The way you would model that is with a centralized process that co-ordinates stuff and orchestrates correctness of data i.e. a the messages should be delivered in order they were sent. In CRDT-land, the same use case would follow but without a centralized process. Each node in a network can be thought of as self-healing, in the sense that they can get messages from other nodes in any order, after any amount of time and still at the end, each node in the network, will converge to the same state. Regardless of how fragile network is, CRDT's have the property to always endup with the same state.

By this point you would've figure out one thing. They are not good in scenarios that require strong consistency. Today, they are used in collaborative or local-first software. Where consistency is desired eventually and not real-time.

### Current status of the project:
This library will be a learning project and focusing on text-based CRDT for now. Future roadmap is still TBD.

### Examples:
Say in one process you do this:
```go
	// Create a new YDoc instance
	doc := ygo.NewYDoc()

	// Insert text into the document
	err := doc.InsertText(0, "Hello, World!")
	if err != nil {
		panic(err)
	}

	fmt.Println(doc.Content())
	// Hello, World!
```

And in another process:
```go
	// Create a new YDoc instance
	doc := ygo.NewYDoc()

	// Insert text into the document
	err := doc.InsertText(0, "Hello, Amazing World!")
	if err != nil {
		panic(err)
	}

	fmt.Println(doc.Content())
	// Hello, World!
```

Eventually when you do this on both nodes:

```go
localDoc.ApplyUpdate(// json data)
```

Both of the nodes will end up with the same data, in this case -
```
Hello, Amazing World
```

CRDT's are protocol and network agnostic so we can use HTTP, QUIC, raw TCP socket, Websockets, anything really as long as we can connect to a process on the Internet.
