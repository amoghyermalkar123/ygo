### What is it?
A Text-based CRDT library in Golang

### What are CRDT's ?
A novel algorithm or a family of data types that have 3 mathematical properties:
- Commutativity
- Associativity
- Idempotence

It's a common practise that such data types fulfill these properties in-order to be what we call as either
- Convergent

Or

- Commutative

It's inspired and loosely based on this wonderful project - https://github.com/yjs/yjs
and the academic paper by the authors that back the same project.

This is ofcourse NOT production grade. It's a result of more than a year spent on learning the fundamental math, distributed systems and nights spent reverse engineering the yjs/ yrs project.

The main project is extensive and highly sophisticated, my attempt tries to follow along the same path. They have a lot of datatypes while `ygo` only supports a text based CRDT.
