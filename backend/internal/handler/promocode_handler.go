package handler

import (
	"errors"
	"log/slog"
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/harbour-wave/harbour-wave-backend/internal/model"
	"github.com/harbour-wave/harbour-wave-backend/internal/service"
	"github.com/harbour-wave/harbour-wave-backend/pkg/httpx"
)

type PromocodeHandler struct {
	promocodeSvc service.PromocodeService
	log          *slog.Logger
}

func NewPromocodeHandler(svc service.PromocodeService, log *slog.Logger) *PromocodeHandler {
	return &PromocodeHandler{promocodeSvc: svc, log: log}
}

// GetPromoCode GET /promocodes/:code
//
// @Summary      Preview a promo code
// @Description  Authenticated user endpoint to check a promo code before booking. Returns only the discount percentage — never usage counts or ownership, which are admin-only.
// @Tags         promocodes
// @Produce      json
// @Param        code path string true "Promo code"
// @Success      200  {object}  model.PromoPreviewResponse
// @Failure      422  {object}  httpx.ErrorBody "PROMO_NOT_FOUND, PROMO_INACTIVE, PROMO_EXHAUSTED"
// @Failure      503  {object}  httpx.ErrorBody
// @Security     BearerAuth
// @Router       /promocodes/{code} [get]
func (h *PromocodeHandler) GetPromoCode(c *gin.Context) {
	pc, err := h.promocodeSvc.Validate(c.Request.Context(), c.Param("code"))
	if err != nil {
		switch {
		case errors.Is(err, service.ErrPromoNotFound):
			httpx.Err(c, http.StatusUnprocessableEntity, httpx.CodePromoNotFound, "")
		case errors.Is(err, service.ErrPromoInactive):
			httpx.Err(c, http.StatusUnprocessableEntity, httpx.CodePromoInactive, "")
		case errors.Is(err, service.ErrPromoExhausted):
			httpx.Err(c, http.StatusUnprocessableEntity, httpx.CodePromoExhausted, "")
		default:
			h.log.Error("get promocode", "err", err)
			httpx.Err(c, http.StatusServiceUnavailable, httpx.CodeServiceUnavailable, "")
		}
		return
	}
	httpx.OK(c, model.PromoPreviewResponse{DiscountPercent: pc.DiscountPercent})
}
