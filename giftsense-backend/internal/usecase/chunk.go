package usecase

import (
	"fmt"
	"strings"

	"github.com/giftsense/backend/internal/domain"
)

var topicKeywords = map[string][]string{
	"cooking":  {"cooking", "cook", "recipe", "pasta", "bake", "baking", "kitchen"},
	"food":     {"food", "eat", "restaurant", "meal", "hungry", "zomato", "swiggy", "dinner", "lunch"},
	"travel":   {"travel", "trip", "visit", "vacation", "holiday", "flight", "hotel", "tour"},
	"reading":  {"book", "reading", "read", "novel", "author", "library"},
	"craft":    {"craft", "pottery", "art", "paint", "watercolor", "draw", "sketch", "clay"},
	"hobby":    {"hobby", "hobbies", "interest", "passionate", "passion"},
	"music":    {"music", "song", "guitar", "sing", "concert", "playlist", "album"},
	"sport":    {"sport", "gym", "fitness", "workout", "run", "running", "yoga", "swim"},
	"outdoor":  {"outdoor", "hiking", "trek", "trekking", "nature", "camping"},
	"creative": {"creative", "create", "design", "artist", "diy"},
}

var emotionKeywords = map[string][]string{
	"humor":      {"haha", "lol", "lmao", "funny", "joke", "laugh", "hilarious", "😂", "😄"},
	"warmth":     {"love", "miss", "sweet", "kind", "heart", "care", "dear", "❤️", "🥰"},
	"excitement": {"excited", "amazing", "awesome", "wow", "omg", "can't wait", "thrilled", "🎉"},
	"nostalgia":  {"remember", "used to", "back then", "childhood", "memories", "miss those"},
	"aspiration": {"want to", "wish", "dream", "someday", "plan to", "hope to"},
}

var preferenceKeywords = []string{
	"want to", "love to", "i like", "i enjoy", "prefer", "favorite", "favourite",
	"i love", "really into", "passionate about", "keen on",
}

var wishKeywords = []string{
	"someday", "wish", "dream of", "plan to", "going to", "would love",
	"hope to", "want to try", "always wanted",
}

func ChunkMessages(sessionID string, messages []domain.Message, windowSize int, overlapSize int) ([]domain.Chunk, error) {
	step := windowSize - overlapSize
	if step <= 0 {
		step = 1
	}

	var chunks []domain.Chunk
	chunkIndex := 0

	for start := 0; start < len(messages); start += step {
		end := start + windowSize
		if end > len(messages) {
			end = len(messages)
		}

		window := messages[start:end]

		if isMediaOnlyWindow(window) {
			continue
		}

		chunk := buildChunk(sessionID, chunkIndex, window, start, end-1)
		chunks = append(chunks, chunk)
		chunkIndex++

		if end == len(messages) {
			break
		}
	}

	return chunks, nil
}

func isMediaOnlyWindow(messages []domain.Message) bool {
	for _, m := range messages {
		if !m.IsMedia {
			return false
		}
	}
	return true
}

func buildChunk(sessionID string, index int, messages []domain.Message, start, end int) domain.Chunk {
	var textParts []string
	for _, m := range messages {
		if !m.IsMedia {
			textParts = append(textParts, fmt.Sprintf("%s: %s", m.Sender, m.Text))
		}
	}
	text := strings.Join(textParts, "\n")

	return domain.Chunk{
		ID:             fmt.Sprintf("%s_chunk_%d", sessionID, index),
		SessionID:      sessionID,
		AnonymizedText: text,
		StartIndex:     start,
		EndIndex:       end,
		Metadata:       enrichMetadata(text),
	}
}

func enrichMetadata(text string) domain.ChunkMetadata {
	lower := strings.ToLower(text)
	return domain.ChunkMetadata{
		Topics:           detectTopics(lower),
		EmotionalMarkers: detectEmotions(lower),
		HasPreference:    containsAny(lower, preferenceKeywords),
		HasWish:          containsAny(lower, wishKeywords),
	}
}

func detectTopics(lowerText string) []string {
	var found []string
	for topic, keywords := range topicKeywords {
		if containsAny(lowerText, keywords) {
			found = append(found, topic)
		}
	}
	return found
}

func detectEmotions(lowerText string) []string {
	var found []string
	for emotion, keywords := range emotionKeywords {
		if containsAny(lowerText, keywords) {
			found = append(found, emotion)
		}
	}
	return found
}

func containsAny(text string, keywords []string) bool {
	for _, kw := range keywords {
		if strings.Contains(text, kw) {
			return true
		}
	}
	return false
}
