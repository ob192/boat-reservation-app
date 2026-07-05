package service

import (
	"context"
	"errors"
	"strings"

	"github.com/google/uuid"

	"github.com/harbour-wave/harbour-wave-backend/internal/model"
	"github.com/harbour-wave/harbour-wave-backend/internal/platform"
	"github.com/harbour-wave/harbour-wave-backend/internal/repository"
)

// Promocode-specific typed errors. Handlers map these to HTTP codes.
var (
	ErrPromoNotFound      = errors.New("PROMO_NOT_FOUND")
	ErrPromoInactive      = errors.New("PROMO_INACTIVE")
	ErrPromoExhausted     = errors.New("PROMO_EXHAUSTED")
	ErrPromoAlreadyExists = errors.New("PROMO_ALREADY_EXISTS")
)

type PromocodeService interface {
	Create(ctx context.Context, code string, discountPercent, maxUses int, adminID uuid.UUID) (*model.AdminPromocodeResponse, error)
	List(ctx context.Context, createdBy *uuid.UUID) (*model.AdminPromocodeListResponse, error)

	// Validate returns the code if it exists, is active and has remaining uses.
	Validate(ctx context.Context, code string) (*model.Promocode, error)
}

type promocodeService struct {
	repo  repository.PromocodeRepository
	clock platform.Clock
}

func NewPromocodeService(repo repository.PromocodeRepository, clock platform.Clock) PromocodeService {
	return &promocodeService{repo: repo, clock: clock}
}

// NormalizeCode makes promocode matching case-insensitive.
func NormalizeCode(code string) string {
	return strings.ToUpper(strings.TrimSpace(code))
}

func (s *promocodeService) Create(
	ctx context.Context,
	code string,
	discountPercent, maxUses int,
	adminID uuid.UUID,
) (*model.AdminPromocodeResponse, error) {
	code = NormalizeCode(code)
	if code == "" || discountPercent < 0 || discountPercent > 100 || maxUses < 1 {
		return nil, ErrInvalidInput
	}

	if _, err := s.repo.FindByCode(ctx, code); err == nil {
		return nil, ErrPromoAlreadyExists
	} else if !errors.Is(err, repository.ErrNotFound) {
		return nil, err
	}

	p := &model.Promocode{
		Code:            code,
		DiscountPercent: discountPercent,
		MaxUses:         maxUses,
		TimesUsed:       0,
		Active:          true,
		CreatedBy:       adminID,
		CreatedAt:       s.clock.Now(),
	}
	if err := s.repo.Create(ctx, p); err != nil {
		return nil, err
	}
	resp := toPromoResponse(p)
	return &resp, nil
}

func (s *promocodeService) List(ctx context.Context, createdBy *uuid.UUID) (*model.AdminPromocodeListResponse, error) {
	rows, err := s.repo.FindAll(ctx, createdBy)
	if err != nil {
		return nil, err
	}
	out := make([]model.AdminPromocodeResponse, 0, len(rows))
	for i := range rows {
		out = append(out, toPromoResponse(&rows[i]))
	}
	return &model.AdminPromocodeListResponse{Promocodes: out}, nil
}

func (s *promocodeService) Validate(ctx context.Context, code string) (*model.Promocode, error) {
	code = NormalizeCode(code)
	p, err := s.repo.FindByCode(ctx, code)
	if errors.Is(err, repository.ErrNotFound) {
		return nil, ErrPromoNotFound
	}
	if err != nil {
		return nil, err
	}
	if !p.Active {
		return nil, ErrPromoInactive
	}
	if p.TimesUsed >= p.MaxUses {
		return nil, ErrPromoExhausted
	}
	return p, nil
}

func toPromoResponse(p *model.Promocode) model.AdminPromocodeResponse {
	return model.AdminPromocodeResponse{
		Code:            p.Code,
		DiscountPercent: p.DiscountPercent,
		MaxUses:         p.MaxUses,
		TimesUsed:       p.TimesUsed,
		Active:          p.Active,
		CreatedBy:       p.CreatedBy.String(),
		CreatedAt:       p.CreatedAt,
	}
}
