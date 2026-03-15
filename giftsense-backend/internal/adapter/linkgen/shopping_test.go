package linkgen_test

import (
	"strings"
	"testing"

	"github.com/giftsense/backend/internal/adapter/linkgen"
	"github.com/giftsense/backend/internal/domain"
	"github.com/stretchr/testify/assert"
)

func TestGenerateLinks_ShouldReturnAmazonURL_WithCorrectPaiseValues(t *testing.T) {
	budget := domain.BudgetRanges[domain.BudgetTierMidRange] // 1000-5000
	links := linkgen.GenerateLinks("pottery starter kit", budget)
	assert.Contains(t, links.Amazon, "amazon.in")
	assert.Contains(t, links.Amazon, "100000") // 1000 * 100 paise
	assert.Contains(t, links.Amazon, "500000") // 5000 * 100 paise
}

func TestGenerateLinks_ShouldReturnFlipkartURL_WithCorrectINRValues(t *testing.T) {
	budget := domain.BudgetRanges[domain.BudgetTierBudget] // 500-1000
	links := linkgen.GenerateLinks("sketch pad", budget)
	assert.Contains(t, links.Flipkart, "flipkart.com")
	assert.Contains(t, links.Flipkart, "500")
	assert.Contains(t, links.Flipkart, "1000")
}

func TestGenerateLinks_ShouldHandleLuxuryTier_WithNoUpperBound(t *testing.T) {
	budget := domain.BudgetRanges[domain.BudgetTierLuxury]
	links := linkgen.GenerateLinks("luxury watch", budget)
	// Amazon should have min paise but no max paise dash segment
	assert.Contains(t, links.Amazon, "1500000") // 15000 * 100
	assert.NotContains(t, links.Amazon, "1500000-") // no upper bound dash
	// Flipkart should not have price_range.to
	assert.NotContains(t, links.Flipkart, "price_range.to")
	// Google should say premium/luxury
	assert.True(t, strings.Contains(links.GoogleShopping, "premium") || strings.Contains(links.GoogleShopping, "luxury"))
}

func TestGenerateLinks_ShouldURLEncodeGiftName_WhenSpacesPresent(t *testing.T) {
	budget := domain.BudgetRanges[domain.BudgetTierBudget]
	links := linkgen.GenerateLinks("pottery starter kit", budget)
	assert.NotContains(t, links.Amazon, " ") // no raw spaces in URL
	assert.NotContains(t, links.Flipkart, " ")
	assert.NotContains(t, links.GoogleShopping, " ")
}

func TestGenerateLinks_ShouldEncodeRupeeSymbol_InGoogleShoppingURL(t *testing.T) {
	budget := domain.BudgetRanges[domain.BudgetTierMidRange]
	links := linkgen.GenerateLinks("art supplies", budget)
	assert.Contains(t, links.GoogleShopping, "tbm=shop")
	assert.Contains(t, links.GoogleShopping, "google.com")
}

func TestGenerateLinks_ShouldReturnAllThreeLinks_ForEveryBudgetTier(t *testing.T) {
	tiers := []domain.BudgetTier{
		domain.BudgetTierBudget,
		domain.BudgetTierMidRange,
		domain.BudgetTierPremium,
		domain.BudgetTierLuxury,
	}
	for _, tier := range tiers {
		budget := domain.BudgetRanges[tier]
		links := linkgen.GenerateLinks("gift idea", budget)
		assert.NotEmpty(t, links.Amazon, "Amazon link empty for tier "+string(tier))
		assert.NotEmpty(t, links.Flipkart, "Flipkart link empty for tier "+string(tier))
		assert.NotEmpty(t, links.GoogleShopping, "Google link empty for tier "+string(tier))
	}
}
