package domain_test

import (
	"testing"

	"github.com/giftsense/backend/internal/domain"
	"github.com/stretchr/testify/assert"
)

func TestBudgetRanges_ShouldHaveFourTiers(t *testing.T) {
	assert.Len(t, domain.BudgetRanges, 4)
}

func TestBudgetRanges_ShouldHaveCorrectINRValues(t *testing.T) {
	tests := []struct {
		tier   domain.BudgetTier
		minINR int
		maxINR int
	}{
		{domain.BudgetTierBudget, 500, 1000},
		{domain.BudgetTierMidRange, 1000, 5000},
		{domain.BudgetTierPremium, 5000, 15000},
		{domain.BudgetTierLuxury, 15000, 0},
	}

	for _, tt := range tests {
		t.Run(string(tt.tier), func(t *testing.T) {
			r, ok := domain.BudgetRanges[tt.tier]
			assert.True(t, ok, "tier should exist in BudgetRanges")
			assert.Equal(t, tt.minINR, r.MinINR)
			assert.Equal(t, tt.maxINR, r.MaxINR)
		})
	}
}

func TestBudgetTierConstants_ShouldMatchExpectedStringValues(t *testing.T) {
	assert.Equal(t, domain.BudgetTier("BUDGET"), domain.BudgetTierBudget)
	assert.Equal(t, domain.BudgetTier("MID_RANGE"), domain.BudgetTierMidRange)
	assert.Equal(t, domain.BudgetTier("PREMIUM"), domain.BudgetTierPremium)
	assert.Equal(t, domain.BudgetTier("LUXURY"), domain.BudgetTierLuxury)
}

func TestSentinelErrors_ShouldBeNonNil(t *testing.T) {
	assert.NotNil(t, domain.ErrFileTooLarge)
	assert.NotNil(t, domain.ErrConversationTooShort)
	assert.NotNil(t, domain.ErrInvalidBudgetTier)
	assert.NotNil(t, domain.ErrInvalidSessionID)
	assert.NotNil(t, domain.ErrInvalidFileType)
	assert.NotNil(t, domain.ErrLLMResponseInvalid)
	assert.NotNil(t, domain.ErrRetrievalFailed)
	assert.NotNil(t, domain.ErrAllSuggestionsFiltered)
}
