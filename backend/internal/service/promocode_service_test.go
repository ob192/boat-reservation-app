package service

import (
	"context"
	"errors"
	"testing"

	"github.com/google/uuid"

	"github.com/harbour-wave/harbour-wave-backend/internal/model"
	"github.com/harbour-wave/harbour-wave-backend/internal/repository"
)

func TestNormalizeCode(t *testing.T) {
	tests := []struct{ in, want string }{
		{"summer", "SUMMER"},
		{"  Summer10  ", "SUMMER10"},
		{"ALREADY", "ALREADY"},
		{"", ""},
		{"   ", ""},
	}
	for _, tt := range tests {
		if got := NormalizeCode(tt.in); got != tt.want {
			t.Errorf("NormalizeCode(%q) = %q, want %q", tt.in, got, tt.want)
		}
	}
}

func TestPromocodeCreate(t *testing.T) {
	adminID := uuid.New()
	clock := fakeClock{t: testNow}

	notFoundRepo := func() *mockPromoRepo {
		return &mockPromoRepo{
			findByCode: func(context.Context, string) (*model.Promocode, error) {
				return nil, repository.ErrNotFound
			},
		}
	}

	t.Run("invalid inputs", func(t *testing.T) {
		svc := NewPromocodeService(&mockPromoRepo{}, clock) // repo must not be touched
		invalid := []struct {
			name    string
			code    string
			pct     int
			maxUses int
		}{
			{"empty code", "", 10, 5},
			{"whitespace code", "   ", 10, 5},
			{"negative percent", "OK", -1, 5},
			{"percent over 100", "OK", 101, 5},
			{"zero max uses", "OK", 10, 0},
			{"negative max uses", "OK", 10, -3},
		}
		for _, tt := range invalid {
			t.Run(tt.name, func(t *testing.T) {
				_, err := svc.Create(context.Background(), tt.code, tt.pct, tt.maxUses, adminID)
				if !errors.Is(err, ErrInvalidInput) {
					t.Errorf("want ErrInvalidInput, got %v", err)
				}
			})
		}
	})

	t.Run("duplicate code", func(t *testing.T) {
		repo := &mockPromoRepo{
			findByCode: func(_ context.Context, code string) (*model.Promocode, error) {
				return &model.Promocode{Code: code}, nil
			},
		}
		_, err := NewPromocodeService(repo, clock).Create(context.Background(), "DUP", 10, 5, adminID)
		if !errors.Is(err, ErrPromoAlreadyExists) {
			t.Errorf("want ErrPromoAlreadyExists, got %v", err)
		}
	})

	t.Run("lookup failure propagates", func(t *testing.T) {
		boom := errors.New("db down")
		repo := &mockPromoRepo{
			findByCode: func(context.Context, string) (*model.Promocode, error) { return nil, boom },
		}
		_, err := NewPromocodeService(repo, clock).Create(context.Background(), "X", 10, 5, adminID)
		if !errors.Is(err, boom) {
			t.Errorf("want repo error, got %v", err)
		}
	})

	t.Run("insert failure propagates", func(t *testing.T) {
		boom := errors.New("insert failed")
		repo := notFoundRepo()
		repo.create = func(context.Context, *model.Promocode) error { return boom }
		_, err := NewPromocodeService(repo, clock).Create(context.Background(), "X", 10, 5, adminID)
		if !errors.Is(err, boom) {
			t.Errorf("want insert error, got %v", err)
		}
	})

	t.Run("success normalises code and freezes defaults", func(t *testing.T) {
		var saved *model.Promocode
		repo := notFoundRepo()
		repo.create = func(_ context.Context, p *model.Promocode) error {
			saved = p
			return nil
		}
		resp, err := NewPromocodeService(repo, clock).Create(context.Background(), "  summer10 ", 15, 100, adminID)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if saved == nil {
			t.Fatal("promocode was not persisted")
		}
		if saved.Code != "SUMMER10" {
			t.Errorf("saved code = %q, want normalised SUMMER10", saved.Code)
		}
		if saved.DiscountPercent != 15 || saved.MaxUses != 100 {
			t.Errorf("saved pct/maxUses = %d/%d, want 15/100", saved.DiscountPercent, saved.MaxUses)
		}
		if saved.TimesUsed != 0 || !saved.Active {
			t.Errorf("new promo must start unused and active, got used=%d active=%v", saved.TimesUsed, saved.Active)
		}
		if saved.CreatedBy != adminID {
			t.Errorf("createdBy = %v, want %v", saved.CreatedBy, adminID)
		}
		if !saved.CreatedAt.Equal(testNow) {
			t.Errorf("createdAt = %v, want clock time %v", saved.CreatedAt, testNow)
		}

		if resp.Code != "SUMMER10" || resp.DiscountPercent != 15 || resp.MaxUses != 100 ||
			resp.TimesUsed != 0 || !resp.Active || resp.CreatedBy != adminID.String() {
			t.Errorf("response not mapped from saved row: %+v", resp)
		}
	})

	t.Run("boundary percents are allowed", func(t *testing.T) {
		for _, pct := range []int{0, 100} {
			repo := notFoundRepo()
			repo.create = func(context.Context, *model.Promocode) error { return nil }
			if _, err := NewPromocodeService(repo, clock).Create(context.Background(), "B", pct, 1, adminID); err != nil {
				t.Errorf("pct=%d should be valid, got %v", pct, err)
			}
		}
	})
}

func TestPromocodeList(t *testing.T) {
	clock := fakeClock{t: testNow}
	adminID := uuid.New()

	t.Run("maps rows and passes filter through", func(t *testing.T) {
		var gotFilter *uuid.UUID
		repo := &mockPromoRepo{
			findAll: func(_ context.Context, createdBy *uuid.UUID) ([]model.Promocode, error) {
				gotFilter = createdBy
				return []model.Promocode{
					{Code: "A", DiscountPercent: 5, MaxUses: 10, TimesUsed: 3, Active: true, CreatedBy: adminID, CreatedAt: testNow},
					{Code: "B", DiscountPercent: 50, MaxUses: 1, TimesUsed: 1, Active: false, CreatedBy: adminID, CreatedAt: testNow},
				}, nil
			},
		}
		resp, err := NewPromocodeService(repo, clock).List(context.Background(), &adminID)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if gotFilter == nil || *gotFilter != adminID {
			t.Errorf("createdBy filter not passed through, got %v", gotFilter)
		}
		if len(resp.Promocodes) != 2 {
			t.Fatalf("got %d promocodes, want 2", len(resp.Promocodes))
		}
		if resp.Promocodes[0].Code != "A" || resp.Promocodes[0].TimesUsed != 3 {
			t.Errorf("first row mapped wrong: %+v", resp.Promocodes[0])
		}
		if resp.Promocodes[1].Active {
			t.Error("second row should stay inactive after mapping")
		}
	})

	t.Run("empty result is an empty slice, not nil", func(t *testing.T) {
		repo := &mockPromoRepo{
			findAll: func(context.Context, *uuid.UUID) ([]model.Promocode, error) { return nil, nil },
		}
		resp, err := NewPromocodeService(repo, clock).List(context.Background(), nil)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if resp.Promocodes == nil || len(resp.Promocodes) != 0 {
			t.Errorf("want empty non-nil slice, got %#v", resp.Promocodes)
		}
	})

	t.Run("repo failure propagates", func(t *testing.T) {
		boom := errors.New("db down")
		repo := &mockPromoRepo{
			findAll: func(context.Context, *uuid.UUID) ([]model.Promocode, error) { return nil, boom },
		}
		if _, err := NewPromocodeService(repo, clock).List(context.Background(), nil); !errors.Is(err, boom) {
			t.Errorf("want repo error, got %v", err)
		}
	})
}

func TestPromocodeValidate(t *testing.T) {
	clock := fakeClock{t: testNow}

	repoWith := func(p *model.Promocode) *mockPromoRepo {
		return &mockPromoRepo{
			findByCode: func(_ context.Context, code string) (*model.Promocode, error) {
				if p != nil && code == p.Code {
					return p, nil
				}
				return nil, repository.ErrNotFound
			},
		}
	}

	t.Run("not found", func(t *testing.T) {
		_, err := NewPromocodeService(repoWith(nil), clock).Validate(context.Background(), "NOPE")
		if !errors.Is(err, ErrPromoNotFound) {
			t.Errorf("want ErrPromoNotFound, got %v", err)
		}
	})

	t.Run("repo failure propagates", func(t *testing.T) {
		boom := errors.New("db down")
		repo := &mockPromoRepo{
			findByCode: func(context.Context, string) (*model.Promocode, error) { return nil, boom },
		}
		if _, err := NewPromocodeService(repo, clock).Validate(context.Background(), "X"); !errors.Is(err, boom) {
			t.Errorf("want repo error, got %v", err)
		}
	})

	t.Run("inactive", func(t *testing.T) {
		p := &model.Promocode{Code: "OLD", Active: false, MaxUses: 10}
		_, err := NewPromocodeService(repoWith(p), clock).Validate(context.Background(), "OLD")
		if !errors.Is(err, ErrPromoInactive) {
			t.Errorf("want ErrPromoInactive, got %v", err)
		}
	})

	t.Run("exhausted at exactly max uses", func(t *testing.T) {
		p := &model.Promocode{Code: "FULL", Active: true, MaxUses: 3, TimesUsed: 3}
		_, err := NewPromocodeService(repoWith(p), clock).Validate(context.Background(), "FULL")
		if !errors.Is(err, ErrPromoExhausted) {
			t.Errorf("want ErrPromoExhausted, got %v", err)
		}
	})

	t.Run("valid with one use left", func(t *testing.T) {
		p := &model.Promocode{Code: "LAST", Active: true, MaxUses: 3, TimesUsed: 2, DiscountPercent: 20}
		got, err := NewPromocodeService(repoWith(p), clock).Validate(context.Background(), "LAST")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if got.DiscountPercent != 20 {
			t.Errorf("returned promo mangled: %+v", got)
		}
	})

	t.Run("lookup is case-insensitive and trimmed", func(t *testing.T) {
		p := &model.Promocode{Code: "SUMMER", Active: true, MaxUses: 10}
		var lookedUp string
		repo := &mockPromoRepo{
			findByCode: func(_ context.Context, code string) (*model.Promocode, error) {
				lookedUp = code
				return p, nil
			},
		}
		if _, err := NewPromocodeService(repo, clock).Validate(context.Background(), "  summer "); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if lookedUp != "SUMMER" {
			t.Errorf("repo queried with %q, want normalised SUMMER", lookedUp)
		}
	})
}
