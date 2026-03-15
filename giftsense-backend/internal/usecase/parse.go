package usecase

import (
	"fmt"
	"regexp"
	"strings"
	"time"

	"github.com/giftsense/backend/internal/domain"
)

var whatsAppLinePattern = regexp.MustCompile(`^\[(\d{2}/\d{2}/\d{4}, \d{2}:\d{2}:\d{2})\] ([^:]+): (.*)$`)
var plainTextSenderPattern = regexp.MustCompile(`^([A-Za-z][A-Za-z\s]{0,20}):\s(.+)$`)

const minParsedMessages = 5

func ParseConversation(text string, maxMessages int) ([]domain.Message, error) {
	lines := strings.Split(strings.TrimSpace(text), "\n")

	messages := parseWhatsApp(lines)
	if len(messages) < minParsedMessages {
		messages = parsePlainText(lines)
	}

	messages = filterSystemMessages(messages)

	if len(messages) < minParsedMessages {
		return nil, fmt.Errorf("parse conversation: %w", domain.ErrConversationTooShort)
	}

	messages = sampleMessages(messages, maxMessages)

	return indexMessages(messages), nil
}

func parseWhatsApp(lines []string) []domain.Message {
	var messages []domain.Message
	var current *domain.Message

	for _, line := range lines {
		matches := whatsAppLinePattern.FindStringSubmatch(line)
		if matches != nil {
			if current != nil {
				messages = append(messages, *current)
			}
			ts, _ := time.Parse("02/01/2006, 15:04:05", matches[1])
			isMedia := strings.Contains(matches[3], "<Media omitted>") || strings.Contains(matches[3], "image omitted")
			current = &domain.Message{
				Sender:    strings.TrimSpace(matches[2]),
				Text:      matches[3],
				Timestamp: ts,
				IsMedia:   isMedia,
			}
		} else if current != nil && !isSystemMessage(line) {
			current.Text += "\n" + line
		}
	}
	if current != nil {
		messages = append(messages, *current)
	}
	return messages
}

func parsePlainText(lines []string) []domain.Message {
	var messages []domain.Message
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		matches := plainTextSenderPattern.FindStringSubmatch(line)
		if matches != nil {
			messages = append(messages, domain.Message{
				Sender: strings.TrimSpace(matches[1]),
				Text:   strings.TrimSpace(matches[2]),
			})
		} else {
			messages = append(messages, domain.Message{
				Sender: "Unknown",
				Text:   line,
			})
		}
	}
	return messages
}

var systemMessagePatterns = []string{
	"end-to-end encrypted",
	"Messages and calls are",
	"changed their phone number",
	"added you",
	"created group",
	"changed the subject",
}

func filterSystemMessages(messages []domain.Message) []domain.Message {
	filtered := make([]domain.Message, 0, len(messages))
	for _, m := range messages {
		if !isSystemMessage(m.Text) {
			filtered = append(filtered, m)
		}
	}
	return filtered
}

func isSystemMessage(text string) bool {
	lower := strings.ToLower(text)
	for _, pattern := range systemMessagePatterns {
		if strings.Contains(lower, strings.ToLower(pattern)) {
			return true
		}
	}
	return false
}

func sampleMessages(messages []domain.Message, maxMessages int) []domain.Message {
	if len(messages) <= maxMessages {
		return messages
	}

	recentCount := maxMessages / 4
	olderCount := maxMessages - recentCount

	recentStart := len(messages) - recentCount
	recent := messages[recentStart:]

	older := messages[:recentStart]
	sampled := sampleEvenly(older, olderCount)

	return append(sampled, recent...)
}

func sampleEvenly(messages []domain.Message, count int) []domain.Message {
	if len(messages) <= count {
		return messages
	}
	result := make([]domain.Message, 0, count)
	step := float64(len(messages)) / float64(count)
	for i := 0; i < count; i++ {
		idx := int(float64(i) * step)
		result = append(result, messages[idx])
	}
	return result
}

func indexMessages(messages []domain.Message) []domain.Message {
	for i := range messages {
		messages[i].Index = i
	}
	return messages
}
