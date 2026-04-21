package domain

import "errors"

var (
	ErrFileTooLarge           = errors.New("file exceeds maximum allowed size")
	ErrConversationTooShort   = errors.New("conversation has too few messages to analyze")
	ErrInvalidBudgetTier      = errors.New("invalid budget tier")
	ErrInvalidSessionID       = errors.New("invalid session ID format")
	ErrInvalidFileType        = errors.New("only .txt files are accepted")
	ErrLLMResponseInvalid     = errors.New("LLM returned invalid or non-conformant JSON")
	ErrRetrievalFailed        = errors.New("retrieval returned no relevant context")
	ErrAllSuggestionsFiltered = errors.New("all suggestions violated budget constraints")
	ErrCardThemeNotFound      = errors.New("no card theme found for the given emotion")
	ErrCardRenderFailed       = errors.New("card render failed")
)
