package usecase_test

import (
	"testing"

	"github.com/giftsense/backend/internal/domain"
	"github.com/giftsense/backend/internal/usecase"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func makeMessages(pairs [][2]string) []domain.Message {
	msgs := make([]domain.Message, len(pairs))
	for i, p := range pairs {
		msgs[i] = domain.Message{Index: i, Sender: p[0], Text: p[1]}
	}
	return msgs
}

func TestAnonymizeMessages_ShouldReplaceAllSenderNames_WhenMessagesProvided(t *testing.T) {
	msgs := makeMessages([][2]string{
		{"Priya", "Hey how are you?"},
		{"Riya", "I am good!"},
	})
	anon, _, err := usecase.AnonymizeMessages(msgs)
	require.NoError(t, err)
	for _, m := range anon {
		assert.NotEqual(t, "Priya", m.Sender)
		assert.NotEqual(t, "Riya", m.Sender)
	}
}

func TestAnonymizeMessages_ShouldAssignPersonA_ToFirstSender(t *testing.T) {
	msgs := makeMessages([][2]string{
		{"Alice", "Hello"},
		{"Bob", "Hi"},
	})
	anon, tokenMap, err := usecase.AnonymizeMessages(msgs)
	require.NoError(t, err)
	assert.Equal(t, "[Person_A]", anon[0].Sender)
	assert.Equal(t, "[Person_A]", tokenMap["Alice"])
}

func TestAnonymizeMessages_ShouldAssignPersonB_ToSecondSender(t *testing.T) {
	msgs := makeMessages([][2]string{
		{"Alice", "Hello"},
		{"Bob", "Hi"},
	})
	anon, tokenMap, err := usecase.AnonymizeMessages(msgs)
	require.NoError(t, err)
	assert.Equal(t, "[Person_B]", anon[1].Sender)
	assert.Equal(t, "[Person_B]", tokenMap["Bob"])
}

func TestAnonymizeMessages_ShouldUseStableTokens_WhenSameNameAppearsMultipleTimes(t *testing.T) {
	msgs := makeMessages([][2]string{
		{"Alice", "Hi Bob"},
		{"Bob", "Hey Alice"},
		{"Alice", "Alice is here"},
	})
	anon, _, err := usecase.AnonymizeMessages(msgs)
	require.NoError(t, err)
	assert.Equal(t, "[Person_A]", anon[0].Sender)
	assert.Equal(t, "[Person_A]", anon[2].Sender)
	assert.Equal(t, "[Person_B]", anon[1].Sender)
}

func TestAnonymizeMessages_ShouldReplaceNamesInMessageBodies_WhenSenderNamesAppear(t *testing.T) {
	msgs := makeMessages([][2]string{
		{"Alice", "Bob said he wants to go hiking"},
		{"Bob", "Yes Alice I do!"},
	})
	anon, _, err := usecase.AnonymizeMessages(msgs)
	require.NoError(t, err)
	assert.NotContains(t, anon[0].Text, "Bob")
	assert.NotContains(t, anon[1].Text, "Alice")
}

func TestAnonymizeMessages_ShouldNotModifyNonPIIText_WhenNoNamesPresent(t *testing.T) {
	msgs := makeMessages([][2]string{
		{"Alice", "I love hiking and cooking"},
		{"Bob", "Me too, great hobbies!"},
	})
	anon, _, err := usecase.AnonymizeMessages(msgs)
	require.NoError(t, err)
	assert.Contains(t, anon[0].Text, "hiking")
	assert.Contains(t, anon[0].Text, "cooking")
	assert.Contains(t, anon[1].Text, "great hobbies")
}
