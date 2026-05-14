package service

import "github.com/harbour-wave/harbour-wave-backend/internal/model"

// Unit prices in EUR. Single source of truth for the whole codebase.
const (
	PriceBig    = 35.00
	PriceMedium = 20.00
	PriceChild  = 17.50
)

// PricingService computes booking totals and resolves admin overrides.
//
// Kept as an interface so a future "seasonal pricing" or "promo code" engine
// can be swapped in without touching the booking service.
type PricingService interface {
	ComputeTotal(q model.Quantities) float64
	EffectiveAmount(total float64, override *float64) float64
}

type pricingService struct{}

func NewPricingService() PricingService { return &pricingService{} }

func (pricingService) ComputeTotal(q model.Quantities) float64 {
	return float64(q.Big)*PriceBig +
		float64(q.Medium)*PriceMedium +
		float64(q.Child)*PriceChild
}

func (pricingService) EffectiveAmount(total float64, override *float64) float64 {
	if override != nil {
		return *override
	}
	return total
}
