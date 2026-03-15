package linkgen

import (
	"fmt"
	"net/url"

	"github.com/giftsense/backend/internal/domain"
)

func GenerateLinks(giftName string, budget domain.BudgetRange) domain.ShoppingLinks {
	encoded := url.QueryEscape(giftName)
	return domain.ShoppingLinks{
		Amazon:         buildAmazonURL(encoded, budget),
		Flipkart:       buildFlipkartURL(encoded, budget),
		GoogleShopping: buildGoogleShoppingURL(encoded, budget),
	}
}

func buildAmazonURL(encodedName string, budget domain.BudgetRange) string {
	minPaise := budget.MinINR * 100
	if budget.MaxINR == 0 {
		return fmt.Sprintf("https://www.amazon.in/s?k=%s&rh=p_36%%3A%d", encodedName, minPaise)
	}
	maxPaise := budget.MaxINR * 100
	return fmt.Sprintf("https://www.amazon.in/s?k=%s&rh=p_36%%3A%d-%d", encodedName, minPaise, maxPaise)
}

func buildFlipkartURL(encodedName string, budget domain.BudgetRange) string {
	if budget.MaxINR == 0 {
		return fmt.Sprintf(
			"https://www.flipkart.com/search?q=%s&p[]=facets.price_range.from%%3D%d",
			encodedName, budget.MinINR,
		)
	}
	return fmt.Sprintf(
		"https://www.flipkart.com/search?q=%s&p[]=facets.price_range.from%%3D%d&p[]=facets.price_range.to%%3D%d",
		encodedName, budget.MinINR, budget.MaxINR,
	)
}

func buildGoogleShoppingURL(encodedName string, budget domain.BudgetRange) string {
	if budget.MaxINR == 0 {
		return fmt.Sprintf("https://www.google.com/search?q=%s+premium+luxury+gift&tbm=shop", encodedName)
	}
	return fmt.Sprintf(
		"https://www.google.com/search?q=%s+under+%%E2%%82%%B9%d&tbm=shop",
		encodedName, budget.MaxINR,
	)
}
