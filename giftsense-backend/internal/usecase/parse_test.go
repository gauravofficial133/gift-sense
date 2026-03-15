package usecase_test

import (
	"strings"
	"testing"

	"github.com/giftsense/backend/internal/domain"
	"github.com/giftsense/backend/internal/usecase"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func sampleWhatsAppText() string {
	return strings.Join([]string{
		"[15/01/2024, 10:00:00] Alice: Hey how are you?",
		"[15/01/2024, 10:01:00] Bob: I'm good! Was just thinking about pottery classes",
		"[15/01/2024, 10:02:00] Alice: Oh really? That sounds fun",
		"[15/01/2024, 10:03:00] Bob: Yeah I love creative things",
		"[15/01/2024, 10:04:00] Alice: Me too! <Media omitted>",
		"[15/01/2024, 10:05:00] Bob: Nice pic!",
	}, "\n")
}

func TestParseConversation_ShouldParseWhatsAppFormat_WhenValidExportProvided(t *testing.T) {
	msgs, err := usecase.ParseConversation(sampleWhatsAppText(), 400)
	require.NoError(t, err)
	assert.Len(t, msgs, 6)
	assert.Equal(t, "Alice", msgs[0].Sender)
	assert.Equal(t, "Hey how are you?", msgs[0].Text)
	assert.Equal(t, 0, msgs[0].Index)
}

func TestParseConversation_ShouldSetIsMedia_WhenMediaOmittedLineDetected(t *testing.T) {
	msgs, err := usecase.ParseConversation(sampleWhatsAppText(), 400)
	require.NoError(t, err)
	mediaMsg := msgs[4]
	assert.True(t, mediaMsg.IsMedia)
}

func TestParseConversation_ShouldFilterSystemMessages_WhenWhatsAppFormatDetected(t *testing.T) {
	text := strings.Join([]string{
		"[15/01/2024, 10:00:00] Alice: Hello",
		"Messages and calls are end-to-end encrypted. No one outside of this chat, not even WhatsApp, can read or listen to them.",
		"[15/01/2024, 10:01:00] Bob: Hi there",
		"[15/01/2024, 10:02:00] Alice: How are you?",
		"[15/01/2024, 10:03:00] Bob: Good!",
		"[15/01/2024, 10:04:00] Alice: Great!",
	}, "\n")
	msgs, err := usecase.ParseConversation(text, 400)
	require.NoError(t, err)
	for _, m := range msgs {
		assert.NotContains(t, m.Text, "end-to-end encrypted")
	}
}

func TestParseConversation_ShouldReturnError_WhenFewerThanFiveMessages(t *testing.T) {
	text := strings.Join([]string{
		"[15/01/2024, 10:00:00] Alice: Hi",
		"[15/01/2024, 10:01:00] Bob: Hey",
	}, "\n")
	_, err := usecase.ParseConversation(text, 400)
	require.Error(t, err)
	assert.ErrorIs(t, err, domain.ErrConversationTooShort)
}

func TestParseConversation_ShouldCapMessages_WhenExceedsMaxMessages(t *testing.T) {
	var lines []string
	for i := 0; i < 20; i++ {
		lines = append(lines, "[15/01/2024, 10:00:00] Alice: message")
		lines = append(lines, "[15/01/2024, 10:01:00] Bob: reply")
	}
	text := strings.Join(lines, "\n")
	msgs, err := usecase.ParseConversation(text, 10)
	require.NoError(t, err)
	assert.LessOrEqual(t, len(msgs), 10)
}

func TestParseConversation_ShouldHandlePlainText_WhenNoWhatsAppFormatDetected(t *testing.T) {
	text := strings.Join([]string{
		"You: Hey there",
		"Friend: Hi! How are you?",
		"You: Good thanks",
		"Friend: What are you up to?",
		"You: Just chilling",
		"Friend: Nice!",
	}, "\n")
	msgs, err := usecase.ParseConversation(text, 400)
	require.NoError(t, err)
	assert.GreaterOrEqual(t, len(msgs), 5)
}
