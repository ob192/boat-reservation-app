package service

import "github.com/harbour-wave/harbour-wave-backend/internal/model"

const (
	RouteDesna    = "Desna"
	RouteKlochkov = "Klochkov"
)

// RoutePricing holds per-boat-class unit prices (EUR) for one route.
type RoutePricing struct {
	Big    float64
	Medium float64
	Small  float64
	Child  float64
}

// routePrices is the single source of truth for route costs.
var routePrices = map[string]RoutePricing{
	RouteDesna:    {Big: 450.00, Medium: 450.00, Small: 450.00, Child: 225.00},
	RouteKlochkov: {Big: 99999.00, Medium: 99999.00, Small: 99999.00, Child: 99999.00},
}

// AllRoutes returns the route catalog in a deterministic order (used by seeding).
func AllRoutes() []string {
	return []string{RouteDesna, RouteKlochkov}
}

// IsValidRoute reports whether name is a known route.
func IsValidRoute(name string) bool {
	_, ok := routePrices[name]
	return ok
}

// RoutePriceFor returns the price table for a route.
func RoutePriceFor(name string) (RoutePricing, bool) {
	p, ok := routePrices[name]
	return p, ok
}

// PricingService computes booking totals and resolves admin overrides.
type PricingService interface {
	ComputeTotal(routeName string, q model.Quantities) float64
	EffectiveAmount(total float64, override *float64) float64
	RoutePrice(routeName string) (RoutePricing, bool)
}

type pricingService struct{}

func NewPricingService() PricingService { return &pricingService{} }

func (pricingService) ComputeTotal(routeName string, q model.Quantities) float64 {
	p, ok := routePrices[routeName]
	if !ok {
		// Unknown route: callers validate the route before reaching here,
		// so this is a defensive zero rather than a silent mispricing path.
		return 0
	}
	return float64(q.Big)*p.Big +
		float64(q.Medium)*p.Medium +
		float64(q.Small)*p.Small +
		float64(q.Child)*p.Child
}

func (pricingService) EffectiveAmount(total float64, override *float64) float64 {
	if override != nil {
		return *override
	}
	return total
}

func (pricingService) RoutePrice(routeName string) (RoutePricing, bool) {
	return RoutePriceFor(routeName)
}
