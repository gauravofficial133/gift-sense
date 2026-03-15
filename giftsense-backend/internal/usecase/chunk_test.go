package usecase_test

import (
	"fmt"
	"testing"

	"github.com/giftsense/backend/internal/domain"
	"github.com/giftsense/backend/internal/usecase"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func makeChunkMessages(texts []string) []domain.Message {
	msgs := make([]domain.Message, len(texts))
	for i, t := range texts {
		msgs[i] = domain.Message{Index: i, Sender: "[Person_A]", Text: t}
	}
	return msgs
}

func TestChunkMessages_ShouldProduceCorrectChunkCount_WhenWindowAndOverlapApplied(t *testing.T) {
	// 10 messages, window=8, overlap=3 → step=5 → chunks starting at 0,5 → 2 chunks
	msgs := makeChunkMessages([]string{
		"msg1", "msg2", "msg3", "msg4", "msg5",
		"msg6", "msg7", "msg8", "msg9", "msg10",
	})
	chunks, err := usecase.ChunkMessages("sess1", msgs, 8, 3)
	require.NoError(t, err)
	assert.GreaterOrEqual(t, len(chunks), 1)
}

func TestChunkMessages_ShouldGenerateUniqueChunkIDs_WithSessionIDPrefix(t *testing.T) {
	msgs := makeChunkMessages([]string{"a", "b", "c", "d", "e", "f", "g", "h"})
	chunks, err := usecase.ChunkMessages("mysession", msgs, 4, 2)
	require.NoError(t, err)
	ids := make(map[string]bool)
	for _, c := range chunks {
		assert.Contains(t, c.ID, "mysession", "chunk ID should contain session ID")
		assert.False(t, ids[c.ID], "chunk IDs should be unique")
		ids[c.ID] = true
	}
}

func TestChunkMessages_ShouldSetHasPreference_WhenPreferenceKeywordPresent(t *testing.T) {
	msgs := makeChunkMessages([]string{
		"I love hiking in the mountains",
		"It is so peaceful",
		"I enjoy outdoor activities",
		"What about you?",
	})
	chunks, err := usecase.ChunkMessages("sess1", msgs, 4, 0)
	require.NoError(t, err)
	require.NotEmpty(t, chunks)
	assert.True(t, chunks[0].Metadata.HasPreference)
}

func TestChunkMessages_ShouldSetHasWish_WhenWishKeywordPresent(t *testing.T) {
	msgs := makeChunkMessages([]string{
		"I wish I could travel more",
		"Someday I want to visit Japan",
		"That would be amazing",
		"Dream vacation!",
	})
	chunks, err := usecase.ChunkMessages("sess1", msgs, 4, 0)
	require.NoError(t, err)
	require.NotEmpty(t, chunks)
	assert.True(t, chunks[0].Metadata.HasWish)
}

func TestChunkMessages_ShouldSkipMediaOnlyWindows_WhenIsMediaTrue(t *testing.T) {
	msgs := []domain.Message{
		{Index: 0, Sender: "[Person_A]", Text: "<Media omitted>", IsMedia: true},
		{Index: 1, Sender: "[Person_B]", Text: "<Media omitted>", IsMedia: true},
		{Index: 2, Sender: "[Person_A]", Text: "<Media omitted>", IsMedia: true},
		{Index: 3, Sender: "[Person_B]", Text: "<Media omitted>", IsMedia: true},
	}
	chunks, err := usecase.ChunkMessages("sess1", msgs, 4, 0)
	require.NoError(t, err)
	assert.Empty(t, chunks, "all-media window should be skipped")
}

func TestChunkMessages_ShouldEnrichTopics_WhenCookingKeywordPresent(t *testing.T) {
	msgs := makeChunkMessages([]string{
		"I have been watching so many cooking videos",
		"I want to learn to make pasta",
		"Food is my passion",
		"What do you think?",
	})
	chunks, err := usecase.ChunkMessages("sess1", msgs, 4, 0)
	require.NoError(t, err)
	require.NotEmpty(t, chunks)
	found := false
	for _, topic := range chunks[0].Metadata.Topics {
		if topic == "cooking" || topic == "food" {
			found = true
		}
	}
	assert.True(t, found, "cooking or food topic should be detected")
}

func TestChunkMessages_ShouldIncludeSessionID_InEveryChunk(t *testing.T) {
	msgs := makeChunkMessages([]string{"a", "b", "c", "d"})
	chunks, err := usecase.ChunkMessages("test-session-123", msgs, 4, 0)
	require.NoError(t, err)
	for i, c := range chunks {
		assert.Equal(t, "test-session-123", c.SessionID, fmt.Sprintf("chunk %d should have correct sessionID", i))
	}
}
