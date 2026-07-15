package service

import (
	"math"
	"testing"

	"github.com/harbour-wave/harbour-wave-backend/internal/model"
)

func TestComputeTotal(t *testing.T) {
	svc := NewPricingService()

	tests := []struct {
		name  string
		route string
		q     model.Quantities
		want  float64
	}{
		{"single big on Desna", RouteDesna, model.Quantities{Big: 1}, 450},
		{"mixed boats on Desna", RouteDesna, model.Quantities{Big: 2, Medium: 1, Small: 1}, 1800},
		{"child is half price on Desna", RouteDesna, model.Quantities{Big: 1, Child: 1}, 675},
		{"children only", RouteDesna, model.Quantities{Child: 2}, 450},
		{"zero quantities", RouteDesna, model.Quantities{}, 0},
		{"unknown route is defensive zero", "Atlantis", model.Quantities{Big: 3}, 0},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := svc.ComputeTotal(tt.route, tt.q); got != tt.want {
				t.Errorf("ComputeTotal(%q, %+v) = %v, want %v", tt.route, tt.q, got, tt.want)
			}
		})
	}
}

func TestEffectiveAmount(t *testing.T) {
	svc := NewPricingService()

	if got := svc.EffectiveAmount(450, nil); got != 450 {
		t.Errorf("nil override: got %v, want 450", got)
	}
	if got := svc.EffectiveAmount(450, fptr(300)); got != 300 {
		t.Errorf("override wins: got %v, want 300", got)
	}
	if got := svc.EffectiveAmount(450, fptr(0)); got != 0 {
		t.Errorf("zero override still wins: got %v, want 0", got)
	}
}

func TestRoutePrice(t *testing.T) {
	svc := NewPricingService()

	p, ok := svc.RoutePrice(RouteDesna)
	if !ok {
		t.Fatal("Desna should be a known route")
	}
	if p.Big != 450 || p.Medium != 450 || p.Small != 450 || p.Child != 225 {
		t.Errorf("unexpected Desna prices: %+v", p)
	}

	if _, ok := svc.RoutePrice("Atlantis"); ok {
		t.Error("unknown route should report ok=false")
	}
}

func TestIsValidRouteAndAllRoutes(t *testing.T) {
	if !IsValidRoute(RouteDesna) || !IsValidRoute(RouteKlochkov) {
		t.Error("known routes must validate")
	}
	if IsValidRoute("") || IsValidRoute("desna") {
		t.Error("unknown / wrong-case routes must not validate")
	}

	routes := AllRoutes()
	if len(routes) != 2 || routes[0] != RouteDesna || routes[1] != RouteKlochkov {
		t.Errorf("AllRoutes() = %v, want deterministic [Desna Klochkov]", routes)
	}
}

func TestApplyDiscount(t *testing.T) {
	svc := NewPricingService()

	tests := []struct {
		name           string
		total          float64
		pct            int
		wantDiscounted float64
		wantAmount     float64
	}{
		{"zero percent is a no-op", 450, 0, 450, 0},
		{"negative percent is a no-op", 450, -10, 450, 0},
		{"zero total is a no-op", 0, 50, 0, 0},
		{"negative total is a no-op", -10, 50, -10, 0},
		{"regular 10%", 450, 10, 405, 45},
		{"rounding to 2dp", 99.99, 33, 66.99, 33},
		{"100% zeroes the total", 450, 100, 0, 450},
		{"over 100% clamps like 100%", 450, 150, 0, 450},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			discounted, amount := svc.ApplyDiscount(tt.total, tt.pct)
			if discounted != tt.wantDiscounted || amount != tt.wantAmount {
				t.Errorf("ApplyDiscount(%v, %d) = (%v, %v), want (%v, %v)",
					tt.total, tt.pct, discounted, amount, tt.wantDiscounted, tt.wantAmount)
			}
		})
	}
}

func TestApplyDiscountInvariant(t *testing.T) {
	// discounted + amount must reconstruct the (rounded) total for regular percentages.
	svc := NewPricingService()
	for pct := 1; pct < 100; pct++ {
		discounted, amount := svc.ApplyDiscount(1125, pct)
		if math.Abs((discounted+amount)-1125) > 0.011 {
			t.Errorf("pct=%d: discounted %v + amount %v drifts from total 1125", pct, discounted, amount)
		}
		if discounted < 0 || amount < 0 {
			t.Errorf("pct=%d: negative outputs (%v, %v)", pct, discounted, amount)
		}
	}
}

func TestRound2(t *testing.T) {
	tests := []struct {
		in   float64
		want float64
	}{
		{1.004, 1.00},
		{1.005, 1.0}, // 1.005 is stored as 1.00499… so it rounds down
		{1.006, 1.01},
		{-1.006, -1.01},
		{0, 0},
	}
	for _, tt := range tests {
		if got := round2(tt.in); got != tt.want {
			t.Errorf("round2(%v) = %v, want %v", tt.in, got, tt.want)
		}
	}
}
