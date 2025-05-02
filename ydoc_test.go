package ygo_test

import (
	"fmt"
	"testing"

	"ygo"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestApplyUpdate_BasicSynchronization tests that simple updates synchronize between docs
func TestApplyUpdate_BasicSynchronization(t *testing.T) {
	// Create two docs
	source := ygo.NewYDoc()
	target := ygo.NewYDoc()

	// Make changes to source doc
	err := source.InsertText(0, "Hello World")
	require.NoError(t, err)

	// Encode state as update
	update, err := source.EncodeStateAsUpdate()
	require.NoError(t, err)

	// Apply to target doc
	err = target.ApplyUpdate(update)
	require.NoError(t, err)

	// Verify content is synchronized
	assert.Equal(t, "Hello World", target.Content())
	assert.Equal(t, source.Content(), target.Content())
}

// TestApplyUpdate_SequentialEdits tests that sequential edits synchronize correctly
func TestApplyUpdate_SequentialEdits(t *testing.T) {
	// Create two docs
	source := ygo.NewYDoc()
	target := ygo.NewYDoc()

	// Initial content
	err := source.InsertText(0, "Hello")
	require.NoError(t, err)

	// Sync initial state
	update1, err := source.EncodeStateAsUpdate()
	require.NoError(t, err)
	err = target.ApplyUpdate(update1)
	require.NoError(t, err)

	// Add more content
	err = source.InsertText(5, " World")
	require.NoError(t, err)

	// Sync again
	update2, err := source.EncodeStateAsUpdate()
	require.NoError(t, err)
	err = target.ApplyUpdate(update2)
	require.NoError(t, err)

	// Verify content
	assert.Equal(t, "Hello World", target.Content())
}

// TestApplyUpdate_ConcurrentInsertions tests handling of concurrent insertions at the same position
func TestApplyUpdate_ConcurrentInsertions(t *testing.T) {
	// Create three docs
	doc1 := ygo.NewYDoc()
	doc2 := ygo.NewYDoc()
	doc3 := ygo.NewYDoc()

	// Initial content in all docs
	err := doc1.InsertText(0, "Hello World")
	require.NoError(t, err)

	// Sync to other docs
	update1, err := doc1.EncodeStateAsUpdate()
	require.NoError(t, err)

	err = doc2.ApplyUpdate(update1)
	require.NoError(t, err)

	err = doc3.ApplyUpdate(update1)
	require.NoError(t, err)

	// Make concurrent changes at the same position
	// Doc2 inserts " Beautiful" after "Hello"
	err = doc2.InsertText(5, " Beautiful")
	require.NoError(t, err)
	assert.Equal(t, "Hello Beautiful World", doc2.Content())

	// Doc3 inserts " Amazing" after "Hello"
	err = doc3.InsertText(5, " Amazing")
	require.NoError(t, err)
	assert.Equal(t, "Hello Amazing World", doc3.Content())

	// Create updates from both docs
	update2, err := doc2.EncodeStateAsUpdate()
	require.NoError(t, err)

	update3, err := doc3.EncodeStateAsUpdate()
	require.NoError(t, err)

	// Apply both updates to doc1
	err = doc1.ApplyUpdate(update2)
	require.NoError(t, err)
	assert.Equal(t, "Hello Beautiful World", doc1.Content())

	err = doc1.ApplyUpdate(update3)
	require.NoError(t, err)
	if doc2.Client() < doc3.Client() {
		assert.Equal(t, "Hello Beautiful Amazing World", doc1.Content())
	}
	if doc3.Client() < doc2.Client() {
		assert.Equal(t, "Hello Amazing Beautiful World", doc1.Content())
	}

}

// TestApplyUpdate_Deletions tests synchronization of deletions
func TestApplyUpdate_Deletions(t *testing.T) {
	// Create two docs
	source := ygo.NewYDoc()
	target := ygo.NewYDoc()

	// Initial content
	err := source.InsertText(0, "Hello World")
	require.NoError(t, err)

	// Sync initial state
	update1, err := source.EncodeStateAsUpdate()
	require.NoError(t, err)
	err = target.ApplyUpdate(update1)
	require.NoError(t, err)

	// Delete "World" from source
	err = source.DeleteText(6, 5)
	require.NoError(t, err)

	// Verify source content
	assert.Equal(t, "Hello ", source.Content())

	// Sync deletion to target
	update2, err := source.EncodeStateAsUpdate()
	require.NoError(t, err)
	err = target.ApplyUpdate(update2)
	require.NoError(t, err)

	// Verify target content after sync
	assert.Equal(t, "Hello ", target.Content())
}

// TestApplyUpdate_InsertionsDeletionsMixed tests mixing of insertions and deletions
func TestApplyUpdate_InsertionsDeletionsMixed(t *testing.T) {
	// Create two docs
	doc1 := ygo.NewYDoc()
	doc2 := ygo.NewYDoc()

	// Initial content in doc1
	err := doc1.InsertText(0, "ABCDEF")
	require.NoError(t, err)

	// Sync to doc2
	update1, err := doc1.EncodeStateAsUpdate()
	require.NoError(t, err)
	err = doc2.ApplyUpdate(update1)
	require.NoError(t, err)

	// Doc1: Delete "CD"
	err = doc1.DeleteText(2, 2)
	require.NoError(t, err)
	fmt.Println("doc 1 content after deletion:", doc1.Content())

	// Doc2: Insert "XY" between "AB" and "CDEF"
	err = doc2.InsertText(2, "XY")
	require.NoError(t, err)

	// Create and apply updates in both directions
	update1to2, err := doc1.EncodeStateAsUpdate()
	require.NoError(t, err)

	update2to1, err := doc2.EncodeStateAsUpdate()
	require.NoError(t, err)

	// Apply updates
	err = doc2.ApplyUpdate(update1to2)
	require.NoError(t, err)

	err = doc1.ApplyUpdate(update2to1)
	require.NoError(t, err)

	// Verify both docs converged to the same state
	assert.Equal(t, doc1.Content(), doc2.Content())

	// The expected result depends on CRDT implementation details,
	// but both "XY" should be present and "CD" should be deleted
	content := doc1.Content()
	assert.Contains(t, content, "AB")
	assert.Contains(t, content, "XY")
	assert.Contains(t, content, "EF")
	assert.NotContains(t, content, "CD")
}

// TestApplyUpdate_EmptyUpdate tests applying an empty update
func TestApplyUpdate_EmptyUpdate(t *testing.T) {
	// Create an empty doc for the source
	empty := ygo.NewYDoc()

	// Create a doc with content as the target
	target := ygo.NewYDoc()
	err := target.InsertText(0, "Hello")
	require.NoError(t, err)

	// Encode empty state and apply to target
	update, err := empty.EncodeStateAsUpdate()
	require.NoError(t, err)

	err = target.ApplyUpdate(update)
	require.NoError(t, err)

	// Verify target content is unchanged
	assert.Equal(t, "Hello", target.Content())
}

// TestApplyUpdate_LargeDocuments tests synchronization with larger documents
func TestApplyUpdate_LargeDocuments(t *testing.T) {
	// Create two docs
	source := ygo.NewYDoc()
	target := ygo.NewYDoc()

	// Create a larger document with multiple operations
	err := source.InsertText(0, "aaaaaaaaaaaaaaaaa")
	require.NoError(t, err)

	err = source.InsertText(18, "bbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbb bbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbb")
	require.NoError(t, err)

	err = source.InsertText(90, "ccccccccccccccccccccccccccccccccccc")
	require.NoError(t, err)

	// Sync to target
	update, err := source.EncodeStateAsUpdate()
	require.NoError(t, err)

	err = target.ApplyUpdate(update)
	require.NoError(t, err)

	// Verify content
	assert.Equal(t, source.Content(), target.Content())
}

func TestApplyUpdate_WeDontCareAboutIndex(t *testing.T) {
	// Create two docs
	source := ygo.NewYDoc()
	target := ygo.NewYDoc()

	// Create a larger document with multiple operations
	err := source.InsertText(0, "aaaaaaaaaaaaaaaaa")
	require.NoError(t, err)

	err = source.InsertText(11118, "bbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbb bbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbb")
	require.NoError(t, err)

	err = source.InsertText(90000, "ccccccccccccccccccccccccccccccccccc")
	require.NoError(t, err)

	// Sync to target
	update, err := source.EncodeStateAsUpdate()
	require.NoError(t, err)

	err = target.ApplyUpdate(update)
	require.NoError(t, err)

	// Verify content
	assert.Equal(t, source.Content(), target.Content())
}

// TestApplyUpdate_SplitUpdates tests applying updates in pieces
func TestApplyUpdate_SplitUpdates(t *testing.T) {
	// Create docs
	doc1 := ygo.NewYDoc()
	doc2 := ygo.NewYDoc()
	doc3 := ygo.NewYDoc()

	// Make changes to all docs
	err := doc1.InsertText(0, "Hello")
	require.NoError(t, err)

	err = doc2.InsertText(0, "World")
	require.NoError(t, err)

	err = doc3.InsertText(0, "Testing")
	require.NoError(t, err)

	// Generate updates from each doc
	update1, err := doc1.EncodeStateAsUpdate()
	require.NoError(t, err)

	update2, err := doc2.EncodeStateAsUpdate()
	require.NoError(t, err)

	update3, err := doc3.EncodeStateAsUpdate()
	require.NoError(t, err)

	// Apply all updates to doc1
	err = doc1.ApplyUpdate(update2)
	require.NoError(t, err)

	err = doc1.ApplyUpdate(update3)
	require.NoError(t, err)

	// Apply all updates to doc2
	err = doc2.ApplyUpdate(update1)
	require.NoError(t, err)

	err = doc2.ApplyUpdate(update3)
	require.NoError(t, err)

	// Apply all updates to doc3
	err = doc3.ApplyUpdate(update1)
	require.NoError(t, err)

	err = doc3.ApplyUpdate(update2)
	require.NoError(t, err)

	// Verify all docs converged to the same content
	assert.Equal(t, doc1.Content(), doc2.Content())
	assert.Equal(t, doc1.Content(), doc3.Content())

	// Content should contain all three inserts
	content := doc1.Content()
	assert.Contains(t, content, "Hello")
	assert.Contains(t, content, "World")
	assert.Contains(t, content, "Testing")
}

// TestApplyUpdate_ConflictResolution tests conflict resolution with explicit client IDs
func TestApplyUpdate_ConflictResolution(t *testing.T) {
	// Create docs with explicit, different client IDs to ensure deterministic conflict resolution
	doc1 := ygo.NewYDoc() // Will have the lowest client ID
	doc2 := ygo.NewYDoc() // Will have a higher client ID

	// Print the client IDs for debugging purposes
	t.Logf("Doc1 client ID: %d", doc1.Client())
	t.Logf("Doc2 client ID: %d", doc2.Client())

	// Initial content
	err := doc1.InsertText(0, "A")
	require.NoError(t, err)

	// Sync to doc2
	update1, err := doc1.EncodeStateAsUpdate()
	require.NoError(t, err)
	err = doc2.ApplyUpdate(update1)
	require.NoError(t, err)

	// Both docs make concurrent insertions at the same position
	// Doc1 inserts "B" after "A"
	err = doc1.InsertText(1, "B")
	require.NoError(t, err)

	// Doc2 inserts "C" after "A" (same position)
	err = doc2.InsertText(1, "C")
	require.NoError(t, err)

	// Generate updates
	update1to2, err := doc1.EncodeStateAsUpdate()
	require.NoError(t, err)

	update2to1, err := doc2.EncodeStateAsUpdate()
	require.NoError(t, err)

	// Apply updates in both directions
	err = doc2.ApplyUpdate(update1to2)
	require.NoError(t, err)

	err = doc1.ApplyUpdate(update2to1)
	require.NoError(t, err)

	// Verify both docs have the same content
	assert.Equal(t, doc1.Content(), doc2.Content())

	// The doc with the lower client ID should have its content come first
	// This tests the specific CRDT conflict resolution algorithm
	if doc1.Client() < doc2.Client() {
		assert.Equal(t, "ABC", doc1.Content())
	} else {
		assert.Equal(t, "ACB", doc1.Content())
	}
}

// TestApplyUpdate_IdempotentUpdates tests that applying the same update multiple times is idempotent
func TestApplyUpdate_IdempotentUpdates(t *testing.T) {
	// Create docs
	sourceDoc := ygo.NewYDoc()
	targetDoc := ygo.NewYDoc()

	// Make changes to source
	err := sourceDoc.InsertText(0, "Hello World")
	require.NoError(t, err)

	// Create update
	update, err := sourceDoc.EncodeStateAsUpdate()
	require.NoError(t, err)

	// Apply update once
	err = targetDoc.ApplyUpdate(update)
	require.NoError(t, err)

	// Verify content
	assert.Equal(t, "Hello World", targetDoc.Content())

	// Save the content
	contentAfterFirstUpdate := targetDoc.Content()

	// Apply the same update again
	err = targetDoc.ApplyUpdate(update)
	require.NoError(t, err)

	// Apply a third time
	err = targetDoc.ApplyUpdate(update)
	require.NoError(t, err)

	// Verify content hasn't changed
	assert.Equal(t, contentAfterFirstUpdate, targetDoc.Content())
}

// TestApplyUpdate_EdgeCaseEmptyDocument tests applying updates to an empty document
func TestApplyUpdate_EdgeCaseEmptyDocument(t *testing.T) {
	// Create an empty doc
	emptyDoc := ygo.NewYDoc()

	// Create a doc with content
	contentDoc := ygo.NewYDoc()
	err := contentDoc.InsertText(0, "Content")
	require.NoError(t, err)

	// Generate update from content doc
	update, err := contentDoc.EncodeStateAsUpdate()
	require.NoError(t, err)

	// Apply to empty doc
	err = emptyDoc.ApplyUpdate(update)
	require.NoError(t, err)

	// Verify empty doc now has content
	assert.Equal(t, "Content", emptyDoc.Content())
}

// TestApplyUpdate_EdgeCaseNilUpdate tests handling of nil updates
func TestApplyUpdate_EdgeCaseNilUpdate(t *testing.T) {
	// Create a doc
	doc := ygo.NewYDoc()
	err := doc.InsertText(0, "Content")
	require.NoError(t, err)

	// Try to apply nil update
	err = doc.ApplyUpdate(nil)

	// Should return an error
	assert.Error(t, err)

	// Document should be unchanged
	assert.Equal(t, "Content", doc.Content())
}

// TestApplyUpdate_EdgeCaseInvalidUpdate tests handling of invalid updates
func TestApplyUpdate_EdgeCaseInvalidUpdate(t *testing.T) {
	// Create a doc
	doc := ygo.NewYDoc()
	err := doc.InsertText(0, "Content")
	require.NoError(t, err)

	// Create invalid update data
	invalidUpdate := []byte("{not valid json}")

	// Try to apply invalid update
	err = doc.ApplyUpdate(invalidUpdate)

	// Should return an error
	assert.Error(t, err)

	// Document should be unchanged
	assert.Equal(t, "Content", doc.Content())
}

// TestApplyUpdate_MergeMultipleUpdates tests merging multiple updates sequentially
func TestApplyUpdate_MergeMultipleUpdates(t *testing.T) {
	// Create three docs to generate independent updates
	doc1 := ygo.NewYDoc()
	doc2 := ygo.NewYDoc()
	doc3 := ygo.NewYDoc()
	merged := ygo.NewYDoc()

	// Make different changes to each doc
	err := doc1.InsertText(0, "Hello ")
	require.NoError(t, err)

	err = doc2.InsertText(0, "World ")
	require.NoError(t, err)

	err = doc3.InsertText(0, "Testing!")
	require.NoError(t, err)

	// Generate updates
	update1, err := doc1.EncodeStateAsUpdate()
	require.NoError(t, err)

	update2, err := doc2.EncodeStateAsUpdate()
	require.NoError(t, err)

	update3, err := doc3.EncodeStateAsUpdate()
	require.NoError(t, err)

	// Apply all updates to the merged doc
	err = merged.ApplyUpdate(update1)
	require.NoError(t, err)

	err = merged.ApplyUpdate(update2)
	require.NoError(t, err)

	err = merged.ApplyUpdate(update3)
	require.NoError(t, err)

	// Verify content contains all three parts
	content := merged.Content()
	assert.Contains(t, content, "Hello")
	assert.Contains(t, content, "World")
	assert.Contains(t, content, "Testing!")
}

// TestMultipleDeletionSynchronization tests synchronizing multiple deletion operations
func TestMultipleDeletionSynchronization(t *testing.T) {
	// Create two docs
	source := ygo.NewYDoc()
	target := ygo.NewYDoc()

	// Add initial content to source
	err := source.InsertText(0, "The quick brown fox jumps over the lazy dog")
	require.NoError(t, err)

	// Synchronize the initial state to target
	update1, err := source.EncodeStateAsUpdate()
	require.NoError(t, err)
	err = target.ApplyUpdate(update1)
	require.NoError(t, err)

	// Perform multiple deletions in source
	err = source.DeleteText(4, 6) // Delete "quick "
	require.NoError(t, err)
	err = source.DeleteText(10, 4) // Delete "fox "
	require.NoError(t, err)
	err = source.DeleteText(20, 9) // Delete "the lazy "
	require.NoError(t, err)

	// Verify source content
	assert.Equal(t, "The brown jumps over dog", source.Content())

	// Create update with deletions
	update2, err := source.EncodeStateAsUpdate()
	require.NoError(t, err)

	// Apply to target
	err = target.ApplyUpdate(update2)
	require.NoError(t, err)

	// Verify target content now matches source
	assert.Equal(t, source.Content(), target.Content())
}
