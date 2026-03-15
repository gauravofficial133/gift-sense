package usecase

import (
	"fmt"
	"strings"

	"github.com/giftsense/backend/internal/domain"
)

func AnonymizeMessages(messages []domain.Message) ([]domain.Message, map[string]string, error) {
	tokenMap := buildTokenMap(messages)

	anonymized := make([]domain.Message, len(messages))
	for i, msg := range messages {
		anonymized[i] = anonymizeMessage(msg, tokenMap)
	}

	return anonymized, tokenMap, nil
}

func buildTokenMap(messages []domain.Message) map[string]string {
	tokenMap := make(map[string]string)
	personIndex := 0

	for _, msg := range messages {
		if msg.Sender == "" {
			continue
		}
		if _, exists := tokenMap[msg.Sender]; !exists {
			tokenMap[msg.Sender] = fmt.Sprintf("[Person_%s]", letterLabel(personIndex))
			personIndex++
		}
	}

	return tokenMap
}

func anonymizeMessage(msg domain.Message, tokenMap map[string]string) domain.Message {
	result := msg

	if token, ok := tokenMap[msg.Sender]; ok {
		result.Sender = token
	}

	result.Text = replaceNamesInText(msg.Text, tokenMap)

	return result
}

func replaceNamesInText(text string, tokenMap map[string]string) string {
	for original, token := range tokenMap {
		text = strings.ReplaceAll(text, original, token)
	}
	return text
}

func letterLabel(index int) string {
	if index < 26 {
		return string(rune('A' + index))
	}
	return fmt.Sprintf("%d", index+1)
}
