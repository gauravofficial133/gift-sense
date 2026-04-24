package usecase_test

import (
	"context"
	"encoding/json"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/giftsense/backend/internal/adapter/vectorstore"
	"github.com/giftsense/backend/internal/domain"
	"github.com/giftsense/backend/internal/port"
	"github.com/giftsense/backend/internal/usecase"
)

// fakeEmbedder returns distinct non-zero vectors so cosine similarity is meaningful.
type fakeEmbedder struct{}

func (f *fakeEmbedder) Embed(_ context.Context, texts []string) ([][]float32, error) {
	vecs := make([][]float32, len(texts))
	for i := range vecs {
		vecs[i] = make([]float32, 8)
		vecs[i][i%8] = float32(i+1) * 0.1
	}
	return vecs, nil
}

type errEmbedder struct{}

func (e *errEmbedder) Embed(_ context.Context, _ []string) ([][]float32, error) {
	return nil, errors.New("embed error")
}

type fakeLLM struct {
	response string
	err      error
}

func (f *fakeLLM) Complete(_ context.Context, _ string, _ port.CompletionOptions) (string, error) {
	return f.response, f.err
}

func validLLMJSON() string {
	resp := map[string]interface{}{
		"personality_insights": []map[string]string{
			{"insight": "Loves hiking", "evidence_summary": "Mentioned hiking three times"},
		},
		"gift_suggestions": []map[string]string{
			{"name": "Hiking poles", "reason": "Great for their hobby", "estimated_price_inr": "₹1500", "category": "outdoor"},
		},
	}
	b, _ := json.Marshal(resp)
	return string(b)
}

func defaultConfig() usecase.AnalyzerConfig {
	return usecase.AnalyzerConfig{
		MaxProcessedMessages: 400,
		ChunkWindowSize:      4,
		ChunkOverlapSize:     1,
		TopK:                 3,
		NumRetrievalQueries:  2,
	}
}

func defaultRecipient() domain.RecipientDetails {
	return domain.RecipientDetails{
		Name:     "Alex",
		Relation: "friend",
		Gender:   "other",
		Occasion: "birthday",
		Budget:   domain.BudgetRanges[domain.BudgetTierMidRange],
	}
}

func noopLinkGen(name string, _ domain.BudgetRange) domain.ShoppingLinks {
	return domain.ShoppingLinks{Amazon: "https://amazon.in/s?k=" + name}
}

func enoughConversation() string {
	return "Alice: I love hiking and being outdoors.\n" +
		"Bob: Really? I went to the mountains last weekend.\n" +
		"Alice: That sounds amazing, I want to try rock climbing too.\n" +
		"Bob: Yes, it's on my list! Also want to get hiking poles someday.\n" +
		"Alice: Same! Let's plan a trip together.\n" +
		"Bob: Definitely, maybe next month.\n" +
		"Alice: I also love photography when I'm out in nature.\n" +
		"Bob: Nice, I've been meaning to get a good camera.\n"
}

func TestAnalyze_ShouldReturnResult_WhenValidConversationProvided(t *testing.T) {
	store := vectorstore.NewMemoryStore()
	analyzer := usecase.NewAnalyzer(&fakeEmbedder{}, &fakeLLM{response: validLLMJSON()}, store, noopLinkGen, defaultConfig(), nil, nil, nil, nil, nil)

	result, err := analyzer.Analyze(context.Background(), "sess-1", enoughConversation(), defaultRecipient())

	require.NoError(t, err)
	assert.Len(t, result.PersonalityInsights, 1)
	assert.Equal(t, "Loves hiking", result.PersonalityInsights[0].Insight)
	assert.Len(t, result.GiftSuggestions, 1)
	assert.Equal(t, "Hiking poles", result.GiftSuggestions[0].Name)
	assert.NotEmpty(t, result.GiftSuggestions[0].Links.Amazon)
}

func TestAnalyze_ShouldReturnError_WhenConversationTooShort(t *testing.T) {
	store := vectorstore.NewMemoryStore()
	analyzer := usecase.NewAnalyzer(&fakeEmbedder{}, &fakeLLM{response: validLLMJSON()}, store, noopLinkGen, defaultConfig(), nil, nil, nil, nil, nil)

	_, err := analyzer.Analyze(context.Background(), "sess-2", "Alice: Hi\nBob: Hey", defaultRecipient())

	require.Error(t, err)
	assert.ErrorIs(t, err, domain.ErrConversationTooShort)
}

func TestAnalyze_ShouldReturnError_WhenLLMReturnsInvalidJSON(t *testing.T) {
	store := vectorstore.NewMemoryStore()
	analyzer := usecase.NewAnalyzer(&fakeEmbedder{}, &fakeLLM{response: "not json"}, store, noopLinkGen, defaultConfig(), nil, nil, nil, nil, nil)

	_, err := analyzer.Analyze(context.Background(), "sess-3", enoughConversation(), defaultRecipient())

	require.Error(t, err)
	assert.ErrorIs(t, err, domain.ErrLLMResponseInvalid)
}

func TestAnalyze_ShouldReturnError_WhenLLMReturnsEmptySuggestions(t *testing.T) {
	store := vectorstore.NewMemoryStore()
	emptyJSON := `{"personality_insights":[],"gift_suggestions":[]}`
	analyzer := usecase.NewAnalyzer(&fakeEmbedder{}, &fakeLLM{response: emptyJSON}, store, noopLinkGen, defaultConfig(), nil, nil, nil, nil, nil)

	_, err := analyzer.Analyze(context.Background(), "sess-4", enoughConversation(), defaultRecipient())

	require.Error(t, err)
	assert.ErrorIs(t, err, domain.ErrLLMResponseInvalid)
}

func TestAnalyze_ShouldDeleteSession_AfterAnalysis(t *testing.T) {
	store := vectorstore.NewMemoryStore()
	analyzer := usecase.NewAnalyzer(&fakeEmbedder{}, &fakeLLM{response: validLLMJSON()}, store, noopLinkGen, defaultConfig(), nil, nil, nil, nil, nil)

	_, err := analyzer.Analyze(context.Background(), "sess-5", enoughConversation(), defaultRecipient())
	require.NoError(t, err)

	// After analysis the namespace should be empty.
	chunks, qErr := store.Query(context.Background(), "sess-5", make([]float32, 8), 10, port.MetadataFilter{})
	require.NoError(t, qErr)
	assert.Empty(t, chunks)
}

func TestBuildAnalysisPrompt_ShouldIncludeRecipientDetails(t *testing.T) {
	chunks := []domain.Chunk{
		{AnonymizedText: "[Person_A] loves hiking outdoors."},
	}
	recipient := defaultRecipient()

	prompt := usecase.BuildAnalysisPrompt(chunks, recipient)

	assert.Contains(t, prompt, "Alex")
	assert.Contains(t, prompt, "birthday")
	assert.Contains(t, prompt, "[Person_A] loves hiking outdoors.")
}
