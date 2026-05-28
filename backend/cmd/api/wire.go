//go:build wireinject
// +build wireinject

package main

import (
	"github.com/google/wire"

	"github.com/harbour-wave/harbour-wave-backend/internal/config"
	"github.com/harbour-wave/harbour-wave-backend/internal/handler"
	"github.com/harbour-wave/harbour-wave-backend/internal/platform"
	"github.com/harbour-wave/harbour-wave-backend/internal/repository"
	"github.com/harbour-wave/harbour-wave-backend/internal/service"
)

// InitializeServer generates the dependency graph for our application.
func InitializeServer() (*Server, error) {
	wire.Build(
		// 1. Config & Platform
		config.Load,
		platform.NewClock,
		ProvideLogger,
		ProvideDB,

		// 2. Repositories
		repository.NewBookingRepository,
		repository.NewSlotRepository,
		repository.NewDateBlockRepository,
		repository.NewSystemRepository,

		// 3. Providers & Gateway
		ProvidePaymentGateway,
		ProvidePosterClient,

		// 4. Services
		service.NewAuthService,
		service.NewPricingService,
		service.NewAvailabilityService,
		service.NewBookingService,
		service.NewCheckoutService,
		service.NewEmailService,
		service.NewWebhookService,
		service.NewAdminService,

		// 5. Handlers
		handler.NewAvailabilityHandler,
		handler.NewBookingHandler,
		handler.NewCheckoutHandler,
		handler.NewWebhookHandler,
		handler.NewAdminHandler,

		// 6. Router & Container
		ProvideRouter,
		wire.Struct(new(Server), "*"),
	)
	return &Server{}, nil
}
