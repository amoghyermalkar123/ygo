### What is it?
A Text-based CRDT library in Golang

### What are CRDT's ?
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

It's inspired and loosely based on this wonderful project - https://github.com/yjs/yjs
and the academic paper by the authors that back the same project.

This is ofcourse NOT production grade. It's a result of more than a year spent on learning the fundamental math, distributed systems and nights spent reverse engineering the yjs/ yrs project.

The main project is extensive and highly sophisticated, my attempt tries to follow along the same path. They have a lot of datatypes while `ygo` only supports a text based CRDT.

You will find that i have added extensive comments throughout the codebase for newer folks to understand how the algorithm works so you can avoid what i had to do-
reverse engineer a rather nuanced codebase like Yjs.

### Current status:
I am working on a replay system so that a run of the algorithm can be visualized to better understand each steps and the subtelties that lie across the codebase.
