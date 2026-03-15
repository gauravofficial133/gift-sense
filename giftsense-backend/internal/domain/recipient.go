package domain

type BudgetTier string

const (
	BudgetTierBudget   BudgetTier = "BUDGET"
	BudgetTierMidRange BudgetTier = "MID_RANGE"
	BudgetTierPremium  BudgetTier = "PREMIUM"
	BudgetTierLuxury   BudgetTier = "LUXURY"
)

type BudgetRange struct {
	Tier   BudgetTier
	MinINR int
	MaxINR int
}

var BudgetRanges = map[BudgetTier]BudgetRange{
	BudgetTierBudget:   {Tier: BudgetTierBudget, MinINR: 500, MaxINR: 1000},
	BudgetTierMidRange: {Tier: BudgetTierMidRange, MinINR: 1000, MaxINR: 5000},
	BudgetTierPremium:  {Tier: BudgetTierPremium, MinINR: 5000, MaxINR: 15000},
	BudgetTierLuxury:   {Tier: BudgetTierLuxury, MinINR: 15000, MaxINR: 0},
}

type RecipientDetails struct {
	Name     string
	Relation string
	Gender   string
	Occasion string
	Budget   BudgetRange
}
